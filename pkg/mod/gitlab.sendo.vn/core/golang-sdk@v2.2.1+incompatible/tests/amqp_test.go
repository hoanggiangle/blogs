package tests

import (
	"testing"
	"time"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/samqp"
	"gitlab.sendo.vn/core/golang-sdk/slog"
)

func TestAmqp(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	amqp := samqp.NewAmqp(&samqp.AmqpConfig{
		App: app,
	})
	app.RegService(amqp)

	app.RegMainService(&nullService{})

	defer executeApp(app)()

	time.Sleep(time.Millisecond * 50)
}
