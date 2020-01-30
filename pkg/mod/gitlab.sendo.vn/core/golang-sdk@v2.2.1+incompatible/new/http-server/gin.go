package http_server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	consulAPI "github.com/hashicorp/consul/api"
	"gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"gitlab.sendo.vn/core/golang-sdk/new/registry"
	"net"
	"net/http"
	"strings"
	"sync"
)

var (
	ginMode     string
	ginNoLogger bool
	defaultPort = 3000
)

type Config struct {
	Enabled      bool
	Port         int
	BindAddr     string
	CertFile     string
	KeyFile      string
	GinNoDefault bool

	// used for register to SD
	// disable reg route/config
	NoRegRoute bool
}

type GinService interface {
	sdms.RunnableService
	// block until ready
	Port() int
	isGinService()
}

type ginService struct {
	Config
	name          string
	logger        logger.Logger
	svr           *myHttpServer
	router        *gin.Engine
	mu            *sync.Mutex
	handlers      []func(*gin.Engine)
	registeredID  string
	registryAgent registry.Agent
}

func New(name string) *ginService {
	return &ginService{
		name:     name,
		logger:   logger.GetCurrent().GetLogger("gin"),
		mu:       &sync.Mutex{},
		handlers: []func(*gin.Engine){},
	}
}

func (gs *ginService) Name() string {
	return gs.name + "-gin"
}

func (gs *ginService) InitFlags() {
	prefix := "gin"
	flag.IntVar(&gs.Config.Port, prefix+"Port", defaultPort, "gin server Port. If 0 => get a random Port")
	flag.StringVar(&gs.BindAddr, prefix+"addr", "", "gin server bind address")
	flag.BoolVar(&gs.Enabled, prefix+"gin-enabled", true, "if gin server is enabled")
	flag.StringVar(&gs.CertFile, prefix+"cert-file", "", "tls certificate file")
	flag.StringVar(&gs.KeyFile, prefix+"key-file", "", "tls key file")
	flag.StringVar(&ginMode, "gin-mode", "", "gin mode")
	flag.BoolVar(&ginNoLogger, "gin-no-logger", false, "disable default gin logger middleware")
}

func (gs *ginService) Configure() error {
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	gs.logger.Debug("init gin engine...")
	gs.router = gin.New()
	if !gs.GinNoDefault {
		if !ginNoLogger {
			gs.router.Use(gin.Logger())
		}
		gs.router.Use(gin.Recovery())
	}

	gs.svr = &myHttpServer{
		Server: http.Server{Handler: gs.router},
	}

	return nil
}

func formatBindAddr(s string, p int) string {
	if strings.Contains(s, ":") && !strings.Contains(s, "[") {
		s = "[" + s + "]"
	}
	return fmt.Sprintf("%s:%d", s, p)
}

func (gs *ginService) Run() error {
	if gs.isDisabled() {
		return nil
	}

	if err := gs.Configure(); err != nil {
		return err
	}

	for _, hdl := range gs.handlers {
		hdl(gs.router)
	}

	addr := formatBindAddr(gs.BindAddr, gs.Config.Port)
	gs.logger.Debugf("start listen tcp %s...", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		gs.logger.Fatalf("failed to listen: %v", err)
	}

	gs.Config.Port = getPort(lis)

	gs.logger.Infof("listen on %s...", lis.Addr())
	if gs.CertFile == "" && gs.KeyFile == "" {
		err = gs.svr.Serve(lis)
	} else {
		err = gs.svr.ServeTLS(lis, gs.CertFile, gs.KeyFile)
	}

	if err != nil && err == http.ErrServerClosed {
		return nil
	}
	return err
}

func getPort(lis net.Listener) int {
	addr := lis.Addr()
	tcp, _ := net.ResolveTCPAddr(addr.Network(), addr.String())
	return tcp.Port
}

func (gs *ginService) Port() int {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.Config.Port
}

func (gs *ginService) Stop() <-chan bool {
	c := make(chan bool)

	go func() {
		if gs.svr != nil {
			gs.svr.Shutdown(context.Background())
		}
		c <- true
	}()
	return c
}

func (gs *ginService) Register(agent registry.Agent) {
	if gs.isDisabled() {
		return
	}

	asr := &consulAPI.AgentServiceRegistration{
		Name: gs.Name(),
		Port: gs.Config.Port,

		Check: &consulAPI.AgentServiceCheck{
			Notes:                          "default check tcp",
			TCP:                            fmt.Sprintf("%s:%d", agent.GetCheckIP(), gs.Config.Port),
			Interval:                       "15s",
			DeregisterCriticalServiceAfter: "15m",
		},
	}
	gs.registeredID = agent.RegisterService(asr)
}

func (gs *ginService) URI() string {
	return formatBindAddr(gs.BindAddr, gs.Config.Port)
}

func (gs *ginService) AddHandler(hdl func(*gin.Engine)) {
	gs.handlers = append(gs.handlers, hdl)
}

func (gs *ginService) CheckKV(agent registry.Agent) {
	if gs.isDisabled() {
		return
	}

	kvValueBytes := agent.GetKVs()
	var cf Config

	if err := json.Unmarshal(kvValueBytes, &cf); err != nil {
		return
	}

	if cf.Port == 0 || cf == gs.Config {
		return
	}

	// Assign new config and reload service if needed
	gs.Config = cf

	if gs.IsRunning() {
		<-gs.Stop()
		go func() { gs.Run() }()
	}
}

func (gs *ginService) Reload(config Config) {
	gs.Config = config
	<-gs.Stop()
	go func() { gs.Run() }()
}

func (gs *ginService) GetConfig() Config {
	return gs.Config
}

func (gs *ginService) Deregister(agent registry.Agent) {
	if gs.isDisabled() {
		return
	}
	agent.DeregisterService(gs.registeredID)
}

func (gs *ginService) IsRunning() bool {
	return gs.svr != nil
}

func (gs *ginService) isDisabled() bool {
	return !gs.Config.Enabled
}
