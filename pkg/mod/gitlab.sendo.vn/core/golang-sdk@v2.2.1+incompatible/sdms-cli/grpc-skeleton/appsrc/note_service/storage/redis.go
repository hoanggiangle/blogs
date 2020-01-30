package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"

	"{{ .ImportPath }}/appsrc/resprovider"
)

const (
	NOTE_REDIS_PREFIX = "note:"
)

// NoteStorage that use redis only
//
// This implement only use for demo
type redisNoteStorage struct {
	log            resprovider.Logger
	notifyChannels map[context.Context]chan *ChangedEvent
	mu             *sync.Mutex
	redis          func() redis.Conn
}

func NewRedisNoteStorage() NoteStorage {
	return &redisNoteStorage{
		notifyChannels: map[context.Context]chan *ChangedEvent{},
		mu:             &sync.Mutex{},
		redis:          resprovider.GetInstance().Redis,
	}
}

func (store *redisNoteStorage) logger() resprovider.Logger {
	if store.log == nil {
		store.log = resprovider.GetInstance().Logger("note.storage")
	}
	return store.log
}

func (store *redisNoteStorage) Add(s string) (*Note, error) {
	r := store.redis()
	defer r.Close()

	n := &Note{
		ID:   time.Now().UnixNano(),
		Text: s,
	}

	b, _ := json.Marshal(n)
	_, err := r.Do("SET", NOTE_REDIS_PREFIX+fmt.Sprintf("%d", n.ID), b)
	if err != nil {
		return nil, err
	}

	store.notifyChanged(&ChangedEvent{
		Type: "i",
		Note: n,
	})

	return n, nil
}

func (store *redisNoteStorage) getPageLimit(data []string, page, limit int) []string {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	l := len(data)
	if endIndex > l {
		endIndex = l
	}
	if startIndex >= endIndex {
		return []string{}
	}
	return data[startIndex:endIndex]
}

func (store *redisNoteStorage) List(page int32, limit int32) ([]*Note, error) {
	r := store.redis()
	defer r.Close()

	keys, err := redis.Strings(r.Do("KEYS", NOTE_REDIS_PREFIX+"*"))
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	keys = store.getPageLimit(keys, int(page), int(limit))

	notes := make([]*Note, 0, len(keys))
	for _, k := range keys {
		b, err := redis.Bytes(r.Do("GET", k))
		if err != nil {
			return nil, err
		}

		n := &Note{}
		err = json.Unmarshal(b, n)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	return notes, nil
}

func (store *redisNoteStorage) Count() (int64, error) {
	r := store.redis()
	defer r.Close()

	keys, err := redis.Strings(r.Do("KEYS", NOTE_REDIS_PREFIX+"*"))
	if err != nil {
		return 0, err
	}

	return int64(len(keys)), nil
}

func (store *redisNoteStorage) WatchChanged(ctx context.Context) chan *ChangedEvent {
	store.mu.Lock()
	defer store.mu.Unlock()

	ch := make(chan *ChangedEvent, 1)
	store.notifyChannels[ctx] = ch

	return ch
}

func (store *redisNoteStorage) notifyChanged(e *ChangedEvent) {
	store.mu.Lock()
	defer store.mu.Unlock()

	for ctx, ch := range store.notifyChannels {
		if ctx.Err() == context.Canceled {
			delete(store.notifyChannels, ctx)
			close(ch)
			continue
		}
		ch <- e
	}
}
