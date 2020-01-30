// Redigo service
package sredigo

import (
	"flag"
	"time"

	"github.com/gomodule/redigo/redis"
	sdms "gitlab.sendo.vn/core/golang-sdk"
)

type RedigoConfig struct {
	App sdms.Application
	// prefix to flag, used to difference multi instance
	FlagPrefix string

	// =====================
	// below are copy from redis.Pool config
	// =====================

	// TestOnBorrow is an optional application supplied function for checking
	// the health of an idle connection before the connection is used again by
	// the application. Argument t is the time that the connection was returned
	// to the pool. If the function returns an error, then the connection is
	// closed.
	TestOnBorrow func(c redis.Conn, t time.Time) error

	// Maximum number of idle connections in the pool.
	MaxIdle int

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	MaxActive int

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration

	// If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning.
	Wait bool
}

type RedigoService interface {
	sdms.Service
	Get() redis.Conn
}

type redigoServiceImpl struct {
	cfg RedigoConfig

	pool *redis.Pool

	log sdms.Logger

	redisUri string
}

func NewRedigo(config *RedigoConfig) RedigoService {
	if config.MaxIdle == 0 {
		config.MaxIdle = 5
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = time.Minute
	}

	p := &redis.Pool{
		MaxIdle:      config.MaxIdle,
		MaxActive:    config.MaxActive,
		IdleTimeout:  config.IdleTimeout,
		TestOnBorrow: config.TestOnBorrow,
		Wait:         config.Wait,
	}
	r := &redigoServiceImpl{
		cfg:  *config,
		log:  config.App.(sdms.SdkApplication).GetLog("redigo"),
		pool: p,
	}
	p.Dial = r.dialFunc

	return r
}

func (r *redigoServiceImpl) InitFlags() {
	flag.StringVar(&r.redisUri, r.cfg.FlagPrefix+"redis-uri", "redis://localhost/0", "Redis connection-string")

	flag.IntVar(&r.pool.MaxActive, r.cfg.FlagPrefix+"redis-pool-max-active", r.pool.MaxActive, "Override redis pool MaxActive")
	flag.IntVar(&r.pool.MaxIdle, r.cfg.FlagPrefix+"redis-pool-max-idle", r.pool.MaxIdle, "Override redis pool MaxIdle")
}

func (r *redigoServiceImpl) Configure() error {
	r.log.Debug("init redigo pool...")
	// just test config
	c := r.pool.Get()
	defer c.Close()
	if err := c.Err(); err != nil {
		r.log.Error(err)
		return err
	}

	return nil
}

func (r *redigoServiceImpl) dialFunc() (redis.Conn, error) {
	return redis.DialURL(r.redisUri)
}

// Get get a redis connection from pool
//
// Must close connection after used
func (r *redigoServiceImpl) Get() redis.Conn {
	return r.pool.Get()
}

func (r *redigoServiceImpl) Cleanup() {
	if r.pool != nil {
		r.pool.Close()
	}
}
