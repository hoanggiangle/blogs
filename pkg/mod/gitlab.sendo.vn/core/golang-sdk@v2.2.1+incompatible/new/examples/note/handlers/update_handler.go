package handlers

import (
	"context"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/repository"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

type updateHandler struct {
	repo    repository.Repository
	context context.Context
	request *demo.Note
}

func NewUpdateHandler(repo repository.Repository, ctx context.Context, req *demo.Note) *updateHandler {
	return &updateHandler{
		repo:    repo,
		context: ctx,
		request: req,
	}
}

func (uh *updateHandler) Handle() (*demo.Note, error) {
	req := uh.request

	if _, err := uh.repo.FindById(req.Id); err != nil {
		return nil, constant.ERR_NOTE_NOT_FOUND
	}

	if err := uh.repo.Update(req.Id, req.Text); err != nil {
		return nil, constant.ERR_NOTE_CANNOT_UPDATE
	}

	return req, nil
}
