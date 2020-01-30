package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/constant"
	"gitlab.sendo.vn/core/golang-sdk/new/examples/note/mocks"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
	"math/rand"
	"testing"
)

var (
	testNote = demo.Note{
		Id:   12345,
		Text: "This is a test note",
	}
	testLoggerService logger.LoggerService = nil
)

func tearUp(t *testing.T) {
	loggerSv := logger.DEFAULT_LOGGER_SERVICE
	testLoggerService = loggerSv
}

func tearDown() {
}

func TestMgoRepo_Add(t *testing.T) {
	tearUp(t)
	defer tearDown()

	dataLayer := &mocks.MgoDataLayer{}
	mgoRepoTest := NewMgoRepository(dataLayer, testLoggerService.GetLogger("test"))

	text := "Some text"
	n := &demo.Note{
		Id:   int64(rand.Uint32()),
		Text: text,
	}

	dataLayer.On("Insert", mock.MatchedBy(func(note *demo.Note) bool {
		return n.Text == note.Text
	})).Return(nil)
	note, err := mgoRepoTest.Add(text)

	assert.Nil(t, err, "should be nil")
	assert.Equal(t, text, note.Text, "should be equal")

	// Simulate add error
	text = "Will error"
	dataLayer.On("Insert", mock.MatchedBy(func(note *demo.Note) bool {
		return note.Text == text
	})).Return(constant.ERR_NOTE_CANNOT_ADD)
	_, err = mgoRepoTest.Add(text)
	assert.Equal(t, constant.ERR_NOTE_CANNOT_ADD, err, "should be equal")

	dataLayer.AssertExpectations(t)
}

func TestMgoRepo_List(t *testing.T) {
	tearUp(t)
	defer tearDown()

	dataLayer := &mocks.MgoDataLayer{}
	mgoRepoTest := NewMgoRepository(dataLayer, testLoggerService.GetLogger("test"))

	argCondition := mock.MatchedBy(func(s interface{}) bool { return true })
	fillStruct := mock.MatchedBy(func(note *[]*demo.Note) bool {
		*note = append(*note, &testNote)
		return true
	})

	dataLayer.On("Count", argCondition).Return(10, nil)
	dataLayer.On("Find",
		argCondition,
		fillStruct,
		mock.MatchedBy(func(s *int) bool { return *s == 0 }),
		argCondition,
	).Return(nil)

	notes, total, _ := mgoRepoTest.List(1, 10)

	assert.Equal(t, int32(10), total, "should be equal")
	assert.Equal(t, 1, len(notes), "should be equal")

	// Simulate find error
	errorOffset := 10
	dataLayer.On("Find",
		argCondition,
		fillStruct,
		mock.MatchedBy(func(s *int) bool { return *s == errorOffset }),
		argCondition,
	).Return(constant.ERR_NOTE_CANNOT_LIST)
	_, _, err := mgoRepoTest.List(2, 10)
	assert.Equal(t, constant.ERR_NOTE_CANNOT_LIST, err, "should be equal")

	// Simulate count error
	dataLayer = &mocks.MgoDataLayer{}
	mgoRepoTest = NewMgoRepository(dataLayer, testLoggerService.GetLogger("test"))
	dataLayer.On("Count", argCondition).Return(0, constant.ERR_NOTE_NOT_FOUND)
	_, _, err = mgoRepoTest.List(1, 10)
	assert.Equal(t, constant.ERR_NOTE_NOT_FOUND, err, "should be equal")

	dataLayer.AssertExpectations(t)
}

func TestMgoRepo_FindById(t *testing.T) {
	tearUp(t)
	defer tearDown()

	// Simulate find ok
	dataLayer := &mocks.MgoDataLayer{}
	mgoRepoTest := NewMgoRepository(dataLayer, testLoggerService.GetLogger("test"))
	fillStruct := mock.MatchedBy(func(note *[]*demo.Note) bool {
		*note = append(*note, &testNote)
		return true
	})

	argCondition := mock.MatchedBy(func(s interface{}) bool { return true })
	dataLayer.On("Find",
		argCondition,
		fillStruct,
		argCondition,
		argCondition,
	).Once().Return(nil)

	note, err := mgoRepoTest.FindById(testNote.Id)
	assert.Equal(t, &testNote, note, "should be equal")
	assert.Nil(t, err, "should be nil")

	// Simalate find return error
	dataLayer.On("Find",
		argCondition,
		argCondition,
		argCondition,
		argCondition,
	).Once().Return(constant.ERR_NOTE_NOT_FOUND)
	_, err = mgoRepoTest.FindById(testNote.Id)
	assert.Equal(t, constant.ERR_NOTE_NOT_FOUND, err, "should be equal")

	// Simulate find return empty slice
	dataLayer.On("Find",
		argCondition,
		argCondition,
		argCondition,
		argCondition,
	).Once().Return(nil)

	_, err = mgoRepoTest.FindById(testNote.Id)
	assert.Equal(t, constant.ERR_NOTE_NOT_FOUND, err, "should be equal")
}

func TestMgoRepo_Delete(t *testing.T) {
	tearUp(t)
	defer tearDown()

	dataLayer := &mocks.MgoDataLayer{}
	mgoRepoTest := NewMgoRepository(dataLayer, testLoggerService.GetLogger("test"))
	argCondition := mock.MatchedBy(func(s interface{}) bool { return true })
	dataLayer.On("Delete", argCondition).Once().Return(nil)

	err := mgoRepoTest.Delete(testNote.Id)
	assert.Nil(t, err, "should be nil")

	// Simulate error
	dataLayer.On("Delete", argCondition).Once().Return(constant.ERR_NOTE_CANNOT_DELETE)
	err = mgoRepoTest.Delete(testNote.Id)
	assert.Equal(t, constant.ERR_NOTE_CANNOT_DELETE, err, "should be equal")
}

func TestMgoRepo_Update(t *testing.T) {
	tearUp(t)
	defer tearDown()

	newText := "Updated text"
	dataLayer := &mocks.MgoDataLayer{}
	mgoRepoTest := NewMgoRepository(dataLayer, testLoggerService.GetLogger("test"))
	argCondition := mock.MatchedBy(func(s interface{}) bool { return true })
	dataLayer.On("Update", argCondition, argCondition).Once().Return(nil)
	err := mgoRepoTest.Update(testNote.Id, newText)
	assert.Nil(t, err, "should be nil")

	// Simulate err
	dataLayer.On("Update", argCondition, argCondition).Once().Return(constant.ERR_NOTE_CANNOT_UPDATE)
	err = mgoRepoTest.Update(testNote.Id, newText)
	assert.Equal(t, constant.ERR_NOTE_CANNOT_UPDATE, err, "should be nil")
}

func TestMgoRepo_WatchChanged(t *testing.T) {
	mgoRepoTest := NewMgoRepository(nil, nil)

	if c := mgoRepoTest.WatchChanged(context.Background()); c == nil {
		t.Error("Expect a channel return")
	}
}
