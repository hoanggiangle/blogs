package sgrpc

import (
	"flag"
	"fmt"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"log"
	"net"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/ssd"
)

const (
	defaultSampleRatingTrace = 1.0
	defaultEnableStdTracing  = false
)

type GrpcConfig struct {
	App          sdms.Application
	SD           ssd.ConsulSD
	RegisterFunc func(*grpc.Server)
	ServiceName  string

	// used for register to SD
	// disable reg route/config
	NoRegRoute bool
	// Prefix/Regex autoset base on grpc server GetServiceInfo()
	// Manual set if need optimize (exclude some routes, optimize regex)
	// Name will alway be setted to ServiceName
	SvcConfig ssd.ServiceConfig

	// 10000 if not set
	DefaultPort int

	// graceful shutdown timeout, default 3s
	StopTimeout time.Duration

	// prefix to flag
	FlagPrefix string

	ServerOptions []grpc.ServerOption
}

type GrpcService interface {
	sdms.RunnableService
	Port() int

	isGrpcService()
}

type grpcServiceImpl struct {
	name string
	cfg  GrpcConfig

	log sdms.Logger
	sd  ssd.ConsulSD

	svr *grpc.Server

	muReady *sync.Mutex

	// flags
	port     int
	bindAddr string
	certFile string
	keyFile  string

	// sample rating for remote tracing from OpenSensus
	sampleTraceRating float64

	// Jeager agent URI to receive tracing data directly (through UDP)
	jaegerAgentURI string

	// enable std tracing
	stdTracingEnabled bool
}

func New(cfg *GrpcConfig) GrpcService {
	if cfg.ServiceName == "" {
		log.Fatal("ServiceName is required")
	}
	if cfg.App == nil {
		log.Fatal("Invalid app")
	}

	if cfg.DefaultPort == 0 {
		cfg.DefaultPort = 10000
	}

	if cfg.StopTimeout == 0 {
		cfg.StopTimeout = time.Second * 3
	}

	mu := &sync.Mutex{}
	mu.Lock()

	return &grpcServiceImpl{
		cfg:     *cfg,
		name:    cfg.ServiceName,
		log:     cfg.App.(sdms.SdkApplication).GetLog("grpc"),
		sd:      cfg.SD,
		muReady: mu,
	}
}

func (s *grpcServiceImpl) InitFlags() {
	prefix := s.cfg.FlagPrefix
	flag.IntVar(&s.port, prefix+"port", s.cfg.DefaultPort, "gRPC server port. If 0 => get a random port")
	flag.StringVar(&s.bindAddr, prefix+"addr", "", "gRPC server bind address")

	flag.StringVar(&s.certFile, prefix+"cert-file", "", "tls certificate file")
	flag.StringVar(&s.keyFile, prefix+"key-file", "", "tls key file")

	// Tracing
	flag.Float64Var(
		&s.sampleTraceRating,
		prefix+"trace-sample-rate",
		defaultSampleRatingTrace,
		"sample rating for remote tracing from OpenSensus: 0.0 -> 1.0 (default is 1.0)",
	)

	flag.StringVar(
		&s.jaegerAgentURI,
		prefix+"jaeger-agent-uri",
		"",
		"jaeger agent URI to receive tracing data directly",
	)

	flag.BoolVar(
		&s.stdTracingEnabled,
		prefix+"trace-std-enabled",
		defaultEnableStdTracing,
		"enable tracing export to std (default is false)",
	)
}

func (s *grpcServiceImpl) Configure() error {
	var opts []grpc.ServerOption

	s.connectToJaegerAgent()
	// Add stat handler for recording span
	s.cfg.ServerOptions = append(s.cfg.ServerOptions, grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	opts = append(opts, s.cfg.ServerOptions...)

	if s.certFile != "" && s.keyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(s.certFile, s.keyFile)
		if err != nil {
			s.log.Errorf("Failed to load cert %v", err)
			return err
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	s.svr = grpc.NewServer(opts...)

	s.cfg.RegisterFunc(s.svr)

	return nil
}

func formatBindAddr(s string, p int) string {
	if strings.Contains(s, ":") && !strings.Contains(s, "[") {
		s = "[" + s + "]"
	}
	return fmt.Sprintf("%s:%d", s, p)
}

func (s *grpcServiceImpl) Run() error {
	addr := formatBindAddr(s.bindAddr, s.port)
	s.log.Debugf("start listen tcp %s...", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.log.Errorf("failed to listen: %v", err)
		return err
	}

	s.port = getPort(lis)
	s.muReady.Unlock()

	s.updateServiceConfig()
	s.registerSD()

	defer s.sd.RemoveAll()

	s.log.Infof("listen on %s...", lis.Addr())
	return s.svr.Serve(lis)
}

func getPort(lis net.Listener) int {
	addr := lis.Addr()
	tcp, _ := net.ResolveTCPAddr(addr.Network(), addr.String())
	return tcp.Port
}

// Connect to Jaeger Agent to send tracing data
func (s *grpcServiceImpl) connectToJaegerAgent() {
	if s.jaegerAgentURI == "" {
		return
	}

	s.log.Infof("connecting to Jaeger Agent on %s...", s.jaegerAgentURI)

	je, _ := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: s.jaegerAgentURI,
		Process:       jaeger.Process{ServiceName: s.name},
	})

	// And now finally register it as a Trace Exporter
	trace.RegisterExporter(je)

	// Trace view for console
	if s.stdTracingEnabled {
		// Register stats and trace exporters to export
		// the collected data.
		view.RegisterExporter(&PrintExporter{})

		// Register the views to collect server request count.
		if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
			s.log.Errorf("jaeger error: %s", err.Error())
		}
	}

	if s.sampleTraceRating >= 1 {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	} else {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(s.sampleTraceRating)})
	}
}

// get service running port
// only available when service ready
func (s *grpcServiceImpl) Port() int {
	s.muReady.Lock()
	defer s.muReady.Unlock()
	return s.port
}

func (s *grpcServiceImpl) Cleanup() {
	s.sd.RemoveAll()
}

func (s *grpcServiceImpl) Stop() {
	s.sd.RemoveAll()
	if s.svr == nil {
		return
	}

	go func() {
		time.Sleep(s.cfg.StopTimeout)
		s.svr.Stop()
	}()
	s.svr.GracefulStop()
}

func (s *grpcServiceImpl) registerSD() {
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

func (s *grpcServiceImpl) updateServiceConfig() {
	if s.cfg.NoRegRoute {
		return
	}

	s.fixSvcConfig()
	s.sd.UpdateServiceConfig(&s.cfg.SvcConfig)
}

func (s *grpcServiceImpl) fixSvcConfig() {
	sr := &s.cfg.SvcConfig
	sr.Name = s.cfg.ServiceName
	sr.Protocol = "grpc"

	if sr.Prefix != "" || sr.Regex != "" {
		return
	}

	si := s.svr.GetServiceInfo()
	if len(si) == 0 {
		s.log.Fatal("No grpc service has been registered")
	}
	keys := make([]string, 0, len(si))
	for k := range si {
		keys = append(keys, k)
	}

	if len(keys) == 1 {
		sr.Prefix = "/" + keys[0] + "/"
		return
	}

	var idx = findCommonPrefix(keys)
	sr.Regex = "/" + regexp.QuoteMeta(keys[0][:idx+1]) + "("

	for i := range keys {
		keys[i] = regexp.QuoteMeta(keys[i][idx+1:])
	}
	sort.Strings(keys)
	sr.Regex += strings.Join(keys, "|") + ")/.*"
}

func (s *grpcServiceImpl) isGrpcService() {}
