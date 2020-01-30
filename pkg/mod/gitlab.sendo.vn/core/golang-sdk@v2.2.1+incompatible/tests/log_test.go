package tests

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/slog"
)

func TestCoreLog(*testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "warn",
		},
		UseNewFlagSet: true,
	})

	nullS := &nullService{}
	app.RegMainService(nullS)

	defer executeApp(app)()

	nullS.WaitRun()
}

func TestAppLog(*testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	logSvc := slog.NewAppLogService(&slog.LoggerConfig{
		DefaultLevel: "debug",
	})
	app.RegService(logSvc)

	nullS := &nullService{}
	app.RegMainService(nullS)

	defer executeApp(app)()

	nullS.WaitRun()

	log := logSvc.GetLogger("test-log")

	log.Info("test info")
	log.WithSrc().Info("test info")
	log.Debug("test debug")
	log.WithSrc().WithSrc().Debug("test debug")
}

func TestNullLogger(*testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	logSvc := slog.NewAppLogService(&slog.LoggerConfig{
		DefaultLevel: "debug",
		ExcludePrefix: func(pre string) bool {
			if pre == "test-log" {
				return true
			}
			return false
		},
	})
	app.RegService(logSvc)

	nullS := &nullService{}
	app.RegMainService(nullS)

	defer executeApp(app)()

	nullS.WaitRun()

	log := logSvc.GetLogger("test-log")

	log.Info("test info")
	log.WithSrc().Info("test info")
	log.Debug("test debug")
	log.WithSrc().WithSrc().Debug("test debug")
	log.Warn("test warn")
	log.Error("test error")
}

func TestMsgLog(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	logSvc := slog.NewMessageLogService(nil, nil)
	app.RegService(logSvc)
	msgLog := logSvc.GetLogger("")

	ch := make(chan string)
	app.RegMainService(newConsumeService(msgLog, ch))

	logPath := "/tmp/test-msg-log"
	os.Remove(logPath)
	os.Setenv("FILE_LOGFILE", logPath)
	defer os.Unsetenv("MSG_LOGFILE")

	go app.Run()

	go func() {
		for i := 0; ; i++ {
			ch <- fmt.Sprintf("Log %d", i)
			time.Sleep(time.Millisecond)
		}
	}()

	time.Sleep(time.Millisecond * 10)

	os.Rename(logPath, logPath+".bak")
	// trigger reload
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	// app.Reload()

	time.Sleep(time.Millisecond * 10)

	<-app.Shutdown()

	testLogFile := func(file string) {
		st, err := os.Stat(file)
		if err != nil {
			t.Fatal("Test write log: ", err)
		}
		if st.Size() == 0 {
			t.Fatal("No log has been written:", file)
		}
	}
	testLogFile(logPath + ".bak")
	testLogFile(logPath)

	os.Remove(logPath + ".bak")
	os.Remove(logPath)
}

func TestMsgLogNoFile(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	logSvc := slog.NewMessageLogService(nil, nil)
	app.RegService(logSvc)
	msgLog := logSvc.GetLogger("")

	ch := make(chan string)
	app.RegMainService(newConsumeService(msgLog, ch))

	defer executeApp(app)()

	for i := 0; i < 2; i++ {
		ch <- fmt.Sprintf("Log %d", i)
		time.Sleep(time.Millisecond)
	}

	// trigger reload
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	// app.Reload()
}
