package repository

import (
	"context"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

type Repository interface {
	Add(text string) (*demo.Note, error)
	List(page int32, limit int32) ([]*demo.Note, int32, error)
	Update(id int64, text string) error
	FindById(id int64) (*demo.Note, error)
	Delete(int64) error
	WatchChanged(context.Context) chan demo.NoteChangedEvent
}
