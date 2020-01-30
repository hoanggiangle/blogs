package sendo

import (
	"context"
	"flag"
	"github.com/joho/godotenv"
	"gitlab.sendo.vn/core/golang-sdk/new/broker"
	"gitlab.sendo.vn/core/golang-sdk/new/http-server"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"gitlab.sendo.vn/core/golang-sdk/new/registry"
	"gitlab.sendo.vn/core/golang-sdk/new/server"
	sdStorage "gitlab.sendo.vn/core/golang-sdk/new/storage"
	"google.golang.org/grpc"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type service struct {
	name         string
	version      string
	opts         []Option
	subServices  []Runnable
	initServices []Runnable
	isRegister   bool
	logger       logger.Logger
	registry     Registry
	broker       Broker
	storage      sdStorage.Storage
	server       Server
	httpServer   HttpServer
	client       Client
	signalChan   chan os.Signal
	cmdLine      *AppFlagSet
	stopFunc     func()
}

func New(opts ...Option) Service {
	sv := &service{
		opts: opts,
		//logger:       logger.GetCurrent().GetLogger("service"),
		signalChan:   make(chan os.Signal, 1),
		subServices:  []Runnable{},
		initServices: []Runnable{},
	}

	// init default logger
	logger.InitServLogger(false)
	sv.logger = logger.GetCurrent().GetLogger("service")

	// default Broker
	sv.broker = broker.NewNotSetBroker()

	for _, opt := range opts {
		opt(sv)
	}

	// Server
	gRPCSv := server.New(sv.name)
	sv.server = gRPCSv
	// Http server
	httpServer := http_server.New(sv.name)
	sv.httpServer = httpServer

	// Append gRPC server and http server to sub services
	sv.subServices = append(sv.subServices, gRPCSv, httpServer)

	// Client
	sv.client = defaultClient(gRPCSv)

	// Registry
	consul := registry.NewConsul(sv.name)
	sv.registry = consul
	// DB
	mongoDB := sdStorage.NewMongoDB("default-mongo", "")
	redisDB := sdStorage.NewRedisDB("default-redis", "")
	goRedis := sdStorage.NewGoRedisDB("default-go-redis", "")
	elasticSearch := sdStorage.NewES("default-es", "")
	gormDB := sdStorage.NewGormDB("default-gorm", "")
	sv.initServices = append(sv.initServices, consul, mongoDB, redisDB, goRedis, elasticSearch, gormDB)

	// Find all runnable with prefix non-empty
	var hpDBs []sdStorage.HasPrefix
	for _, srv := range sv.initServices {
		if v, ok := srv.(sdStorage.HasPrefix); ok && v.GetPrefix() != "" {
			hpDBs = append(hpDBs, v)
		}
	}

	sv.storage = sdStorage.New(sdStorage.Drivers{
		MongoDB:   mongoDB,
		RedisDB:   redisDB,
		GoRedisDB: goRedis,
		ES:        elasticSearch,
		GormDB:    gormDB,
	}, hpDBs)

	sv.initFlags()

	if sv.name == "" {
		if len(os.Args) >= 2 {
			sv.name = strings.Join(os.Args[:2], " ")
		}
	}

	sv.cmdLine = newFlagSet(sv.name, flag.CommandLine)
	sv.parseFlags()

	return sv
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Version() string {
	return s.version
}

func (s *service) Init() error {
	for _, dbSv := range s.initServices {
		if err := dbSv.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) IsRegistered() bool {
	return s.isRegister
}

func (s *service) Start() error {
	signal.Notify(s.signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	c := s.run()
	s.stopFunc = s.activeRegistry()

	for {
		select {
		case err := <-c:
			if err != nil {
				s.logger.Error(err.Error())
				s.Stop()
				return err
			}

		case <-s.registry.SyncChan():
			for _, sv := range s.subServices {
				if tsv, ok := sv.(Trackable); ok {
					tsv.CheckKV(s.registry)
				}
			}

		case sig := <-s.signalChan:
			s.logger.Infoln(sig)
			switch sig {
			case syscall.SIGHUP:
				return nil
			default:
				s.Stop()
				return nil
			}
		}
	}

	return nil
}

func (s *service) activeRegistry() func() {
	if !s.registry.IsRunning() {
		return func() {}
	}

	ctx, cancel := context.WithCancel(context.Background())
	registerTimer := time.NewTicker(time.Second * 60)

	go func() {
		for {
			select {
			case <-ctx.Done():
				registerTimer.Stop()
				return
			case <-registerTimer.C:
				for _, sv := range s.subServices {
					if tsv, ok := sv.(Trackable); ok {
						tsv.Register(s.registry)
					}
				}
			}
		}
	}()

	return cancel
}

func (s *service) initFlags() {
	for _, subService := range s.subServices {
		subService.InitFlags()
	}

	for _, dbService := range s.initServices {
		dbService.InitFlags()
	}
}

func (s *service) run() <-chan error {
	c := make(chan error, 1)

	// Start all services
	for _, subService := range s.subServices {
		go func(subSv Runnable) {
			// Since we use registry to store config
			// so the service need to check the config before run
			if s.registry.IsRunning() {
				if tsv, ok := subSv.(Trackable); ok {
					tsv.CheckKV(s.registry)
					tsv.Register(s.registry)
				}
			}
			c <- subSv.Run()
		}(subService)
	}

	return c
}

func (s *service) Stop() {
	s.logger.Infoln("Stopping service...")
	stopChan := make(chan bool)
	for _, subService := range s.subServices {
		go func(subSv Runnable) {
			if tsv, ok := subSv.(Trackable); ok && s.registry.IsRunning() {
				tsv.Deregister(s.registry)
			}

			stopChan <- <-subSv.Stop()
		}(subService)
	}

	for _, dbSv := range s.initServices {
		go func(subSv Runnable) {
			stopChan <- <-subSv.Stop()
		}(dbSv)
	}

	for i := 0; i < len(s.subServices)+len(s.initServices); i++ {
		<-stopChan
	}

	s.stopFunc()
	s.logger.Infoln("Service stopped")
}

func (s *service) RunFunction(fn Function) error {
	return fn(s)
}

func (s *service) RunClient(fn ClientFunction, opts ...grpc.DialOption) error {
	if opts == nil {
		opts = []grpc.DialOption{grpc.WithInsecure()}
	}

	cc, err := s.client.Get(opts...)
	if err != nil {
		return err
	}
	return fn(cc, s)
}

func (s *service) Server() Server {
	return s.server
}

func (s *service) HTTPServer() HttpServer {
	return s.httpServer
}

func (s *service) Client() Client {
	return s.client
}

func (s *service) Storage() sdStorage.Storage {
	return s.storage
}

func (s *service) Broker() Broker {
	return s.broker
}

func (s *service) Logger(prefix string) logger.Logger {
	return logger.GetCurrent().GetLogger(prefix)
}

func (s *service) OutEnv() {
	s.cmdLine.GetSampleEnvs()
}

func (s *service) parseFlags() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	_, err := os.Stat(envFile)
	if err == nil {
		err := godotenv.Load(envFile)
		if err != nil {
			s.logger.Fatalf("Loading env(%s): %s", envFile, err.Error())
		}
	} else if envFile != ".env" {
		s.logger.Fatalf("Loading env(%s): %s", envFile, err.Error())
	}

	s.cmdLine.Parse([]string{})
}

func WithName(name string) Option {
	return func(s *service) {
		s.name = name
	}
}

func WithVersion(version string) Option {
	return func(s *service) {
		s.version = version
	}
}

// Add Runnable component to SDK
// These components will run parallel in when service run
func WithRunnable(r Runnable) Option {
	//r.InitFlags()

	return func(s *service) {
		s.subServices = append(s.subServices, r)
	}
}

// Add init component to SDK
// These components will run sequentially before service run
func WithInitRunnable(r Runnable) Option {

	return func(s *service) {
		s.initServices = append(s.initServices, r)
	}
}

// Add broker component to SDK
// Broker is the way for async calling between services (pubsub system)
func WithBroker(pubEvents, subEvents []string) Option {
	brk := broker.New(&broker.Config{
		PublishEvents:   pubEvents,
		SubscribeEvents: subEvents,
	})

	return func(s *service) {
		s.broker = brk
		s.initServices = append(s.initServices, brk)
	}
}

// Service will write log data to file with this option
func WithFileLogger() Option {
	return func(s *service) {
		logger.InitServLogger(true)
		s.logger = logger.GetCurrent().GetLogger("service")
	}
}
