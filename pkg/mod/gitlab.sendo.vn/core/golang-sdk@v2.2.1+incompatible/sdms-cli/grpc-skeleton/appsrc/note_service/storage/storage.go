package storage

import (
	"context"

	_ "{{ .ImportPath }}/appsrc/resprovider"
)

type Note struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type ChangedEvent struct {
	Type string
	Note *Note
}

type NoteStorage interface {
	Add(string) (*Note, error)
	List(page int32, limit int32) ([]*Note, error)
	Count() (int64, error)
	WatchChanged(context.Context) chan *ChangedEvent
}
