package note_service

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
	"{{ .ImportPath }}/appsrc/note_service/storage"
	"{{ .ImportPath }}/appsrc/resprovider"
)

type NoteServer struct {
	store     storage.NoteStorage
	chChanged chan storage.ChangedEvent
	log       resprovider.Logger
}

func NewNoteServer() demo.NoteServiceServer {
	ns := &NoteServer{
		store:     storage.NewRedisNoteStorage(),
		chChanged: make(chan storage.ChangedEvent),
	}

	return ns
}

func (ns *NoteServer) logger() resprovider.Logger {
	if ns.log == nil {
		ns.log = resprovider.GetInstance().Logger("note.storage")
	}
	return ns.log
}

func (ns *NoteServer) changedListener(storage.ChangedEvent) {

}

func (ns *NoteServer) Add(ctx context.Context, req *demo.NoteAddReq) (*demo.Note, error) {
	n, err := ns.store.Add(req.GetText())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &demo.Note{
			Id:   n.ID,
			Text: n.Text},
		nil
}

func (ns *NoteServer) List(ctx context.Context, req *demo.NoteListReq) (*demo.Notes, error) {
	t, err := ns.store.Count()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &demo.Notes{
		Total: t,
	}

	arr, err := ns.store.List(req.GetPagination().GetPage(), req.GetPagination().GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, n := range arr {
		note := &demo.Note{
			Id:   n.ID,
			Text: n.Text,
		}
		resp.Notes = append(resp.Notes, note)
	}

	return resp, nil
}

func (ns *NoteServer) Update(context.Context, *demo.Note) (*demo.Note, error) {
	return nil, status.Error(codes.Unimplemented, "TODO")
}

func (ns *NoteServer) Delete(ctx context.Context, d *demo.DeleteReq) (*demo.DeleteRes, error) {
	return nil, status.Error(codes.Unimplemented, "TODO")
}

func (ns *NoteServer) NotifyChanged(f *demo.NoteFilter, notify demo.NoteService_NotifyChangedServer) error {
	ctx, cancelWatch := context.WithCancel(context.Background())
	defer cancelWatch()
	ch := ns.store.WatchChanged(ctx)

	for {
		select {
		case <-notify.Context().Done():
			return notify.Context().Err()
		case e := <-ch:
			notify.Send(ns.storageChEv_demoChEv(e))
		}
	}
	return nil
}

func (ns *NoteServer) storageChEv_demoChEv(e *storage.ChangedEvent) *demo.NoteChangedEvent {
	de := demo.NoteChangedEvent{
		Note: &demo.Note{
			Id:   e.Note.ID,
			Text: e.Note.Text,
		},
	}
	switch strings.ToLower(e.Type[:1]) {
	case "i", "a": // insert, add
		de.Type = demo.EventType_INSERTED
	case "u": // update
		de.Type = demo.EventType_UPDATED
	default:
		log.Fatal("Unknown ChangedEvent type:", e.Type)
	}
	return &de
}
