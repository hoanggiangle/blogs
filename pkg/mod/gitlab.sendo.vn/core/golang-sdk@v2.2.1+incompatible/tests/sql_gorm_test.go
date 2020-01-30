package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/jinzhu/gorm"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/slog"
	"gitlab.sendo.vn/core/golang-sdk/ssql"
	sgorm "gitlab.sendo.vn/core/golang-sdk/ssql/gorm"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
	Desc  string `gorm:"type:varchar(255) CHARACTER SET utf8mb4"`
}

func TestSQL_mysql_gorm(t *testing.T) {
	sqlUrl := fmt.Sprintf("mysql://%s:%s@tcp(%s)/%s",
		os.Getenv("TEST_SQL_USER"),
		os.Getenv("TEST_SQL_PASS"),
		os.Getenv("TEST_SQL_HOST"),
		os.Getenv("TEST_SQL_DB"))

	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{
			"-sql-uri", sqlUrl,
		},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "warn",
		},
		UseNewFlagSet: true,
	})

	g := sgorm.NewSqlService(&ssql.SqlConfig{
		App: app,
	})
	app.RegService(g)

	nullS := &nullService{}
	app.RegMainService(nullS)

	defer executeApp(app)()

	nullS.WaitRun()

	db := g.GormDB()

	db.CommonDB().Exec("DROP TABLE products")
	db.AutoMigrate(&Product{})

	p := &Product{Code: "L1212", Price: 1000, Desc: "thông tin sản phẩm"}
	db.Create(p)
	if (p.ID) < 1 {
		t.Fatal("ID must be > 0")
	}
}

func TestSQLManager_mysql_gorm(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{
			"-sqlman-config-string", sqlManCfg,
		},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "warn",
		},
		UseNewFlagSet: true,
	})

	sqlm := sgorm.NewSqlManagerService(&ssql.SqlManagerServiceConfig{
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
		db, err := sqlm.GormDB(info.ID)
		if err != nil {
			t.Fatal(err)
		}

		ret := db.AutoMigrate(&Product{})
		if ret.Error != nil {
			t.Fatal(ret.Error)
		}

		p := &Product{Code: "L1212", Price: 1000}
		db.Create(p)
		if (p.ID) < 1 {
			t.Fatal("ID must be > 0")
		}
	}
}

/**** JUST DON"T KNOW WHY not working with sqlite */

// func TestSQL_sqlite_gorm(t *testing.T) {
// 	app := sdms.NewApp(&sdms.AppConfig{
// 		Args: []string{
// 			"-sql-uri", "sqlite3://:memory:",
// 		},
// 		LogConfig: &slog.LoggerConfig{
// 			DefaultLevel: "warn",
// 		},
// 		UseNewFlagSet: true,
// 	})

// 	g := sgorm.NewSqlService(&ssql.SqlConfig{
// 		App: app,
// 	})
// 	app.RegService(g)

// 	nullS := &nullService{}
// 	app.RegMainService(nullS)

// 	defer executeApp(app)()

// 	nullS.WaitRun()

// 	db := g.GormDB()

// 	db.AutoMigrate(&Product{})

// 	p := Product{Code: "L1212", Price: 1000}
// 	db.Create(p)

// 	// p = Product{}
// 	db.First(&p)
// 	// t.Log(p)
// 	// if (p.ID) != 123 {
// 	// 	t.Fatal("ID must be 123")
// 	// }
// }

// func TestSQLManager_sqlite_gorm(t *testing.T) {
// 	app := sdms.NewApp(&sdms.AppConfig{
// 		Args: []string{
// 			"-sqlman-config-string", sqlManCfg,
// 		},
// 		LogConfig: &slog.LoggerConfig{
// 			DefaultLevel: "warn",
// 		},
// 		UseNewFlagSet: true,
// 	})

// 	sqlm := sgorm.NewSqlManagerService(&ssql.SqlManagerServiceConfig{
// 		App:            app,
// 		SqlUriTemplate: "sqlite3://:memory:",
// 	})
// 	app.RegService(sqlm)

// 	nullS := &nullService{}
// 	app.RegMainService(nullS)

// 	defer executeApp(app)()

// 	nullS.WaitRun()

// 	for _, info := range sqlm.GetAllDatabaseInfo() {
// 		db, err := sqlm.GormDB(info.ID)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		db.AutoMigrate(&Product{})

// 		p := &Product{Code: "L1212", Price: 1000}
// 		db.Create(p)
// 	}
// }
