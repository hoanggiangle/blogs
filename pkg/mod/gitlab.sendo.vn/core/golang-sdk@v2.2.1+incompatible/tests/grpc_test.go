package tests

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/sgrpc"
	"gitlab.sendo.vn/core/golang-sdk/sgrpc/middlewares"
	"gitlab.sendo.vn/core/golang-sdk/slog"
	"gitlab.sendo.vn/core/golang-sdk/ssd"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

var testNoteListResponse = &demo.Notes{
	Total: 10,
	Notes: []*demo.Note{
		&demo.Note{Id: 1, Text: "example"},
		&demo.Note{Id: 2, Text: "example2"},
	},
}

type testNoteServer struct {
}

func (ns *testNoteServer) Add(ctx context.Context, req *demo.NoteAddReq) (*demo.Note, error) {
	return nil, status.Error(codes.Unimplemented, "TODO")
}

func (ns *testNoteServer) List(ctx context.Context, req *demo.NoteListReq) (*demo.Notes, error) {
	return testNoteListResponse, nil
}

func (ns *testNoteServer) Update(context.Context, *demo.Note) (*demo.Note, error) {
	return nil, status.Error(codes.Unimplemented, "TODO")
}

func (ns *testNoteServer) NotifyChanged(f *demo.NoteFilter, notify demo.NoteService_NotifyChangedServer) error {
	notify.Send(&demo.NoteChangedEvent{Type: demo.EventType_INSERTED})
	notify.Send(&demo.NoteChangedEvent{Type: demo.EventType_UPDATED})
	notify.Send(&demo.NoteChangedEvent{Type: demo.EventType_INSERTED})
	return nil
}

func doGrpcTest(t *testing.T, client demo.NoteServiceClient) {
	_, err := client.Add(context.Background(), &demo.NoteAddReq{})
	if err == nil {
		t.Fatal("Must have umimplemented error here")
	}

	stat := status.Convert(err)
	if stat.Code() != codes.Unimplemented {
		t.Fatal("Must be Umimplemented error:", err)
	}

	notes, err := client.List(context.Background(), &demo.NoteListReq{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(notes, testNoteListResponse) {
		t.Fatal("Wrong response data")
	}

	stream, err := client.NotifyChanged(context.Background(),
		&demo.NoteFilter{Type: demo.EventType_UPDATED})
	if err != nil {
		t.Fatal(err)
	}
	for {
		ev, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}

		_ = ev
	}
}

func testGrpcRequest(t *testing.T, port int) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), opts...)
	if err != nil {
		t.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := demo.NewNoteServiceClient(conn)

	doGrpcTest(t, client)
}

type testGrpcClientService struct {
	sgrpc.GrpcClientService

	noteAddr string
	noteCli  demo.NoteServiceClient
}

func (s *testGrpcClientService) InitFlags() {
	s.GrpcClientService.InitFlags()
	s.GrpcClientService.EndpointFlag(&s.noteAddr, "note", "")
}

func (s *testGrpcClientService) GetNoteClient() demo.NoteServiceClient {
	if s.noteCli == nil {
		conn, err := s.GetConnection(s.noteAddr)
		if err != nil {
			panic(err.Error())
		}
		s.noteCli = demo.NewNoteServiceClient(conn)
	}
	return s.noteCli
}

func TestGrpc(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{
			"-port", "0",
			"-addr", "127.0.0.1",
			"-core-loglevel", "warn",
		},
		UseNewFlagSet: true,
	})

	cliSvc := &testGrpcClientService{
		GrpcClientService: sgrpc.NewGrpcClientService(app),
	}
	app.RegService(cliSvc)

	gCnf := sgrpc.GrpcConfig{
		App: app,
		SD:  ssd.NewNullConsulSD(),
		RegisterFunc: func(grpcServer *grpc.Server) {
			demo.RegisterNoteServiceServer(grpcServer, &testNoteServer{})
		},
		ServiceName: "test-grpc",
	}
	grpcSvc := sgrpc.New(&gCnf)
	app.RegMainService(grpcSvc)

	defer executeApp(app)()

	testGrpcRequest(t, grpcSvc.Port())

	cliSvc.noteAddr = fmt.Sprintf("127.0.0.1:%d", grpcSvc.Port())
	cli := cliSvc.GetNoteClient()
	doGrpcTest(t, cli)
}

func TestGrpcLogging(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{
			"-port", "0",
			"-addr", "127.0.0.1",
			"-core-loglevel", "warn",
		},
		UseNewFlagSet: true,
	})

	unaryIntercep := middlewares.NewLoggingUnaryServerInterceptor(nil, slog.NewAnonLogger())
	streamIntercep := middlewares.NewLoggingStreamServerInterceptor(nil, slog.NewAnonLogger())

	gCnf := sgrpc.GrpcConfig{
		App: app,
		SD:  ssd.NewNullConsulSD(),
		RegisterFunc: func(grpcServer *grpc.Server) {
			demo.RegisterNoteServiceServer(grpcServer, &testNoteServer{})
		},
		ServiceName: "test-grpc",
		ServerOptions: []grpc.ServerOption{
			grpc.UnaryInterceptor(unaryIntercep),
			grpc.StreamInterceptor(streamIntercep),
		},
	}
	grpcSvc := sgrpc.New(&gCnf)
	app.RegMainService(grpcSvc)

	defer executeApp(app)()

	testGrpcRequest(t, grpcSvc.Port())
}
