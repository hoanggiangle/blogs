package handlers

import (
	"context"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/repository"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

func NewAddHandler(repo repository.Repository, ctx context.Context, req *demo.NoteAddReq) *addHandler {
	return &addHandler{
		repo:    repo,
		context: ctx,
		request: req,
	}
}

type addHandler struct {
	repo    repository.Repository
	context context.Context
	request *demo.NoteAddReq
}

func (ah *addHandler) Handle() (*demo.Note, error) {
	if ah.request.Text == "" {
		return nil, constant.ERR_EMPTY_TEXT
	}

	return ah.repo.Add(ah.request.Text)
}
