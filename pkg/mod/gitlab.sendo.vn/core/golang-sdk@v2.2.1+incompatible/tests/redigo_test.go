package tests

import (
	"testing"

	"github.com/gomodule/redigo/redis"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/slog"
	"gitlab.sendo.vn/core/golang-sdk/sredigo"
)

type redigoTest struct {
	nullService
	t *testing.T

	sred sredigo.RedigoService
}

func (m *redigoTest) Run() error {
	var err error

	c := m.sred.Get()
	defer c.Close()

	if err = c.Err(); err != nil {
		m.t.Fatal(err)
	}

	_, err = c.Do("SET", "hello", "world")
	if err != nil {
		m.t.Fatal(err)
	}
	defer c.Do("DEL", "hello")

	s, err := redis.String(c.Do("GET", "hello"))
	if err != nil {
		m.t.Fatal(err)
	}

	if s != "world" {
		m.t.Fatal("Can't get setted-value")
	}

	return nil
}
func TestRedigo(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	red := sredigo.NewRedigo(&sredigo.RedigoConfig{
		App: app,
	})
	app.RegService(red)

	app.RegMainService(&redigoTest{t: t, sred: red})

	app.Run()
}
