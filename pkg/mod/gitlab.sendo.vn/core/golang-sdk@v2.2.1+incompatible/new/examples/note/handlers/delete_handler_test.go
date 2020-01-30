package handlers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/mocks"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
	"testing"
)

func TestDeleteHandler_Handle(t *testing.T) {
	repo := &mocks.Repository{}

	ctx := context.Background()
	req := &demo.DeleteReq{Id: 1}
	hdl := NewDeleteHandler(repo, ctx, req)

	repo.On("Delete", req.Id).Return(nil)
	repo.On("FindById", req.Id).Return(&demo.Note{
		Id:   req.Id,
		Text: "Some text",
	}, nil)

	actual, err := hdl.Handle()

	assert.NotNil(t, actual, "should be not nil")
	assert.Nil(t, err, "should be nil")
	assert.Equal(t, true, actual.Success, "should be nil")

	// Simulate id = 0
	req.Id = 0
	repo.On("FindById", req.Id).Return(nil, constant.ERR_NOTE_NOT_FOUND)
	_, err = hdl.Handle()
	assert.Equal(t, constant.ERR_NOTE_NOT_FOUND, err, "Should be error")

	// Simulate error when id = 10
	req.Id = 10
	repo.On("FindById", req.Id).Return(&demo.Note{
		Id:   req.Id,
		Text: "Some text",
	}, nil)
	repo.On("Delete", req.Id).Return(constant.ERR_NOTE_CANNOT_DELETE)

	_, err = hdl.Handle()
	assert.Equal(t, constant.ERR_NOTE_CANNOT_DELETE, err, "Should be error")

	repo.AssertExpectations(t)
}
