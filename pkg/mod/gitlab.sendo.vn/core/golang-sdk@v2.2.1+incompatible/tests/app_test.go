package tests

import (
	"testing"
	"time"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/slog"
)

type appTestService struct {
	nullService
	name  string
	sleep time.Duration
}

func (m *appTestService) Run() error {
	time.Sleep(m.sleep)
	return nil
}

func TestAppCriticalBefore(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	begin := time.Now()

	s1 := &appTestService{sleep: time.Millisecond * 50}
	app.RegService(s1)
	app.SetCriticalService(s1)

	app.RegMainService(&appTestService{sleep: time.Millisecond * 200})

	s2 := &appTestService{sleep: time.Millisecond * 50}
	app.RegService(s2)

	app.Run()

	if time.Now().Sub(begin) > time.Millisecond*100 {
		t.Fatal("Critical app MUST quit after 50ms")
	}
}

func TestAppCriticalAfter(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	begin := time.Now()

	s1 := &appTestService{sleep: time.Millisecond * 50}
	app.RegService(s1)

	app.RegMainService(&appTestService{sleep: time.Millisecond * 200})

	s2 := &appTestService{sleep: time.Millisecond * 50}
	app.RegService(s2)
	app.SetCriticalService(s2)

	app.Run()

	if time.Now().Sub(begin) > time.Millisecond*100 {
		t.Fatal("Critical app MUST quit after 50ms")
	}
}

func TestAppExitHandler(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	app.RegMainService(&nullService{})

	customIsCalled := false
	app.RegisterExitHandler(func() {
		customIsCalled = true
	})

	cancel := executeApp(app)
	cancel()

	if !customIsCalled {
		t.Fatal("app.RegisterExitHandler is not working")
	}
}
