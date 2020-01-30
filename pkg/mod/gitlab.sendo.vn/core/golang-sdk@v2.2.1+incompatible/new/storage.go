package sendo

import (
	"github.com/globalsign/mgo"
	"github.com/gomodule/redigo/redis"
)

type storage struct {
	mgo   *mgo.Session
	redis *redis.Pool
}

func (store *storage) Mgo() *mgo.Session {
	return store.mgo
}

func (store *storage) Redis() redis.Conn {
	return store.redis.Get()
}
