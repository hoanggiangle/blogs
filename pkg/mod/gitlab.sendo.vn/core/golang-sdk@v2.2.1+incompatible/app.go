// Framework for build Sendo microservice in Golang
//
package sdms

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"gitlab.sendo.vn/core/golang-sdk/slog"
)

// For easy use
type Logger slog.Logger

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type AppConfig struct {
	// os.Args[0] if empty
	Name string
	// os.Args[1:] if empty
	Args []string
	// log config for sdk
	LogConfig *slog.LoggerConfig

	// don't use flag.CommandLine
	UseNewFlagSet bool
}

type Application interface {
	RegService(svc Service) error
	// mark a background servie as critical
	// so if Run() method is exit -> app will be stop
	SetCriticalService(svc RunnableService)
	RegMainService(svc RunnableService) error
	OutputEnv()
	Run()
	// reload all services config, log, ... Auto trigger by HUP signal
	Reload()
	Shutdown() <-chan struct{}
	IsShutdown() bool
	GetMainService() RunnableService
	GetServices() []Service
	// Register function to run before exit (include normal exit & fatal)
	RegisterExitHandler(func())
}

// Export some method that only used by itself
//
// External package should avoid to use
type SdkApplication interface {
	Application
	GetLog(prefix string) Logger
}

type serviceInfo struct {
	service  Service
	critical bool
}

type genericApplication struct {
	logSvc slog.LoggerService
	log    Logger

	services []serviceInfo

	chSignal chan os.Signal

	conf AppConfig

	cmdLine *AppFlagSet

	mainSVC RunnableService

	isShutdown bool

	mu *sync.Mutex

	exitHandlers []func()

	shutWaitChans []chan<- struct{}
}

// Create a new genericApplication
func NewApp(config *AppConfig) Application {
	app := &genericApplication{
		chSignal: make(chan os.Signal, 1),
		mu:       &sync.Mutex{},
	}

	if config.Name == "" {
		config.Name = os.Args[0]
	}

	if config.Args == nil {
		if len(os.Args) > 1 {
			config.Args = os.Args[1:]
		} else {
			config.Args = []string{}
		}
	}

	if config.UseNewFlagSet {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}
	app.cmdLine = newFlagSet(config.Name, flag.CommandLine)

	app.conf = *config

	app.initLogging()

	return app
}

func (app *genericApplication) initLogging() {
	if app.conf.LogConfig == nil {
		app.conf.LogConfig = &slog.LoggerConfig{}
	}
	if app.conf.LogConfig.FlagPrefix == "" {
		app.conf.LogConfig.FlagPrefix = "core-"
	}
	if app.conf.LogConfig.BasePrefix == "" {
		app.conf.LogConfig.BasePrefix = "core"
	}
	if app.conf.LogConfig.DefaultLevel == "" {
		app.conf.LogConfig.DefaultLevel = "info"
	}

	app.logSvc = slog.NewAppLogService(app.conf.LogConfig)
	app.RegService(app.logSvc)
	app.log = app.logSvc.GetLogger("app")

	slog.RegisterExitHandler(func() {
		<-app.Shutdown()
	})
}

func (app *genericApplication) GetLog(prefix string) Logger {
	return app.logSvc.GetLogger(prefix)
}

// Register a service
//
// Example: mongo, redis, ...
//
// It will be clean up when app shutdown
//
func (app *genericApplication) RegService(svc Service) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	for _, s := range app.services {
		if s.service == svc {
			app.log.Fatal("Can't register service twice!")
		}
	}

	svc.InitFlags()

	app.services = append(app.services, serviceInfo{service: svc})
	return nil
}

// Register a service as a main service
//
// Application will exit when this service exit
func (app *genericApplication) RegMainService(svc RunnableService) error {
	if err := app.RegService(svc); err != nil {
		return err
	}
	app.mainSVC = svc

	return nil
}

func (app *genericApplication) OutputEnv() {
	app.cmdLine.GetSampleEnvs()
}

