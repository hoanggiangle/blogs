package tests

import (
	"fmt"
	"testing"

	"github.com/globalsign/mgo/bson"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/slog"
	"gitlab.sendo.vn/core/golang-sdk/smgo"
)

type userInfo struct {
	name string
}

type mgoTest struct {
	nullService
	ms smgo.MgoService
	t  *testing.T
}

func (m *mgoTest) Run() error {
	t := m.t

	// just test close
	{
		sess := m.ms.Session()
		err := sess.Ping()
		if err != nil {
			t.Error(err)
			return nil
		}
		sess.Close()
	}

	{
		users, cleanup := m.ms.C("users")
		for i := 0; i < 5; i++ {
			inf := userInfo{name: fmt.Sprintf("user %d", i)}
			err := users.Insert(inf)
			if err != nil {
				t.Error(err)
				return nil
			}
		}

		count, err := users.Find(bson.M{}).Count()
		if err != nil {
			t.Error(err)
			return nil
		}
		if count != 5 {
			t.Error("Count must be 5")
		}
		cleanup()
	}

	{
		users, cleanup := m.ms.C("users", smgo.OptionNewConn{})
		err := users.Database.DropDatabase()
		if err != nil {
			t.Error(err)
			return nil
		}
		cleanup()
	}

	return nil
}

func TestMgo(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "error",
		},
		UseNewFlagSet: true,
	})

	mgo := smgo.NewMgoService(&smgo.MgoConfig{
		App:           app,
		DefaultDBName: "test-mgo",
	})
	app.RegService(mgo)

	app.RegMainService(&mgoTest{ms: mgo, t: t})

	app.Run()
}
