package tests

import (
	"os"
	"strings"
	"testing"
	"time"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/slog"
	"gitlab.sendo.vn/core/golang-sdk/ssql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var sqlManCfg string

func init() {
	sqlManCfg = `
user: __USER__
pass: __PWD__
# max_idle_conns: 10
members:
- servers:
  - __HOST__
  dbname: test_dbmain
  dbrange: 0
- servers:
  - __HOST__
  dbname: test_db%02d  # will be format like db00123
  dbrange: 1-4  # can be like "1,2-9,10-16"
  dbrange_exhausted: 2-4 # or use "all"
- servers:
  - __HOST__
  dbname: test_db%02d  # will be format like db00123
  dbrange: 5-8  # can be like "1,2-9,10-16"
  dbrange_exhausted: 8 # or use "all"
`
	sqlManCfg = strings.Replace(sqlManCfg, "__USER__", os.Getenv("TEST_SQL_USER"), 1)
	sqlManCfg = strings.Replace(sqlManCfg, "__PWD__", os.Getenv("TEST_SQL_PASS"), 1)
	sqlManCfg = strings.Replace(sqlManCfg, "__HOST__", os.Getenv("TEST_SQL_HOST"), -1)
}

func TestSQLManager_sqlite(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{
			"-sqlman-config-string", sqlManCfg,
		},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "warn",
		},
		UseNewFlagSet: true,
	})

	sqlm := ssql.NewSqlManagerService(&ssql.SqlManagerServiceConfig{
		App:            app,
		SqlUriTemplate: "sqlite3://:memory:",
	})
	app.RegService(sqlm)

	nullS := &nullService{}
	app.RegMainService(nullS)

	defer executeApp(app)()

	nullS.WaitRun()

	for i := ssql.DatabaseID(0); i < 5; i++ {
		db, err := sqlm.DB(i)
		if err != nil {
			t.Fatal(err)
		}

		err = db.Ping()
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Millisecond * 500)
	}
}

func TestSQLManager_mysql(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{
			"-sqlman-config-string", sqlManCfg,
		},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "warn",
		},
		UseNewFlagSet: true,
	})

	sqlm := ssql.NewSqlManagerService(&ssql.SqlManagerServiceConfig{
		App: app,
	})
	app.RegService(sqlm)

	nullS := &nullService{}
	app.RegMainService(nullS)

	defer executeApp(app)()

	nullS.WaitRun()

	defer func() {
		for _, info := range sqlm.GetAllDatabaseInfo() {
			db, err := sqlm.DB(info.ID)
			if err != nil {
				t.Fatal(err)
			}
			db.Exec("DROP DATABASE " + info.DBName)
		}
	}()

	for _, info := range sqlm.GetAllDatabaseInfo() {
		db, err := sqlm.DB(info.ID)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY)`)
		if err != nil {
			t.Error(err)
		}
	}
}
