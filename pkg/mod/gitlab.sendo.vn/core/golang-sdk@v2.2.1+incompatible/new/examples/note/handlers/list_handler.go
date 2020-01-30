package handlers

import (
	"context"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/repository"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

type listHandler struct {
	repo    repository.Repository
	context context.Context
	request *demo.NoteListReq
}

func NewListHandler(repo repository.Repository, ctx context.Context, req *demo.NoteListReq) *listHandler {
	return &listHandler{
		repo:    repo,
		context: ctx,
		request: req,
	}
}

func (lh *listHandler) Handle() (*demo.Notes, error) {
	req := lh.request
	var limit, page int32

	if req.Pagination == nil {
		limit = 20
		page = 1
	} else {
		limit = req.Pagination.Limit
		page = req.Pagination.Page
	}

	notes, total, err := lh.repo.List(page, limit)
	if err != nil {
		return nil, constant.ERR_NOTE_CANNOT_LIST
	}

	return &demo.Notes{
		Total: int64(total),
		Notes: notes,
	}, nil
}
