package slog

import (
	"flag"
	"strings"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	haveInitGlobalFlag bool
	includeSource      bool
)

type ExcludePrefixFunc func(string) bool

type LoggerConfig struct {
	// prefix to flag
	FlagPrefix string

	DefaultLevel string
	// log prefix
	BasePrefix string

	ExcludePrefix ExcludePrefixFunc
}

// clone of sdk Service
type appService interface {
	// config flags here
	InitFlags()

	// configure service
	Configure() error

	// Cleanup service like db connection, remove temp files
	Cleanup()
}

type LoggerService interface {
	appService
	GetLogger(prefix string) Logger
}

// A default app logger
//
// Just write everything to console
type appLogServiceImpl struct {
	logger *logrus.Logger
	cfg    LoggerConfig

	// flags
	logLevel string
}

func NewAppLogService(config *LoggerConfig) LoggerService {
	if config == nil {
		config = &LoggerConfig{}
	}

	if config.DefaultLevel == "" {
		config.DefaultLevel = "info"
	}

	logger := logrus.New()
	logger.Formatter = logrus.Formatter(&prefixed.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})
	lv := mustParseLevel(config.DefaultLevel)
	logger.SetLevel(lv)

	return &appLogServiceImpl{
		logger:   logger,
		cfg:      *config,
		logLevel: config.DefaultLevel,
	}
}

func (s *appLogServiceImpl) InitFlags() {
	flag.StringVar(&s.logLevel, s.cfg.FlagPrefix+"loglevel", s.cfg.DefaultLevel, "Loglevel")

	if !haveInitGlobalFlag {
		flag.BoolVar(&includeSource, "log-source", false, "Include source line number in log message")
		haveInitGlobalFlag = true
	}
}

func (l *appLogServiceImpl) Configure() error {
	lv := mustParseLevel(l.logLevel)
	l.logger.SetLevel(lv)

	return nil
}

func (s *appLogServiceImpl) Cleanup() {
}

func (s *appLogServiceImpl) GetLogger(prefix string) Logger {
	var entry *logrus.Entry

	prefix = s.cfg.BasePrefix + "." + prefix
	prefix = strings.Trim(prefix, ".")
	if prefix == "" {
		entry = logrus.NewEntry(s.logger)
	} else {
		entry = s.logger.WithField("prefix", prefix)
	}

	l := &logger{entry}
	var log Logger = l

	if includeSource {
		log = &loggerWithSrc{l}
	}

	if s.cfg.ExcludePrefix != nil && s.cfg.ExcludePrefix(prefix) {
		log = &nullLogger{log}
	}

	return log
}
