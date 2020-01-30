package handlers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/mocks"
	"gitlab.sendo.vn/protobuf/internal-apis-go/base"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
	"testing"
)

func TestListHandler_Handle(t *testing.T) {
	repo := &mocks.Repository{}

	ctx := context.Background()
	req := &demo.NoteListReq{
		Pagination: &base.Pagination{
			Limit: 20,
			Page:  1,
		},
	}

	hdl := NewListHandler(repo, ctx, req)

	repo.On("List", req.Pagination.Page, req.Pagination.Limit).
		Return([]*demo.Note{{Id: 1, Text: "some text"}}, int32(1), nil)

	notes, _ := hdl.Handle()
	assert.NotNil(t, notes, "should be not nil")

	req.Pagination.Page = 0
	req.Pagination.Limit = 0
	repo.On("List", req.Pagination.Page, req.Pagination.Limit).
		Return(nil, int32(1), constant.ERR_NOTE_CANNOT_LIST)

	_, err := hdl.Handle()
	assert.Equal(t, constant.ERR_NOTE_CANNOT_LIST, err, "should be not nil")

	req.Pagination = nil
	notes, err = hdl.Handle()
	assert.NotNil(t, notes, "should be not nil")
	assert.Nil(t, err, "should be nil")

	repo.AssertExpectations(t)
}
