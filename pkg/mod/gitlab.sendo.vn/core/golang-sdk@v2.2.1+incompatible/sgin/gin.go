package sgin

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	consulapi "github.com/hashicorp/consul/api"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/ssd"
)

type Config struct {
	App         sdms.Application
	SD          ssd.ConsulSD
	ServiceName string

	// used to register middlewares & handlers
	RegFunc func(r *gin.Engine)
	// create empty gin.Engine (no middleware)
	GinNoDefault bool

	// used for register to SD
	// disable reg route/config
	NoRegRoute bool
	// Prefix/Regex autoset base on gin Engine.Routes()
	// Manual set if need optimize (exclude some routes, optimize regex)
	// Name will alway be setted to ServiceName
	SvcConfig ssd.ServiceConfig

	// TODO: Custom HTTP health-check URL
	HealthCheckUrl string

	// 3000 if not set
	DefaultPort int

	// graceful shutdown timeout, default 3s
	StopTimeout time.Duration

	// prefix to flag
	FlagPrefix string
}

type GinService interface {
	sdms.RunnableService
	// block until ready
	Port() int

	isGinService()
}

type ginServiceImpl struct {
	app sdms.Application
	log sdms.Logger
	sd  ssd.ConsulSD

	cfg Config

	svr    *myHttpServer
	router *gin.Engine

	// flags
	port     int
	bindAddr string
	certFile string
	keyFile  string
	muReady  *sync.Mutex
}

var (
	haveInitGlobalFlag bool
	ginMode            string
	ginNoLogger        bool
)

func New(cfg *Config) (GinService, error) {
	if cfg.ServiceName == "" {
		return nil, errors.New("ServiceName is required")
	}

	if cfg.DefaultPort == 0 {
		cfg.DefaultPort = 3000
	}

	if cfg.StopTimeout == 0 {
		cfg.StopTimeout = time.Second * 3
	}

	mu := &sync.Mutex{}
	mu.Lock()

	return &ginServiceImpl{
		app:     cfg.App,
		log:     cfg.App.(sdms.SdkApplication).GetLog("gin"),
		sd:      cfg.SD,
		cfg:     *cfg,
		muReady: mu,
	}, nil
}

func (s *ginServiceImpl) InitFlags() {
	prefix := s.cfg.FlagPrefix

	flag.IntVar(&s.port, prefix+"port", s.cfg.DefaultPort, "gin server port. If 0 => get a random port")
	flag.StringVar(&s.bindAddr, prefix+"addr", "", "gin server bind address")
	flag.StringVar(&s.certFile, prefix+"cert-file", "", "tls certificate file")
	flag.StringVar(&s.keyFile, prefix+"key-file", "", "tls key file")

	if !haveInitGlobalFlag {
		flag.StringVar(&ginMode, "gin-mode", "", "gin mode")
		flag.BoolVar(&ginNoLogger, "gin-no-logger", false, "disable default gin logger middleware")
		haveInitGlobalFlag = true
	}
}

func (s *ginServiceImpl) Configure() error {
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	s.log.Debug("init gin engine...")
	s.router = gin.New()
	if !s.cfg.GinNoDefault {
		if !ginNoLogger {
			s.router.Use(gin.Logger())
		}
		s.router.Use(gin.Recovery())
	}

	if s.cfg.RegFunc != nil {
		s.cfg.RegFunc(s.router)
	}

	s.svr = &myHttpServer{
		Server: http.Server{Handler: s.router},
	}

	return nil
}

func formatBindAddr(s string, p int) string {
	if strings.Contains(s, ":") && !strings.Contains(s, "[") {
		s = "[" + s + "]"
	}
	return fmt.Sprintf("%s:%d", s, p)
}

func (s *ginServiceImpl) Run() error {
	addr := formatBindAddr(s.bindAddr, s.port)
	s.log.Debugf("start listen tcp %s...", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.log.Fatalf("failed to listen: %v", err)
	}

	s.port = getPort(lis)
	s.muReady.Unlock()

	s.updateServiceConfig()
	s.registerSD()

	defer s.sd.RemoveAll()

	s.log.Infof("listen on %s...", lis.Addr())
	if s.certFile == "" && s.keyFile == "" {
		err = s.svr.Serve(lis)
	} else {
		err = s.svr.ServeTLS(lis, s.certFile, s.keyFile)
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

func (s *ginServiceImpl) Port() int {
	s.muReady.Lock()
	defer s.muReady.Unlock()
	return s.port
}

func (s *ginServiceImpl) Cleanup() {
	s.sd.RemoveAll()
}

func (s *ginServiceImpl) Stop() {
	s.sd.RemoveAll()
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.StopTimeout)
	defer cancel()

	if s.svr != nil {
		s.svr.Shutdown(ctx)
	}
}

func (s *ginServiceImpl) registerSD() {
	asr := &consulapi.AgentServiceRegistration{
		Name: s.cfg.ServiceName,
		Port: s.port,
		Check: &consulapi.AgentServiceCheck{
			Notes:                          "default check tcp",
			TCP:                            fmt.Sprintf("%s:%d", s.sd.GetCheckIP(), s.port),
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30m",
		},
	}
	s.sd.Add(asr)
}

func (s *ginServiceImpl) updateServiceConfig() {
	if s.cfg.NoRegRoute {
		return
	}

	s.fixSvcConfig()
	s.sd.UpdateServiceConfig(&s.cfg.SvcConfig)
}

func (s *ginServiceImpl) fixSvcConfig() {
	sr := &s.cfg.SvcConfig
	sr.Name = s.cfg.ServiceName

	if sr.Protocol == "" && s.keyFile != "" && s.certFile != "" {
		sr.Protocol = "http2"
	}

	if sr.Prefix != "" || sr.Regex != "" {
		return
	}

	re1 := regexp.MustCompile(`:\w+`)
	re2 := regexp.MustCompile(`/\\\*\w+`)

	routes := []string{}
	for _, r := range s.router.Routes() {
		route := regexp.QuoteMeta(r.Path)
		route = re1.ReplaceAllString(route, "[^/]*")
		route = re2.ReplaceAllString(route, ".*")
		routes = append(routes, route)
	}
	if len(routes) == 0 {
		s.log.Fatal("No gin handler has been registered")
	} else if len(routes) == 1 {
		sr.Prefix = routes[0]
	} else {
		// sort route & dedup
		sort.Strings(routes)
		regex := ""
		for idx, r := range routes {
			if idx > 0 && r == routes[idx-1] {
				continue
			}
			regex = regex + "|" + r
		}
		regex = "(" + regex[1:] + ")"
		sr.Regex = regex
	}
}

func (g *ginServiceImpl) isGinService() {}
