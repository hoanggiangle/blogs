package sgorm

import (
	"flag"

	"github.com/jinzhu/gorm"
	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/ssql"
)

var (
	haveInitGlobalFlag bool
	gormDebug          bool
)

func initGlobalFlag() {
	if haveInitGlobalFlag {
		return
	}

	flag.BoolVar(&gormDebug, "gorm-debug", false, "Turn on gorm.Debug (global)")
	haveInitGlobalFlag = true
}

type GormSqlService interface {
	sdms.Service
	DriverName() string

	GormDB() *gorm.DB
}

func NewSqlService(config *ssql.SqlConfig) GormSqlService {
	return &gormSqlService{
		SqlService: ssql.NewSqlService(config),
	}
}

type gormSqlService struct {
	ssql.SqlService
	gormDb *gorm.DB
}

func (g *gormSqlService) InitFlags() {
	g.SqlService.InitFlags()
	initGlobalFlag()
}

func (g *gormSqlService) GormDB() *gorm.DB {
	if g.gormDb == nil {
		g.gormDb, _ = gorm.Open(g.DriverName(), g.DB())
	}
	return g.gormDb
}

type GormSqlManagerService interface {
	ssql.SqlManagerService

	GormDB(dbId ssql.DatabaseID) (*gorm.DB, error)
}

func NewSqlManagerService(config *ssql.SqlManagerServiceConfig) GormSqlManagerService {
	return &gormSqlManagerServiceImpl{
		SqlManagerService: ssql.NewSqlManagerService(config),

		gPool: make(map[ssql.DatabaseID]*gorm.DB),
	}
}

type gormSqlManagerServiceImpl struct {
	ssql.SqlManagerService

	gPool map[ssql.DatabaseID]*gorm.DB
}

func (g *gormSqlManagerServiceImpl) InitFlags() {
	g.SqlManagerService.InitFlags()
	initGlobalFlag()
}

func (g *gormSqlManagerServiceImpl) GormDB(dbId ssql.DatabaseID) (*gorm.DB, error) {
	if gormDb, ok := g.gPool[dbId]; ok {
		return gormDb, nil
	}

	drv, err := g.DriverName(dbId)
	if err != nil {
		return nil, err
	}

	db, err := g.DB(dbId)
	if err != nil {
		return nil, err
	}

	gormDb, err := gorm.Open(drv, db)
	if err == nil {
		if gormDebug {
			gormDb = gormDb.Debug()
		}
		g.gPool[dbId] = gormDb
	}

	return gormDb, err
}
