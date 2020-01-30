package ssd

import (
	"testing"

	sdms "gitlab.sendo.vn/core/golang-sdk"
)

func TestInterface(*testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{})
	_ = sdms.RunnableService(NewConsul(app))
}
