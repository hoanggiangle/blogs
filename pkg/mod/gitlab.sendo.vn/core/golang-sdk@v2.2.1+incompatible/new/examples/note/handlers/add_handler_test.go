package handlers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/mocks"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
	"testing"
)

func TestAddHandler_Handle(t *testing.T) {
	repo := &mocks.Repository{}

	ctx := context.Background()
	req := &demo.NoteAddReq{Text: "note content"}
	hdl := NewAddHandler(repo, ctx, req)

	expect := &demo.Note{
		Id:   1,
		Text: req.Text,
	}
	repo.On("Add", req.Text).Return(expect, nil)

	actual, _ := hdl.Handle()

	assert.Equal(t, expect, actual, "should return a note")

	req.Text = ""
	_, err := hdl.Handle()
	assert.Equal(t, constant.ERR_EMPTY_TEXT, err, "should return a error")

	repo.AssertExpectations(t)
}
