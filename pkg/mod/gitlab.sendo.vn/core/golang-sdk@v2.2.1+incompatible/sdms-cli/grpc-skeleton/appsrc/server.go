package appsrc

import (
	"google.golang.org/grpc"

	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
	"{{ .ImportPath }}/appsrc/note_service"
)

func registerServices(g *grpc.Server) {
	demo.RegisterNoteServiceServer(g, note_service.NewNoteServer())
}
