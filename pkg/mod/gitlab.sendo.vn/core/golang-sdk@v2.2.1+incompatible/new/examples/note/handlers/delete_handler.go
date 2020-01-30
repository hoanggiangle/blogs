package handlers

import (
	"context"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/repository"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

type deleteHandler struct {
	repo    repository.Repository
	context context.Context
	request *demo.DeleteReq
}

func NewDeleteHandler(repo repository.Repository, ctx context.Context, req *demo.DeleteReq) *deleteHandler {
	return &deleteHandler{
		repo:    repo,
		context: ctx,
		request: req,
	}
}

func (dh *deleteHandler) Handle() (*demo.DeleteRes, error) {
	req := dh.request

	if _, err := dh.repo.FindById(req.Id); err != nil {
		return nil, constant.ERR_NOTE_NOT_FOUND
	}

	if err := dh.repo.Delete(req.Id); err != nil {
		return nil, err
	}

	return &demo.DeleteRes{
		Success: true,
	}, nil
}
