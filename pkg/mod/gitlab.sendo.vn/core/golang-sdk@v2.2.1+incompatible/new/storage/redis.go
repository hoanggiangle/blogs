package storage

import (
	"flag"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	defaultMaxActive = 0 // 0 is unlimited max active connection
	defaultMaxIdle   = 10
)

type RedisDBOpt struct {
	RedisUri string
	Prefix   string
}

type redisDB struct {
	name   string
	pool   *redis.Pool
	logger logger.Logger
	*RedisDBOpt
}

func NewRedisDB(name, flagPrefix string) *redisDB {

	p := &redis.Pool{
		IdleTimeout: time.Minute,
	}

	r := &redisDB{
		name:   name,
		logger: logger.GetCurrent().GetLogger(name),
		pool:   p,
		RedisDBOpt: &RedisDBOpt{
			Prefix: flagPrefix,
		},
	}
	p.Dial = r.dialFunc

	return r
}

func (r *redisDB) GetPrefix() string {
	return r.Prefix
}

func (r *redisDB) isDisabled() bool {
	return r.RedisUri == ""
}

func (r *redisDB) InitFlags() {
	prefix := r.Prefix
	if r.Prefix != "" {
		prefix += "-"
	}

	flag.StringVar(&r.RedisUri, prefix+"redis-uri", "", "Redis connection-string. Ex: redis://localhost/0")
	flag.IntVar(&r.pool.MaxActive, prefix+"redis-pool-max-active", defaultMaxActive, "Override redis pool MaxActive")
	flag.IntVar(&r.pool.MaxIdle, prefix+"redis-pool-max-idle", defaultMaxIdle, "Override redis pool MaxIdle")
}

func (r *redisDB) Configure() error {
	if r.isDisabled() {
		return nil
	}

	r.logger.Info("Connecting to Redis at ", r.RedisUri, "...")

	// just test config
	c := r.pool.Get()

	if err := c.Err(); err != nil {
		r.logger.Error("Cannot connect Redis. ", err.Error())
		return err
	}

	_ = c.Close()

	return nil
}

func (r *redisDB) dialFunc() (redis.Conn, error) {
	return redis.DialURL(r.RedisUri)
}

func (r *redisDB) Name() string {
	return r.name
}

func (r *redisDB) Redis() *redis.Pool {
	return r.pool
}

func (r *redisDB) Run() error {
	return r.Configure()
}

func (r *redisDB) Stop() <-chan bool {
	if r.pool != nil {
		_ = r.pool.Close()
	}

	c := make(chan bool)
	go func() { c <- true }()
	return c
}
