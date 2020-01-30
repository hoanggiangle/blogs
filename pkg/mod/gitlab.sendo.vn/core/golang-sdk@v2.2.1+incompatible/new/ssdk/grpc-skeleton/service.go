package main

import (
	"context"
	"gitlab.sendo.vn/core/golang-sdk/new"
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
	return &demo.Note{
		Id:   100,
		Text: req.Text,
	}, nil
}

func (ns *noteService) List(sdCtx sendo.ServiceContext, ctx context.Context, req *demo.NoteListReq) (*demo.Notes, error) {
	return &demo.Notes{
		Total: 1,
		Notes: []*demo.Note{
			{
				Id:   100,
				Text: "This is a test note",
			},
		},
	}, nil
}

func (ns *noteService) Update(sdCtx sendo.ServiceContext, ctx context.Context, note *demo.Note) (*demo.Note, error) {
	return note, nil
}

func (ns *noteService) Delete(sdCtx sendo.ServiceContext, ctx context.Context, req *demo.DeleteReq) (*demo.DeleteRes, error) {
	return &demo.DeleteRes{Success: true}, nil
}

func (noteService) NotifyChanged(sendo.ServiceContext, *demo.NoteFilter, demo.NoteService_NotifyChangedServer) error {
	return nil
}
