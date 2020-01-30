package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/viettranx/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/joho/godotenv"

	gw "gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

var (
	PORT_GATEWAY = "8000"
	PORT_GRPC    = "10000"
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

	if PORT_GRPC == "" {
		PORT_GRPC = "10000"
	}

	if PORT_GATEWAY == "" {
		PORT_GATEWAY = "8000"
	}
}

func run() error {
	loadConfig()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := gw.RegisterNoteServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf(":%s", PORT_GRPC), opts)
	if err != nil {
		return err
	}

	log.Printf("Running gateway server at: :%s", PORT_GATEWAY)
	return http.ListenAndServe(fmt.Sprintf(":%s", PORT_GATEWAY), mux)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
