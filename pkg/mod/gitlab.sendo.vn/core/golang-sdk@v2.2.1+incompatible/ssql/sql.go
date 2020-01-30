package ssql

import (
	"database/sql"
	"flag"
	"fmt"
	"strings"
	"time"

	sdms "gitlab.sendo.vn/core/golang-sdk"
)

type SqlConfig struct {
	App sdms.Application

	// prefix to flag, used to difference multi instance
	FlagPrefix string

	DefaultURI string

	DefaultConnMaxLifetime int
	DefaultMaxIdleConns    int
	DefaultMaxOpenConns    int

	// embedded service: it will log less info
	Embedded bool
}

type SqlService interface {
	sdms.Service

	// DO NOT CALL CLOSE ON THIS
	DB() *sql.DB

	DriverName() string
}

type sqlServiceImpl struct {
	cfg SqlConfig
	log sdms.Logger

	db *sql.DB

	// flags
	sqlUri string

	connMaxLifetime int
	maxIdleConns    int
	maxOpenConns    int
}

func NewSqlService(config *SqlConfig) SqlService {
	cfg := *config
	if cfg.DefaultConnMaxLifetime == 0 {
		cfg.DefaultConnMaxLifetime = 120
	}

	return &sqlServiceImpl{
		cfg: cfg,
	}
}

func (s *sqlServiceImpl) logger() sdms.Logger {
	if s.log == nil {
		s.log = s.cfg.App.(sdms.SdkApplication).GetLog("sql")
	}
	return s.log
}

func (s *sqlServiceImpl) InitFlags() {
	flag.StringVar(&s.sqlUri, s.cfg.FlagPrefix+"sql-uri",
		s.cfg.DefaultURI, "SQL connection-string. Format: driver://dataSourceName")

	flag.IntVar(&s.connMaxLifetime, s.cfg.FlagPrefix+"sql-connmaxlifetime",
		s.cfg.DefaultConnMaxLifetime, "SQL connection max life time (second)")
	flag.IntVar(&s.maxIdleConns, s.cfg.FlagPrefix+"sql-maxidleconns",
		s.cfg.DefaultMaxIdleConns, "SQL max idle connections")
	flag.IntVar(&s.maxOpenConns, s.cfg.FlagPrefix+"sql-maxopenconns",
		s.cfg.DefaultMaxOpenConns, "SQL max open connections")
}

func (s *sqlServiceImpl) Configure() error {
	log := s.logger()

	if s.sqlUri == "" {
		err := fmt.Errorf("No config value for %s", s.cfg.FlagPrefix+"sql-uri")
		log.Error(err)
		return err
	}

	parts := strings.SplitN(s.sqlUri, "://", 2)
	if len(parts) != 2 {
		err := fmt.Errorf("Invalid %s: %s", s.cfg.FlagPrefix+"sql-uri", s.sqlUri)
		log.Error(err)
		return err
	}

	if !s.cfg.Embedded {
		log.Infof("Connect to %s%s...", s.cfg.FlagPrefix, s.DriverName())
	}
	log.Debugf("Connect to %s", s.sqlUri)

	db, err := sql.Open(parts[0], parts[1])

	if err != nil {
		if !s.cfg.Embedded {
			log.Error(err)
		}
		return err
	}

	if err = db.Ping(); err != nil {
		if !s.cfg.Embedded {
			log.Error(err)
		}
		return err
	}

	db.SetConnMaxLifetime(time.Second * time.Duration(s.connMaxLifetime))
	db.SetMaxIdleConns(s.maxIdleConns)
	db.SetMaxOpenConns(s.maxOpenConns)

	s.db = db

	return nil
}

func (s *sqlServiceImpl) Cleanup() {
	if s.db != nil {
		s.db.Close()
		s.db = nil
	}
}

func (s *sqlServiceImpl) DB() *sql.DB {
	return s.db
}

func (s *sqlServiceImpl) DriverName() string {
	return strings.SplitN(s.sqlUri, "://", 2)[0]
}
