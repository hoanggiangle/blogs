package storage

import (
	"errors"
	"flag"
	"github.com/jinzhu/gorm"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"gitlab.sendo.vn/core/golang-sdk/new/storage/gormdialects"
	"strings"
)

type GormDBType int

const (
	GormDBTypeMySQL GormDBType = iota + 1
	GormDBTypePostgres
	GormDBTypeSQLite
	GormDBTypeMSSQL
	GormDBTypeNotSupported
)

type GormOpt struct {
	Uri    string
	Prefix string
	DBType string
}

type gormDB struct {
	name      string
	logger    logger.Logger
	db        *gorm.DB
	isRunning bool
	*GormOpt
}

func NewGormDB(name, prefix string) *gormDB {
	return &gormDB{
		GormOpt: &GormOpt{
			Prefix: prefix,
		},
		name:      name,
		logger:    logger.GetCurrent().GetLogger(name),
		isRunning: false,
	}
}

func (gdb *gormDB) GetPrefix() string {
	return gdb.Prefix
}

func (gdb *gormDB) Name() string {
	return gdb.name
}

func (gdb *gormDB) InitFlags() {
	prefix := gdb.Prefix
	if gdb.Prefix != "" {
		prefix += "-"
	}

	flag.StringVar(&gdb.Uri, prefix+"gorm-db-uri", "", "Gorm database connection-string.")
	flag.StringVar(&gdb.DBType, prefix+"gorm-db-type", "", "Gorm database type (mysql, postgres, sqlite, mssql)")
}

func (gdb *gormDB) isDisabled() bool {
	return gdb.Uri == ""
}

func (gdb *gormDB) Configure() error {
	if gdb.isDisabled() || gdb.isRunning {
		return nil
	}

	dbType := getDBType(gdb.DBType)
	if dbType == GormDBTypeNotSupported {
		return errors.New("gorm database type is not supported")
	}

	gdb.logger.Info("Connect to Gorm DB at ", gdb.Uri, " ...")

	var err error
	gdb.db, err = getDBConn(dbType, gdb.Uri)
	if err != nil {
		gdb.logger.Error("Error connect to gorm database at ", gdb.Uri, ". ", err.Error())
		return err
	}
	gdb.isRunning = true

	return nil
}

func (gdb *gormDB) Run() error {
	return gdb.Configure()
}

func (gdb *gormDB) Stop() <-chan bool {
	if gdb.db != nil {
		_ = gdb.db.Close()
	}
	gdb.isRunning = false

	c := make(chan bool)
	go func() { c <- true }()
	return c
}

func (gdb *gormDB) GormDB() *gorm.DB {
	return gdb.db
}

func getDBType(dbType string) GormDBType {
	switch strings.ToLower(dbType) {
	case "mysql":
		return GormDBTypeMySQL
	case "postgres":
		return GormDBTypePostgres
	case "sqlite":
		return GormDBTypeSQLite
	case "mssql":
		return GormDBTypeMSSQL
	}

	return GormDBTypeNotSupported
}

func getDBConn(dbType GormDBType, uri string) (dbConn *gorm.DB, err error) {
	switch dbType {
	case GormDBTypeMySQL:
		return gormdialects.MysqlDB(uri)
	case GormDBTypePostgres:
		return gormdialects.PostgresDB(uri)
	case GormDBTypeSQLite:
		return gormdialects.SQLiteDB(uri)
	case GormDBTypeMSSQL:
		return gormdialects.MSSQLDB(uri)
	}

	return nil, nil
}
