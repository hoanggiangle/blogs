package server

import (
	"encoding/json"
	"flag"
	"fmt"
	consulAPI "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"gitlab.sendo.vn/core/golang-sdk/new/registry"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"strings"
)

const (
	defaultPort              = 10000
	defaultSampleRatingTrace = 1.0
)

type gRPCServer struct {
	options       []grpc.ServerOption
	handlers      []func(*grpc.Server)
	logger        logger.Logger
	svr           *grpc.Server
	stopChan      chan bool
	name          string
	registeredID  string
	registryAgent registry.Agent
	Config
}

type Config struct {
	Port     int    `json:"gRPCPort"`
	BindAddr string `json:"gRPCAddress"`
	CertFile string `json:"gRPCCertFile"`
	KeyFile  string `json:"gRPCKeyFile"`

	// sample rating for remote tracing from OpenSensus
	SampleTraceRating float64

	// Jeager agent URI to receive tracing data directly (through UDP)
	JaegerAgentURI string

	// enable std tracing
	StdTracingEnabled bool
}

func New(name string) *gRPCServer {
	server := &gRPCServer{
		name:     name,
		logger:   logger.GetCurrent().GetLogger("gRPC"),
		stopChan: make(chan bool),
		options:  []grpc.ServerOption{},
		handlers: []func(*grpc.Server){},
	}

	return server
}

func (s *gRPCServer) InitFlags() {
	flag.IntVar(&s.Config.Port, "port", defaultPort, "gRPC server port. If 0 => get a random port")
	flag.StringVar(&s.BindAddr, "addr", "", "gRPC server bind address")
	flag.StringVar(&s.CertFile, "cert-file", "", "tls certificate file")
	flag.StringVar(&s.KeyFile, "key-file", "", "tls key file")

	flag.Float64Var(
		&s.SampleTraceRating,
		"trace-sample-rate",
		defaultSampleRatingTrace,
		"sample rating for remote tracing from OpenSensus: 0.0 -> 1.0 (default is 1.0)",
	)

	flag.StringVar(
		&s.JaegerAgentURI,
		"jaeger-agent-uri",
		"",
		"jaeger agent URI to receive tracing data directly",
	)

	flag.BoolVar(
		&s.StdTracingEnabled,
		"trace-std-enabled",
		false,
		"enable tracing export to std (default is false)",
	)
}

func (s *gRPCServer) Configure() error {
	var opts []grpc.ServerOption

	s.connectToJaegerAgent()

	s.options = append(s.options, grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	opts = append(opts, s.options...)

	if s.CertFile != "" && s.KeyFile != "" {
		cred, err := credentials.NewServerTLSFromFile(s.CertFile, s.KeyFile)
		if err != nil {
			s.logger.Errorf("Failed to load cert %v", err)
			return err
		}
		opts = []grpc.ServerOption{grpc.Creds(cred)}
	}

	s.svr = grpc.NewServer(opts...)
	for _, hdl := range s.handlers {
		hdl(s.svr)
	}

	return nil
}

func formatBindAddr(s string, p int) string {
	if strings.Contains(s, ":") && !strings.Contains(s, "[") {
		s = "[" + s + "]"
	}
	return fmt.Sprintf("%s:%d", s, p)
}

func (s *gRPCServer) Run() error {
	if err := s.Configure(); err != nil {
		return err
	}

	if strings.TrimSpace(s.name) == "" {
		s.logger.Error("service must have a name")
		return errors.New("service must have a name")
	}

	addr := formatBindAddr(s.BindAddr, s.Config.Port)
	s.logger.Debugf("start listen tcp %s...", addr)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Errorf("failed to listen: %v", err)
		return err
	}

	s.Config.Port = getPort(lis)
	s.logger.Infof("listen on %s...", lis.Addr())
	return s.svr.Serve(lis)
}

func getPort(lis net.Listener) int {
	addr := lis.Addr()
	tcp, _ := net.ResolveTCPAddr(addr.Network(), addr.String())
	return tcp.Port
}

func (s *gRPCServer) Name() string {
	return s.name
}

func (s *gRPCServer) URI() string {
	return formatBindAddr(s.BindAddr, s.Config.Port)
}

func (s *gRPCServer) Port() int {
	return s.Config.Port
}

func (s *gRPCServer) Stop() <-chan bool {
	go func() {
		if s.svr == nil {
			s.stopChan <- true
			return
		}

		s.svr.Stop()
		s.stopChan <- true
	}()

	return s.stopChan
}

func (s *gRPCServer) Get() *grpc.Server {
	return s.svr
}

func (s *gRPCServer) AddOption(opt grpc.ServerOption) {
	s.options = append(s.options, opt)
}

func (s *gRPCServer) AddHandler(hdl func(*grpc.Server)) {
	s.handlers = append(s.handlers, hdl)
}

func (s *gRPCServer) Register(agent registry.Agent) {
	asr := &consulAPI.AgentServiceRegistration{
		Name: s.Name(),
		Port: s.Config.Port,

		Check: &consulAPI.AgentServiceCheck{
			Notes:                          "default check tcp",
			TCP:                            fmt.Sprintf("%s:%d", agent.GetCheckIP(), s.Config.Port),
			Interval:                       "15s",
			DeregisterCriticalServiceAfter: "15m",
		},
	}
	s.registeredID = agent.RegisterService(asr)
}

func (s *gRPCServer) CheckKV(agent registry.Agent) {
	kvValueBytes := agent.GetKVs()
	var cf Config

	if err := json.Unmarshal(kvValueBytes, &cf); err != nil {
		return
	}

	if cf.Port == 0 || cf == s.Config {
		return
	}

	// Assign new config and reload service if needed
	s.Config = cf

	if s.IsRunning() {
		<-s.Stop()
		go func() { s.Run() }()
	}
}

func (s *gRPCServer) Reload(config Config) {
	s.Config = config
	<-s.Stop()
	go func() { s.Run() }()
}

func (s *gRPCServer) GetConfig() Config {
	return s.Config
}

func (s *gRPCServer) Deregister(agent registry.Agent) {
	agent.DeregisterService(s.registeredID)
}

func (s *gRPCServer) IsRunning() bool {
	return s.svr != nil
}

// Connect to Jaeger Agent to send tracing data
func (s *gRPCServer) connectToJaegerAgent() {
	if s.JaegerAgentURI == "" {
		return
	}

	s.logger.Infof("connecting to Jaeger Agent on %s...", s.JaegerAgentURI)

	je, _ := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: s.JaegerAgentURI,
		Process:       jaeger.Process{ServiceName: s.name},
	})

	// And now finally register it as a Trace Exporter
	trace.RegisterExporter(je)

	// Trace view for console
	if s.StdTracingEnabled {
		// Register stats and trace exporters to export
		// the collected data.
		view.RegisterExporter(&PrintExporter{})

		// Register the views to collect server request count.
		if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
			s.logger.Errorf("jaeger error: %s", err.Error())
		}
	}

	if s.SampleTraceRating >= 1 {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	} else {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(s.SampleTraceRating)})
	}
}
