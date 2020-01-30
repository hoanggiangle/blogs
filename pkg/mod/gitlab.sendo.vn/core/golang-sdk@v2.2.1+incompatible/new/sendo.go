package sendo

import (
	"context"
	"github.com/gin-gonic/gin"
	"gitlab.sendo.vn/core/golang-sdk/new/broker"
	"gitlab.sendo.vn/core/golang-sdk/new/http-server"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"gitlab.sendo.vn/core/golang-sdk/new/registry"
	"gitlab.sendo.vn/core/golang-sdk/new/server"
	sdStorage "gitlab.sendo.vn/core/golang-sdk/new/storage"
	"google.golang.org/grpc"
)

// Convenience method for create a new SDK service
type Option func(*service)

// Handler function for gRPC server, this method is called
// right before the server start to serve
type ServerHandler = func(*grpc.Server)

// HTTP Server Handler for register some routes and gin handlers
type HttpServerHandler = func(*gin.Engine)

// A kind of server job for supporting developer
type Function = func(ServiceContext) error

// A convenience client connection to create a client connect to gRPC server
type ClientFunction = func(*grpc.ClientConn, ServiceContext) error

// Service Context: A wrapper for all things needed for developing a service
type ServiceContext interface {
	// Storage contains all DB drivers supported
	Storage() sdStorage.Storage
	// Broker is a client connect to pub/sub server
	Broker() Broker
	// Logger for a specific service, usually it has a prefix to distinguish
	// with each others
	Logger(prefix string) logger.Logger
}

// The heart of SDK, Service manages all components in SDK
type Service interface {
	// A part of Service is ServiceContext, it's passed to all handlers/functions
	ServiceContext
	// Name of the service
	Name() string
	// Version of the service
	Version() string
	// gRPC Server wrapper
	Server() Server
	// Gin HTTP Server wrapper
	HTTPServer() HttpServer
	// Client wrapper: support connect to local and remote gRPC
	Client() Client

	// This method will start all registry and database services
	Init() error
	// This method returns service if it is registered
	IsRegistered() bool
	// Start service and its all component.
	// It will be stopped if any service return error
	Start() error
	// Stop service and its all component.
	Stop()
	// Method to run a function user-defined
	RunFunction(Function) error
	// Method to run a ClientFunction, it often uses for test local gRPC server
	RunClient(ClientFunction, ...grpc.DialOption) error
	// Method export all flags to std/terminal
	// We might use: "> .env" to move its content .env file
	OutEnv()
}

// Runnable is an abstract object in SDK
// Almost components are Runnable. SDK will manage their lifecycle
type Runnable interface {
	Name() string
	InitFlags()
	Configure() error
	Run() error
	Stop() <-chan bool
}

// Trackable is an abstraction for any component need to health check
// and update configuration on Registry
type Trackable interface {
	Register(registry.Agent)
	CheckKV(registry.Agent)
	Deregister(registry.Agent)
	// IsRegistered()
}

// Server is a component to manage gRPC Server
type Server interface {
	// It's a runnable component
	Runnable
	// Server
	Get() *grpc.Server
	// Method to add some ServerOption before gRPC server serving
	AddOption(grpc.ServerOption)
	// Method to define which servers will be served for protobuf generated files
	AddHandler(ServerHandler)
	// Reload with new config
	Reload(config server.Config)
	// Return server config
	GetConfig() server.Config
	// URI that the server is listening
	URI() string
}

// GIN HTTP server for REST API
type HttpServer interface {
	Runnable
	// Add handlers to GIN
	AddHandler(HttpServerHandler)
	// Return server config
	GetConfig() http_server.Config
	// URI that the server is listening
	URI() string
}

// Client is an object for connecting to local or remote gRPC
type Client interface {
	// Get the client connection
	Get(...grpc.DialOption) (*grpc.ClientConn, error)
	// Disconnect for the gRPC server
	Disconnect()
}

// Broker is a heart in pub/sub system,
// it's broker client connect to pub/sub server by gRPC streaming
type Broker interface {
	// Public an event to pub/sub server
	Publish(*broker.Publishing) (string, error)
	// Subscribe events from pub/sub server
	// It returns a channel to receive the messages
	Subscribe(context.Context, *broker.SubscribeOption) (<-chan *broker.Message, error)
	// Check client if it connected
	IsConnected() bool
	// Get a publish token by an event name
	GetPubToken(e string) string
	// Get a subscribe token by an event name
	GetSubToken(e string) string
}

// Registry is an object to health check components
// and service discovery
type Registry interface {
	Runnable
	// Agent is an object supports all methods for health check
	// and updating configs for components
	registry.Agent
	// Channel to notify component need to reload
	SyncChan() <-chan bool
}
