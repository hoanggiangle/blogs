package slog

import (
	"flag"
	"time"

	"github.com/sirupsen/logrus"
)

type reloadLoggerService interface {
	LoggerService
	Reload()
}

// Used to log message that need send to remote like elasticsearch
type messageLogServiceImpl struct {
	*appLogServiceImpl

	// logo error message
	log Logger

	// flags
	logPath string
}

func NewMessageLogService(config *LoggerConfig, internal Logger) LoggerService {
	if config == nil {
		config = &LoggerConfig{}
	}
	if config.FlagPrefix == "" {
		config.FlagPrefix = "file-"
	}

	appLog := NewAppLogService(config).(*appLogServiceImpl)
	appLog.logger.Formatter = &logrus.JSONFormatter{}

	if internal == nil {
		newLog := logrus.New()
		newLog.Formatter = logrus.Formatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "15:04:05",
		})
		internal = &logger{logrus.NewEntry(newLog)}
	}

	var s reloadLoggerService
	s = &messageLogServiceImpl{
		appLogServiceImpl: appLog,
		log:               internal,
	}
	return s
}

func (m *messageLogServiceImpl) InitFlags() {
	m.appLogServiceImpl.InitFlags()

	flag.StringVar(&m.logPath, m.cfg.FlagPrefix+"logfile", "", "File to write log to. Default write to console")
}

func (m *messageLogServiceImpl) Configure() error {
	err := m.appLogServiceImpl.Configure()
	if err != nil {
		return err
	}

	if m.logPath == "" {
		return nil
	}

	m.logger.Out, err = newReloadFile(m.logPath)
	if err != nil {
		m.log.Fatal("Fail to open log file: ", err.Error())
	}

	return nil
}

func (m *messageLogServiceImpl) Reload() {
	log := m.log

	file, ok := m.logger.Out.(*reloadFile)
	if ok {
		log.Info("Rotate log ", m.logPath)
		err := file.ReOpen()
		if err != nil {
			log.Fatal("Fail to open log file: ", err.Error())
		}
	}
}

func (m *messageLogServiceImpl) Cleanup() {
	oldF, ok := m.logger.Out.(*reloadFile)
	if ok {
		oldF.Sync()
		go func() {
			// close after 10s, so other process can write final log
			time.Sleep(time.Second * 10)
			oldF.Close()
		}()
	}
	m.appLogServiceImpl.Cleanup()
}
