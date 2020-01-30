package repository

import (
	"context"
	"github.com/globalsign/mgo/bson"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/repository/data_layer"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
	"math/rand"
)

type mgoRepo struct {
	mdl    dataLayer.MgoDataLayer
	logger logger.Logger
}

func NewMgoRepository(mdl dataLayer.MgoDataLayer, logger logger.Logger) Repository {
	return &mgoRepo{
		mdl:    mdl,
		logger: logger,
	}
}

func (repo *mgoRepo) Add(text string) (*demo.Note, error) {
	note := &demo.Note{
		Id:   int64(rand.Uint32()),
		Text: text,
	}

	if err := repo.mdl.Insert(note); err != nil {
		repo.logger.Error(err)
		return nil, constant.ERR_NOTE_CANNOT_ADD
	}

	return note, nil
}

func (repo *mgoRepo) List(page int32, limit int32) ([]*demo.Note, int32, error) {
	offset := int((page - 1) * limit)
	l := int(limit)

	var notes []*demo.Note

	total, err := repo.mdl.Count(bson.M{})
	if err != nil {
		return nil, 0, err
	}

	if err := repo.mdl.Find(bson.M{}, &notes, &offset, &l); err != nil {
		repo.logger.Error(err)
		return nil, 0, constant.ERR_NOTE_CANNOT_LIST
	}

	return notes, int32(total), nil
}

func (repo *mgoRepo) Update(id int64, text string) error {

	if err := repo.mdl.Update(
		bson.M{"id": id},
		bson.M{"$set": bson.M{"text": text}},
	); err != nil {
		repo.logger.Error(err)
		return constant.ERR_NOTE_CANNOT_UPDATE
	}

	return nil
}

func (repo *mgoRepo) FindById(id int64) (*demo.Note, error) {
	var notes []*demo.Note

	if err := repo.mdl.Find(bson.M{"id": id}, &notes, nil, nil); err != nil {
		repo.logger.Error(err)
		return nil, constant.ERR_NOTE_NOT_FOUND
	}

	if len(notes) == 0 {
		return nil, constant.ERR_NOTE_NOT_FOUND
	}

	return notes[0], nil
}

func (repo *mgoRepo) Delete(id int64) error {
	if err := repo.mdl.Delete(bson.M{
		"id": id,
	}); err != nil {
		repo.logger.Error(err)
		return constant.ERR_NOTE_CANNOT_DELETE
	}

	return nil
}

func (repo mgoRepo) WatchChanged(context.Context) chan demo.NoteChangedEvent {
	return make(chan demo.NoteChangedEvent)
}
