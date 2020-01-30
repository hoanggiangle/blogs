package storage

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"gitlab.sendo.vn/core/golang-sdk/slog"
)

// just test notify, not really connect redis
func TestRedisWatchChanged(t *testing.T) {
	store := &redisNoteStorage{
		notifyChannels: map[context.Context]chan *ChangedEvent{},
		mu:             &sync.Mutex{},
		log:            slog.NewAnonLogger(),
	}

	e := &ChangedEvent{}
	store.notifyChanged(e)

	if len(store.notifyChannels) != 0 {
		t.Fatal("notifyChannels must be empty")
	}

	ctx, cancel := context.WithCancel(context.Background())
	ch := store.WatchChanged(ctx)
	if len(store.notifyChannels) != 1 {
		t.Fatal("notifyChannels must have one")
	}
	go store.notifyChanged(e)

	e2 := <-ch
	if e != e2 {
		t.Fatal("Wrong received ChangedEvent")
	}

	cancel()
	store.notifyChanged(e)

	if len(store.notifyChannels) != 0 {
		t.Fatal("notifyChannels must be empty")
	}

	_, ok := <-ch
	if ok {
		t.Fatal("Channel must be closed after cancel() + notify()")
	}
}

func TestRedisPageLimit(t *testing.T) {
	store := &redisNoteStorage{}

	data := []string{"a", "b", "d", "c"}

	if !reflect.DeepEqual(store.getPageLimit(data, 0, 2), data[:2]) {
		t.Fatal("getPageLimit return wrong data")
	}

	if !reflect.DeepEqual(store.getPageLimit(data, 1, 2), data[:2]) {
		t.Fatal("getPageLimit return wrong data")
	}

	if !reflect.DeepEqual(store.getPageLimit(data, 2, 2), data[2:4]) {
		t.Fatal("getPageLimit return wrong data")
	}

	if !reflect.DeepEqual(store.getPageLimit(data, 3, 2), []string{}) {
		t.Fatal("getPageLimit return wrong data")
	}
}
