package tests

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"

	"google.golang.org/grpc"

	"gitlab.sendo.vn/core/golang-sdk/sgrpc"
)

func initConnection(port int) (*grpc.ClientConn, error) {
	bind := os.Getenv("ADDR")
	if bind == "" || bind == "0.0.0.0" {
		bind = "127.0.0.1"
	}
	var err error
	if port <= 0 {
		port, err = strconv.Atoi(os.Getenv("PORT"))
		if err != nil || port <= 0 {
			port = 10000
		}
	}
	addr := fmt.Sprintf("%s:%d", bind, port)

	return grpc.Dial(addr, grpc.WithInsecure())
}

func setupTest(t *testing.T) (*grpc.ClientConn, func()) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Setenv("PORT", "0")
	os.Setenv("CORE_LOGLEVEL", "warn")

	app := appsrc.CreateApp([]string{}, "")
	go app.Run()

	grpcSvc := app.GetMainService().(sgrpc.GrpcService)
	conn, err := initConnection(grpcSvc.Port())
	if err != nil {
		t.Fatalf("fail to dial: %v", err)
	}

	app.RegisterExitHandler(func() {
		conn.Close()
	})

	return conn, func() {
		<-app.Shutdown()
	}
}
