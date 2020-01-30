package http_server

import (
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"time"
)

type ServerHandler = func(*gin.Engine)

type HttpServer interface {
	Name() string
	InitFlags()
	Configure() error
	Run() error
	Stop() <-chan bool
	// Server
	AddHandler(ServerHandler)
	URI() string
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

type myHttpServer struct {
	http.Server
}

func (srv *myHttpServer) Serve(lis net.Listener) error {
	return srv.Server.Serve(tcpKeepAliveListener{lis.(*net.TCPListener)})
}

func (srv *myHttpServer) ServeTLS(lis net.Listener, certFile, keyFile string) error {
	return srv.Server.ServeTLS(tcpKeepAliveListener{lis.(*net.TCPListener)}, certFile, keyFile)
}
