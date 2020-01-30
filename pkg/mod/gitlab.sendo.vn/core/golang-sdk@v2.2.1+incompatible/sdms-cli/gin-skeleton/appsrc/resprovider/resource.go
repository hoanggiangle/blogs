package resprovider

import (
    "log"

	"github.com/globalsign/mgo"
	"github.com/gomodule/redigo/redis"
	"github.com/streadway/amqp"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/samqp"
	"gitlab.sendo.vn/core/golang-sdk/slog"
	"gitlab.sendo.vn/core/golang-sdk/smgo"
	"gitlab.sendo.vn/core/golang-sdk/sredigo"
)

type Logger slog.Logger

// provide interface to access remote resources
type ResourceProvider interface {
	// Absolute dir to app root
	GetAppRoot() string
	// get console Logger
	Logger(prefix string) Logger
	{{- if .Resources.logfile }}
	// get file logger
	Flogger(prefix string) Logger
	{{- end }}
	{{- if .Resources.mgo }}

	// mongo
	Session(opts ...smgo.Option) *mgo.Session
	DB(opts ...smgo.Option) (db *mgo.Database, cleanup func())
	C(name string, opts ...smgo.Option) (c *mgo.Collection, cleanup func())
	{{- end }}
	{{- if .Resources.redis }}

	// get a redis connection from pool
	Redis() redis.Conn
	{{- end }}
	{{- if .Resources.amqp }}

	// Open a new RabbitMQ connection
	NewRabbitConn() (c *amqp.Connection, err error)
	{{- end }}
}

var rp ResourceProvider

func InitResourceProvider(app sdms.Application) {
	if rp != nil {
		app := rp.(*myResourceProvider).app
		if !app.IsShutdown() {
			log := rp.Logger("resprovider")
			log.Fatal("InitResourceProvider can not called twice!")
		}
	}
	rp = newRP(app)
}

// get ResourceProvider instace
func GetInstance() ResourceProvider {
	if rp == nil {
		log.Fatal("InitResourceProvider is not called yet!")
	}
	return rp
}

// private implement
type myResourceProvider struct {
	app sdms.Application

	rootDir string

	// console log
	clog slog.LoggerService

	{{- if .Resources.logfile }}
	// file logging
	flog slog.LoggerService
	{{- end }}

	{{- if .Resources.redis }}
	sredigo.RedigoService
	{{- end }}

	{{- if .Resources.mgo }}
	smgo.MgoService
	{{- end }}

	{{- if .Resources.amqp }}
	samqp.AmqpService
	{{- end }}
}

func newRP(app sdms.Application) *myResourceProvider {
	rp := &myResourceProvider{}
	rp.app = app

	// logger
	rp.clog = slog.NewAppLogService(nil)
	app.RegService(rp.clog)

	{{- if .Resources.logfile }}

	rp.flog = slog.NewMessageLogService(nil, rp.clog.GetLogger("msglog"))
	app.RegService(rp.flog)
	{{- end }}

	{{- if .Resources.redis }}

	// redis
	rp.RedigoService = sredigo.NewRedigo(&sredigo.RedigoConfig{
		App: app,
	})
	app.RegService(rp.RedigoService)
	{{- end }}

	{{- if .Resources.mgo }}

	// mongo
	rp.MgoService = smgo.NewMgoService(&smgo.MgoConfig{
		App:           app,
		DefaultDBName: "test",
	})
	app.RegService(rp.MgoService)
	{{- end }}

	{{- if .Resources.amqp }}

	// rabbitmq
	rp.AmqpService = samqp.NewAmqp(&samqp.AmqpConfig{
		App: app,
	})
	app.RegService(rp.AmqpService)
	{{- end }}

	return rp
}

func (m *myResourceProvider) GetAppRoot() string {
	if m.rootDir == "" {
		_, file, _, _ := runtime.Caller(0)
		m.rootDir = path.Dir(path.Dir(path.Dir(file)))
	}
	return m.rootDir
}

func (m *myResourceProvider) Logger(prefix string) Logger {
	return m.clog.GetLogger(prefix)
}

{{- if .Resources.logfile }}

func (m *myResourceProvider) Flogger(prefix string) Logger {
	return m.flog.GetLogger(prefix)
}
{{- end }}

{{- if .Resources.redis }}

func (m *myResourceProvider) Redis() redis.Conn {
	return m.RedigoService.Get()
}
{{- end }}

{{- if .Resources.amqp }}

func (m *myResourceProvider) NewRabbitConn() (c *amqp.Connection, err error) {
	return m.AmqpService.NewConn()
}
{{- end }}
