package storage

// Go-Redis is an alternative option for Redis
// Github: https://github.com/go-redis/redis
//
// It supports:
// 		Redis 3 commands except QUIT, MONITOR, SLOWLOG and SYNC.
// 		Automatic connection pooling with circuit breaker support.
// 		Pub/Sub.
// 		Transactions.
// 		Pipeline and TxPipeline.
// 		Scripting.
// 		Timeouts.
// 		Redis Sentinel.
// 		Redis Cluster.
// 		Cluster of Redis Servers without using cluster mode and Redis Sentinel.
// 		Ring.
// 		Instrumentation.
// 		Cache friendly.
// 		Rate limiting.
// 		Distributed Locks.

import (
	"flag"
	"github.com/go-redis/redis"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
)

var (
	defaultGoRedisMaxActive = 0 // 0 is unlimited max active connection
	defaultGoMaxIdle        = 10
)

type GoRedisDBOpt struct {
	Prefix    string
	RedisUri  string
	MaxActive int
	MaxIde    int
}

type goRedisDB struct {
	name   string
	client *redis.Client
	logger logger.Logger
	*GoRedisDBOpt
}

func NewGoRedisDB(name, flagPrefix string) *goRedisDB {
	return &goRedisDB{
		name:   name,
		logger: logger.GetCurrent().GetLogger(name),
		GoRedisDBOpt: &GoRedisDBOpt{
			Prefix:    flagPrefix,
			MaxActive: defaultGoMaxIdle,
			MaxIde:    defaultGoRedisMaxActive,
		},
	}
}

func (r *goRedisDB) GetPrefix() string {
	return r.Prefix
}

func (r *goRedisDB) isDisabled() bool {
	return r.RedisUri == ""
}

func (r *goRedisDB) InitFlags() {
	prefix := r.Prefix
	if r.Prefix != "" {
		prefix += "-"
	}

	flag.StringVar(&r.RedisUri, prefix+"go-redis-uri", "", "(For go-redis) Redis connection-string. Ex: redis://localhost/0")
	flag.IntVar(&r.MaxActive, prefix+"go-redis-pool-max-active", defaultMaxActive, "(For go-redis) Override redis pool MaxActive")
	flag.IntVar(&r.MaxIde, prefix+"go-redis-pool-max-idle", defaultMaxIdle, "(For go-redis) Override redis pool MaxIdle")
}

func (r *goRedisDB) Configure() error {
	if r.isDisabled() {
		return nil
	}

	r.logger.Info("Connecting to Redis at ", r.RedisUri, "...")

	client := redis.NewClient(&redis.Options{
		Addr:         r.RedisUri,
		PoolSize:     r.MaxActive,
		MinIdleConns: r.MaxIde,
	})

	// Ping to test Redis connection
	if err := client.Ping().Err(); err != nil {
		r.logger.Error("Cannot connect Redis. ", err.Error())
		return err
	}

	// Connect successfully, assign client to goRedisDB
	r.client = client
	return nil
}

func (r *goRedisDB) Name() string {
	return r.name
}

func (r *goRedisDB) GoRedis() *redis.Client {
	return r.client
}

func (r *goRedisDB) Run() error {
	return r.Configure()
}

func (r *goRedisDB) Stop() <-chan bool {
	if r.client != nil {
		if err := r.client.Close(); err != nil {
			r.logger.Info("cannot close ", r.name)
		}
	}

	c := make(chan bool)
	go func() { c <- true }()
	return c
}