// Run the app
func (app *genericApplication) Run() {
	isErr := false
	log := app.log

	if app.mainSVC == nil {
		log.Fatal("Please register a main service with RegMainService!")
	}

	app.parseFlags()

	signal.Notify(app.chSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	err := app.configure()
	if err != nil {
		app.log.Error(err)
		isErr = true
		goto SHUTDOWN
	}
	app.beginRun()

	for sig := range app.chSignal {
		switch sig {
		case syscall.SIGHUP:
			app.Reload()
			break
		default:
			goto SHUTDOWN
		}
	}

SHUTDOWN:
	log.Info("shutdown...")

	// run exit handlers
	app.mu.Lock()
	for _, handler := range app.exitHandlers {
		handler()
	}
	app.mu.Unlock()

	app.stop()
	app.cleanup()

	app.mu.Lock()
	app.isShutdown = true
	for _, ch := range app.shutWaitChans {
		ch <- struct{}{}
	}
	app.mu.Unlock()

	if isErr {
		os.Exit(1)
	}
}

// Shutdown app
func (app *genericApplication) Shutdown() <-chan struct{} {
	app.chSignal <- syscall.SIGTERM

	ch := make(chan struct{}, 1)

	app.mu.Lock()
	defer app.mu.Unlock()
	if app.isShutdown {
		ch <- struct{}{}
	} else {
		app.shutWaitChans = append(app.shutWaitChans, ch)
	}

	return ch
}

func (app *genericApplication) parseFlags() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	_, err := os.Stat(envFile)
	if err == nil {
		err := godotenv.Load(envFile)
		if err != nil {
			app.log.Fatalf("Loading env(%s): %s", envFile, err.Error())
		}
	} else if envFile != ".env" {
		app.log.Fatalf("Loading env(%s): %s", envFile, err.Error())
	}

	app.cmdLine.Parse(app.conf.Args)
}

func (app *genericApplication) configure() error {
	app.mu.Lock()
	defer app.mu.Unlock()

	for _, s := range app.services {
		if err := s.service.Configure(); err != nil {
			return err
		}
	}
	return nil
}

func (app *genericApplication) Reload() {
	app.mu.Lock()
	defer app.mu.Unlock()

	for _, s := range app.services {
		if rS, ok := s.service.(ReloadableService); ok {
			rS.Reload()
		}
	}
}

func (app *genericApplication) beginRun() {
	app.mu.Lock()
	defer app.mu.Unlock()

	runService := func(s RunnableService) {
		if err := s.Run(); err != nil {
			app.Shutdown()
		}
	}
	log := app.log
	for _, s := range app.services {
		runS, ok := s.service.(RunnableService)
		if !ok {
			continue
		}

		if app.mainSVC == runS {
			go func(runS RunnableService) {
				log.Info("start...")
				runService(runS)
				app.Shutdown()
			}(runS)
		} else if s.critical {
			go func(runS RunnableService) {
				runService(runS)
				app.Shutdown()
			}(runS)
		} else {
			go runService(runS)
		}
	}

}

func (app *genericApplication) stop() {
	app.mu.Lock()
	defer app.mu.Unlock()

	for i := len(app.services) - 1; i >= 0; i-- {
		s := app.services[i].service
		if runS, ok := s.(RunnableService); ok {
			runS.Stop()
		}
	}
}

func (app *genericApplication) cleanup() {
	app.mu.Lock()
	defer app.mu.Unlock()

	log := app.log

	log.Info("cleanup...")
	for i := len(app.services) - 1; i >= 0; i-- {
		app.services[i].service.Cleanup()
	}
}

func (app *genericApplication) IsShutdown() bool {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app == nil || app.isShutdown
}

func (app *genericApplication) GetMainService() RunnableService {
	if app == nil || app.IsShutdown() {
		return nil
	}
	return app.mainSVC
}

func (app *genericApplication) GetServices() []Service {
	app.mu.Lock()
	defer app.mu.Unlock()

	if app == nil || app.IsShutdown() {
		return nil
	}
	services := make([]Service, 0, len(app.services))
	for _, s := range app.services {
		services = append(services, s.service)
	}
	return services
}

func (app *genericApplication) SetCriticalService(svc RunnableService) {
	app.mu.Lock()
	defer app.mu.Unlock()

	for i, s := range app.services {
		if s.service == svc {
			app.services[i].critical = true
		}
	}
}

func (app *genericApplication) RegisterExitHandler(handler func()) {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.exitHandlers = append(app.exitHandlers, handler)
}
