package storage

import (
	"errors"
	"github.com/globalsign/mgo"
	goRedis "github.com/go-redis/redis"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/olivere/elastic"
)

type HasPrefix interface {
	GetPrefix() string
}

type Storage interface {
	Mgo() *mgo.Session
	Redis() *redis.Pool
	GoRedis() *goRedis.Client // don't close it
	ElasticSearch() *elastic.Client
	GormDB() *gorm.DB
	FromPrefix(prefix string) (db interface{}, err error)
}

type storage struct {
	dbs []HasPrefix
	*mongoDB
	*redisDB
	*goRedisDB
	*es
	*gormDB
}

type Drivers struct {
	MongoDB   *mongoDB
	RedisDB   *redisDB
	GoRedisDB *goRedisDB
	ES        *es
	GormDB    *gormDB
}

func New(drivers Drivers, dbs []HasPrefix) *storage {
	return &storage{
		dbs,
		drivers.MongoDB,
		drivers.RedisDB,
		drivers.GoRedisDB,
		drivers.ES,
		drivers.GormDB,
	}
}

func (s *storage) FromPrefix(prefix string) (db interface{}, err error) {
	for _, db := range s.dbs {
		if db.GetPrefix() == prefix {
			switch v := db.(type) {
			case *mongoDB:
				return v.Mgo(), nil
			case *redisDB:
				return v.Redis(), nil
			case *goRedisDB:
				return v.GoRedis(), nil
			case *es:
				return v.ElasticSearch(), nil
			case *gormDB:
				return v.GormDB(), nil
			}
		}
	}

	return nil, errors.New("could not found DB with prefix: " + prefix)
}
