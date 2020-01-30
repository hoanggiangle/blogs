package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/slog"
)

func TestServiceMuxExitNow(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	mux := sdms.NewServiceMux(
		&nullService{RunFunc: func() error { return nil }},
		&nullService{},
	)

	app.RegMainService(mux)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	go func() {
		app.Run()
		cancel()
	}()

	<-ctx.Done()

	if ctx.Err() != context.Canceled {
		t.Fatal("Something wrong with ServiceMux")
	}
}

func TestServiceMuxExitNowError(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	mux := sdms.NewServiceMux(
		&nullService{RunFunc: func() error { return errors.New("test") }},
		&nullService{},
	)

	app.RegMainService(mux)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	go func() {
		app.Run()
		cancel()
	}()

	<-ctx.Done()

	if ctx.Err() != context.Canceled {
		t.Fatal("Something wrong with ServiceMux")
	}
}

func TestServiceMuxForever(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	mux := sdms.NewServiceMux(
		&nullService{},
		&nullService{},
	)

	app.RegMainService(mux)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	go func() {
		app.Run()
		cancel()
	}()

	<-ctx.Done()

	if ctx.Err() != context.DeadlineExceeded {
		t.Fatal("Something wrong with ServiceMux")
	}
}
