package tests

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"gitlab.sendo.vn/protobuf/internal-apis-go/base"
	"gitlab.sendo.vn/protobuf/internal-apis-go/demo"
)

func TestNoteAdd(t *testing.T) {
	conn, tearDown := setupTest(t)
	defer tearDown()

	cli := demo.NewNoteServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	txt := "hello note!"
	resp, err := cli.Add(ctx, &demo.NoteAddReq{Text: txt})
	if err != nil {
		t.Fatalf("Add error: %s", err.Error())
	}

	if resp.GetId() == 0 {
		t.Fatal("Id must set on response!")
	}

	if resp.GetText() != txt {
		t.Fatalf(`resp.Text != "%s"`, txt)
	}
}

func TestNoteList(t *testing.T) {
	conn, tearDown := setupTest(t)
	defer tearDown()

	cli := demo.NewNoteServiceClient(conn)

	r1, _ := cli.Add(context.Background(), &demo.NoteAddReq{Text: "hello 1"})
	r2, _ := cli.Add(context.Background(), &demo.NoteAddReq{Text: "hello 2"})
	r3, _ := cli.Add(context.Background(), &demo.NoteAddReq{Text: "hello 3"})
	notes := []*demo.Note{r3, r2, r1}

	for i := 0; i < len(notes); i++ {
		resp, err := cli.List(context.Background(),
			&demo.NoteListReq{
				Pagination: &base.Pagination{
					Page:  int32(i + 1),
					Limit: 1,
				},
			},
		)
		if err != nil {
			t.Fatalf("List error: %s", err.Error())
		}

		if !reflect.DeepEqual(resp.Notes[0], notes[i]) {
			t.Fatal("Response notes is wrong")
		}
	}

	resp, err := cli.List(context.Background(),
		&demo.NoteListReq{
			Pagination: &base.Pagination{
				Limit: int32(len(notes)),
			},
		},
	)
	if err != nil {
		t.Fatalf("List error: %s", err.Error())
	}

	if !reflect.DeepEqual(resp.Notes, notes) {
		t.Fatal("Response notes is wrong")
	}
}

func TestNoteNotifyChanged(t *testing.T) {
	conn, tearDown := setupTest(t)
	defer tearDown()

	cli := demo.NewNoteServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	notiChange, err := cli.NotifyChanged(ctx, &demo.NoteFilter{})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for i := 0; i < 5; i++ {
			cli.Add(context.Background(), &demo.NoteAddReq{Text: fmt.Sprintf("hello %d", i)})
		}
	}()

	var count int
	for count < 5 {
		_, err := notiChange.Recv()
		if err != nil {
			t.Log(err)
			break
		}

		count++
	}
}
