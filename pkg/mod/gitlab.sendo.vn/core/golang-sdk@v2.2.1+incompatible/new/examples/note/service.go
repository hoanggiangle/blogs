package main

import (
	"context"
	"gitlab.sendo.vn/core/golang-sdk/new"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/handlers"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/repository"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/repository/data_layer"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

// NoteService is a interface generated from protoc-gen-sendo
// It adds one parameter ServiceContext to all RPC methods.
func NewNoteService() demo.NoteService {
	return &noteService{}
}

type noteService struct {
}

func (ns *noteService) Add(sdCtx sendo.ServiceContext, ctx context.Context, req *demo.NoteAddReq) (*demo.Note, error) {
	return handlers.NewAddHandler(getRepo(sdCtx), ctx, req).Handle()
}

func (ns *noteService) List(sdCtx sendo.ServiceContext, ctx context.Context, req *demo.NoteListReq) (*demo.Notes, error) {
	return handlers.NewListHandler(getRepo(sdCtx), ctx, req).Handle()
}

func (ns *noteService) Update(sdCtx sendo.ServiceContext, ctx context.Context, note *demo.Note) (*demo.Note, error) {
	return handlers.NewUpdateHandler(getRepo(sdCtx), ctx, note).Handle()
}

func (ns *noteService) Delete(sdCtx sendo.ServiceContext, ctx context.Context, req *demo.DeleteReq) (*demo.DeleteRes, error) {
	return handlers.NewDeleteHandler(getRepo(sdCtx), ctx, req).Handle()
}

func (noteService) NotifyChanged(sendo.ServiceContext, *demo.NoteFilter, demo.NoteService_NotifyChangedServer) error {
	return nil
}

func getRepo(sdCtx sendo.ServiceContext) repository.Repository {
	dl := dataLayer.NewMgoDataLayer(sdCtx.Storage().Mgo(), "test")
	return repository.NewMgoRepository(dl, sdCtx.Logger(SERVICE_NAME))
}
