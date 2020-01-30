package handlers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/mocks"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
	"testing"
)

func TestUpdateHandler_Handle(t *testing.T) {
	repo := &mocks.Repository{}

	ctx := context.Background()
	req := &demo.Note{
		Id:   1,
		Text: "Updated",
	}
	hdl := NewUpdateHandler(repo, ctx, req)

	repo.On("FindById", req.Id).Return(req, nil)
	repo.On("Update", req.Id, req.Text).Return(nil)
	actual, _ := hdl.Handle()

	assert.Equal(t, actual, req, "should be equal")

	req.Id = 10
	repo.On("FindById", req.Id).Return(nil, constant.ERR_NOTE_NOT_FOUND)
	_, err := hdl.Handle()
	assert.Equal(t, constant.ERR_NOTE_NOT_FOUND, err, "should be equal")

	req.Id = 5
	repo.On("FindById", req.Id).Return(req, nil)
	repo.On("Update", req.Id, req.Text).Return(constant.ERR_NOTE_CANNOT_UPDATE)
	_, err = hdl.Handle()
	assert.Equal(t, constant.ERR_NOTE_CANNOT_UPDATE, err, "should be equal")

	repo.AssertExpectations(t)
}
