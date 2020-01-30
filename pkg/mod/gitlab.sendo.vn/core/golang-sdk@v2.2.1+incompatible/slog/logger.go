// Logging provider
//
// Log only fully init when app.Run() called
package slog

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Fields logrus.Fields

type Logger interface {
	Debug(...interface{})
	Debugln(...interface{})
	Debugf(string, ...interface{})

	Info(...interface{})
	Infoln(...interface{})
	Infof(string, ...interface{})

	Warn(...interface{})
	Warnln(...interface{})
	Warnf(string, ...interface{})

	Error(...interface{})
	Errorln(...interface{})
	Errorf(string, ...interface{})

	Fatal(...interface{})
	Fatalln(...interface{})
	Fatalf(string, ...interface{})

	Panic(...interface{})
	Panicln(...interface{})
	Panicf(string, ...interface{})

	With(key string, value interface{}) Logger
	Withs(Fields) Logger
	// add source field to log
	WithSrc() Logger
}

type logger struct {
	*logrus.Entry
}

func (l *logger) debugSrc() *logrus.Entry {
	if _, ok := l.Entry.Data["source"]; ok {
		return l.Entry
	}

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		file = file[slash+1:]
	}
	return l.Entry.WithField("source", fmt.Sprintf("%s:%d", file, line))
}

func (l *logger) Debug(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.DebugLevel {
		l.debugSrc().Debug(args...)
	}
}

func (l *logger) Debugln(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.DebugLevel {
		l.debugSrc().Debugln(args...)
	}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.DebugLevel {
		l.debugSrc().Debugf(format, args...)
	}
}

func (l *logger) With(key string, value interface{}) Logger {
	return &logger{l.Entry.WithField(key, value)}
}

func (l *logger) Withs(fields Fields) Logger {
	return &logger{l.Entry.WithFields(logrus.Fields(fields))}
}

func (l *logger) WithSrc() Logger {
	return &logger{l.debugSrc()}
}

// always include withsource, used for debug only
type loggerWithSrc struct {
	*logger
}

func (l *loggerWithSrc) Info(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.InfoLevel {
		l.debugSrc().Info(args...)
	}
}

func (l *loggerWithSrc) Infoln(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.InfoLevel {
		l.debugSrc().Infoln(args...)
	}
}

func (l *loggerWithSrc) Infof(format string, args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.InfoLevel {
		l.debugSrc().Infof(format, args...)
	}
}

func (l *loggerWithSrc) Warn(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.WarnLevel {
		l.debugSrc().Warn(args...)
	}
}

func (l *loggerWithSrc) Warnln(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.WarnLevel {
		l.debugSrc().Warnln(args...)
	}
}

func (l *loggerWithSrc) Warnf(format string, args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.WarnLevel {
		l.debugSrc().Warnf(format, args...)
	}
}

func (l *loggerWithSrc) Error(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.ErrorLevel {
		l.debugSrc().Error(args...)
	}
}

func (l *loggerWithSrc) Errorln(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.ErrorLevel {
		l.debugSrc().Errorln(args...)
	}
}

func (l *loggerWithSrc) Errorf(format string, args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.ErrorLevel {
		l.debugSrc().Errorf(format, args...)
	}
}

func (l *loggerWithSrc) Fatal(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.FatalLevel {
		l.debugSrc().Fatal(args...)
	}
}

func (l *loggerWithSrc) Fatalln(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.FatalLevel {
		l.debugSrc().Fatalln(args...)
	}
}

func (l *loggerWithSrc) Fatalf(format string, args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.FatalLevel {
		l.debugSrc().Fatalf(format, args...)
	}
}

func (l *loggerWithSrc) Panic(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.PanicLevel {
		l.debugSrc().Panic(args...)
	}
}

func (l *loggerWithSrc) Panicln(args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.PanicLevel {
		l.debugSrc().Panicln(args...)
	}
}

func (l *loggerWithSrc) Panicf(format string, args ...interface{}) {
	if l.Entry.Logger.Level >= logrus.PanicLevel {
		l.debugSrc().Panicf(format, args...)
	}
}

func (l *loggerWithSrc) With(key string, value interface{}) Logger {
	return &loggerWithSrc{
		&logger{l.Entry.WithField(key, value)},
	}
}

func (l *loggerWithSrc) Withs(fields Fields) Logger {
	return &loggerWithSrc{
		&logger{l.Entry.WithFields(logrus.Fields(fields))},
	}
}

func mustParseLevel(level string) logrus.Level {
	lv, err := logrus.ParseLevel(level)
	if err != nil {
		log.Fatal(err.Error())
	}
	return lv
}

// just ignore all message level < Error
type nullLogger struct {
	Logger
}

func (l *nullLogger) Debug(args ...interface{})                 {}
func (l *nullLogger) Debugln(args ...interface{})               {}
func (l *nullLogger) Debugf(format string, args ...interface{}) {}
func (l *nullLogger) Info(args ...interface{})                  {}
func (l *nullLogger) Infoln(args ...interface{})                {}
func (l *nullLogger) Infof(format string, args ...interface{})  {}
func (l *nullLogger) Warn(args ...interface{})                  {}
func (l *nullLogger) Warnln(args ...interface{})                {}
func (l *nullLogger) Warnf(format string, args ...interface{})  {}
func (l *nullLogger) With(key string, value interface{}) Logger {
	return &nullLogger{l.Logger.With(key, value)}
}
func (l *nullLogger) Withs(fields Fields) Logger {
	return &nullLogger{l.Logger.Withs(Fields(fields))}
}
func (l *nullLogger) WithSrc() Logger {
	return &nullLogger{l.Logger.WithSrc()}
}

// return a anonymous logger
//
// mainly used for testing
// it's always include source line
func NewAnonLogger() Logger {
	log := logrus.New()
	log.Formatter = logrus.Formatter(&prefixed.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})

	if level := os.Getenv("CORE_LOGLEVEL"); level != "" {
		lv := mustParseLevel(level)
		log.SetLevel(lv)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	entry := logrus.NewEntry(log)
	return &loggerWithSrc{&logger{entry}}
}

// register handler for fatal
//
// should be used by sdms.Application only
func RegisterExitHandler(handler func()) {
	logrus.RegisterExitHandler(handler)
}
