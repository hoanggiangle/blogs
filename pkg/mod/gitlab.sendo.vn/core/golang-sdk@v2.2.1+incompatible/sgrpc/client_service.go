package sgrpc

import (
	"flag"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	sdms "gitlab.sendo.vn/core/golang-sdk"
)

var (
	haveInitGlobalFlag bool
	defaultAddr        string
)

// Simple service for manage grpc connection pool
//
// Need optimize a lot.
// REF https://github.com/processout/grpc-go-pool
type GrpcClientService interface {
	sdms.Service
	// define flag for a grpc endpoint
	EndpointFlag(val *string, name string, desc string)
	// always open a new grpc.ClientConn, this connection will not track, you must closed by yourself
	OpenNewConnection(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	// get exists grpc.ClientConn to an address, if conn not exists or closed, make new conn
	GetConnection(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	// deprecated alias for GetConnection
	MakeConnection(addr string) *grpc.ClientConn
}

type grpcClientService struct {
	app sdms.Application

	// TODO use global conns map
	conns map[string]*grpc.ClientConn
	mu    sync.Mutex

	warnedMakeConn bool
}

func NewGrpcClientService(app sdms.Application) GrpcClientService {
	return &grpcClientService{
		app:   app,
		conns: make(map[string]*grpc.ClientConn),
	}
}

func (s *grpcClientService) logger() sdms.Logger {
	return s.app.(sdms.SdkApplication).GetLog("grpc.client")
}

func (s *grpcClientService) InitFlags() {
	if !haveInitGlobalFlag {
		flag.StringVar(&defaultAddr, "grpc-endpoint", "127.0.0.1:10000", "address of grpc server for connecting to")
		haveInitGlobalFlag = true
	}
}

func (g *grpcClientService) EndpointFlag(val *string, name string, desc string) {
	if desc == "" {
		desc = fmt.Sprintf("override default grpc-endpoint for %s service", name)
	}
	flag.StringVar(val, "grpc-endpoint-"+name, "", desc)
}

func (s *grpcClientService) Configure() error {
	return nil
}

func (s *grpcClientService) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, con := range s.conns {
		con.Close()
	}
	s.conns = nil
}

func (s *grpcClientService) OpenNewConnection(serverAddr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if strings.TrimSpace(serverAddr) == "" {
		serverAddr = defaultAddr
	}

	if len(opts) == 0 {
		opts = append(opts, grpc.WithInsecure())
	}

	log := s.logger()
	log.Info("New GRPC connection to ", serverAddr)

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *grpcClientService) GetConnection(serverAddr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if strings.TrimSpace(serverAddr) == "" {
		serverAddr = defaultAddr
	}

	var err error
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, ok := s.conns[serverAddr]
	if !ok || conn.GetState() == connectivity.Shutdown {
		conn, err = s.OpenNewConnection(serverAddr, opts...)
		if err != nil {
			return nil, err
		}
		s.conns[serverAddr] = conn
	}
	return conn, nil
}

func (s *grpcClientService) MakeConnection(serverAddr string) *grpc.ClientConn {
	log := s.logger()

	if !s.warnedMakeConn {
		log.Warn("MakeConnection is deprecated. Please replace with GetConnection")
		s.warnedMakeConn = true
	}
	c, err := s.GetConnection(serverAddr)
	if err != nil {
		log.Panicf("fail to dial: %v", err)
	}
	return c
}
