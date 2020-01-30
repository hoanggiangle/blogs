package main

import (
	"context"
	"gitlab.sendo.vn/core/golang-sdk/new"
	"gitlab.sendo.vn/core/golang-sdk/new/util"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
	"google.golang.org/grpc"
)

var (
	BUILD_DATE   string
	BUILD_BRANCH string
	BUILD_REV    string
	SERVICE_NAME = "note-service"
	VERSION      = "1.0"
)

func newService() sendo.Service {
	service := sendo.New(
		sendo.WithName(SERVICE_NAME),
		sendo.WithVersion(VERSION),
	)

	// AddNoteServiceHandler is a function generated from protoc-gen-sendo
	demo.AddNoteServiceHandler(service, NewNoteService())
	return service
}

func main() {
	util.SetBuildInfo(&util.BuildInfo{
		Date:     BUILD_DATE,
		Branch:   BUILD_BRANCH,
		Revision: BUILD_REV,
	})

	m := util.SimpleMain{}
	m.Add(&util.SimpleCommand{
		Name: "server",
		Desc: "Run " + SERVICE_NAME + " gRPC server",
		Func: func(args []string) {
			s := newService()

			// Init: Start Consul, DBs
			if err := s.Init(); err != nil {
				return
			}
			s.Start()
		},
	})

	m.Add(&util.SimpleCommand{
		Name: "outenv",
		Desc: "output all environment variables",
		Func: func(args []string) { newService().OutEnv() },
	})

	m.Add(&util.SimpleCommand{
		Name: "client",
		Desc: "run client",
		Func: func(args []string) { newService().RunClient(clientHandler) },
	})

	m.Add(&util.SimpleCommand{
		Name: "func",
		Desc: "run a test function",
		Func: func(args []string) {
			s := newService()

			// Init: Start Consul, DBs
			if err := s.Init(); err != nil {
				return
			}

			s.RunFunction(someFunc)
		},
	})

	m.Add(&util.SimpleCommand{
		Name: "gw",
		Desc: "run a gateway to bridge HTTP/1 -> HTTP/2",
		Func: func(args []string) { startGateway() },
	})
	m.Execute()
}

// A local Client to connect and test the server
// SSDK maintains the gRPC client connection and pass
// the ServiceContext so we can use Storage, Broker, etc.
func clientHandler(cc *grpc.ClientConn, ctx sendo.ServiceContext) error {
	noteClient := demo.NewNoteServiceClient(cc)

	// Call a RPC Method through the client connection
	note, err := noteClient.Add(context.Background(), &demo.NoteAddReq{Text: "This is a new note"})
	ctx.Logger("client").Info(note)
	return err
}

// This is demo a function call in SSDK
// It helps us create a scheduler job or run-and-exit easily
// The function receive a ServiceContext so we can use Storage, Broker, etc.
func someFunc(ctx sendo.ServiceContext) error {
	ctx.Logger("someFunc").Infoln("This is a demo function")
	return nil
}
