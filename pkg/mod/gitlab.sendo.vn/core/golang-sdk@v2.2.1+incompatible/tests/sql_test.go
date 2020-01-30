package tests

import (
	"fmt"
	"os"
	"testing"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/slog"
	"gitlab.sendo.vn/core/golang-sdk/ssql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func doSimpleSqlTest(t *testing.T, app sdms.Application) {
	sqls := ssql.NewSqlService(&ssql.SqlConfig{
		App: app,
	})
	app.RegService(sqls)
	sqls2 := ssql.NewSqlService(&ssql.SqlConfig{
		App:        app,
		FlagPrefix: "abc-",
	})
	app.RegService(sqls2)

	nullS := &nullService{}
	app.RegMainService(nullS)

	defer executeApp(app)()

	nullS.WaitRun()

	db := sqls.DB()
	err := db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow("SELECT 1+1")
	var n int
	if err = row.Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatal("Result must be 2")
	}
}

func TestSQL_sqlite(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{
			"-sql-uri", "sqlite3://:memory:",
			"-abc-sql-uri", "sqlite3://:memory:",
		},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "warn",
		},
		UseNewFlagSet: true,
	})

	doSimpleSqlTest(t, app)
}

func TestSQL_mysql(t *testing.T) {
	sqlUrl := fmt.Sprintf("mysql://%s:%s@tcp(%s)/%s",
		os.Getenv("TEST_SQL_USER"),
		os.Getenv("TEST_SQL_PASS"),
		os.Getenv("TEST_SQL_HOST"),
		os.Getenv("TEST_SQL_DB"))

	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{
			"-sql-uri", sqlUrl,
			"-abc-sql-uri", sqlUrl,
		},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "warn",
		},
		UseNewFlagSet: true,
	})

	doSimpleSqlTest(t, app)
}
