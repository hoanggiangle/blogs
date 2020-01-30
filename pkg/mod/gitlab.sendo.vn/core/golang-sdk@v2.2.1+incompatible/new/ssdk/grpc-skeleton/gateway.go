package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/viettranx/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	gw "gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

var (
	PORT_GATEWAY = "8000"
	PORT_GRPC    = "10000"
	HOST_GRPC    = "localhost"
)

func loadConfig() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	_, err := os.Stat(envFile)
	if err == nil {
		godotenv.Load(envFile)
	}

	PORT_GATEWAY = os.Getenv("PORT_GW")
	PORT_GRPC = os.Getenv("PORT")
	HOST_GRPC = os.Getenv("HOST_GRPC")

	if PORT_GRPC == "" {
		PORT_GRPC = "10000"
	}

	if PORT_GATEWAY == "" {
		PORT_GATEWAY = "8000"
	}

	if HOST_GRPC == "" {
		HOST_GRPC = "localhost"
	}
}

func run() error {
	loadConfig()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	serverMuxOptionHeader := runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		return key, true
	})

	mux := runtime.NewServeMux(serverMuxOptionHeader)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	gRPC_URI := fmt.Sprintf("%s:%s", HOST_GRPC, PORT_GRPC)
	err := gw.RegisterNoteServiceHandlerFromEndpoint(ctx, mux, gRPC_URI, opts)
	if err != nil {
		return err
	}

	log.Printf("Running gateway server at: :%s", PORT_GATEWAY)
	return http.ListenAndServe(fmt.Sprintf(":%s", PORT_GATEWAY), mux)
}

func startGateway() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
