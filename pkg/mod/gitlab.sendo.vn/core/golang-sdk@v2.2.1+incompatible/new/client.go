package sendo

import (
	"flag"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"google.golang.org/grpc"
)

type client struct {
	server Server
	uri    string
	conn   *grpc.ClientConn
	logger logger.Logger
}

func defaultClient(server Server) *client {
	c := &client{
		server: server,
		logger: logger.DEFAULT_LOGGER_SERVICE.GetLogger("client"),
	}
	c.initFlags()
	return c
}

func (c *client) initFlags() {
	flag.StringVar(&c.uri, "grpc-endpoint", "", "address of gRPC server for connecting to")
}

func (c *client) Get(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if c.uri == "" {
		c.uri = c.server.URI()
	}

	c.logger.Infof("gRPC client is connecting to %s", c.uri)
	return grpc.Dial(c.uri, opts...)
}

func (c *client) Disconnect() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
