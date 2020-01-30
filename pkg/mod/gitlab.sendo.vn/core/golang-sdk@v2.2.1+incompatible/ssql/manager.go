package ssql

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"

	sdms "gitlab.sendo.vn/core/golang-sdk"
)

var (
	DatabaseIDNotFound = errors.New("DatabaseID is not found!")
)

type SqlManagerServiceConfig struct {
	App sdms.Application

	DefaultConfigFile string

	SqlUriTemplate string
}

type DatabaseInfo struct {
	ID          DatabaseID
	DBName      string
	Description string
	Exhausted   bool

	pool SqlService
}

// A db manager that primary designed for MySQL Galera
type SqlManagerService interface {
	sdms.Service

	GetAllDatabaseInfo() []DatabaseInfo

	// DO NOT CALL CLOSE ON THIS
	DB(dbId DatabaseID) (*sql.DB, error)
	DriverName(dbId DatabaseID) (string, error)
}

type sqlManagerServiceImpl struct {
	cfg SqlManagerServiceConfig
	log sdms.Logger

	dbsCfg *sqlDatabasesConfig
	mu     *sync.Mutex

	dbPools map[DatabaseID]*DatabaseInfo

	// flags
	configFile   string
	configString string
}

func NewSqlManagerService(config *SqlManagerServiceConfig) SqlManagerService {
	return &sqlManagerServiceImpl{
		cfg: *config,
		mu:  &sync.Mutex{},
	}
}

func (s *sqlManagerServiceImpl) logger() sdms.Logger {
	if s.log == nil {
		s.log = s.cfg.App.(sdms.SdkApplication).GetLog("sql.mngr")
	}
	return s.log
}

func (s *sqlManagerServiceImpl) InitFlags() {
	if s.cfg.DefaultConfigFile == "" {
		s.cfg.DefaultConfigFile = "sqlman.yaml"
	}
	flag.StringVar(&s.configFile, "sqlman-config-file", s.cfg.DefaultConfigFile, "SQLManager config file (format: yaml)")
	flag.StringVar(&s.configString, "sqlman-config-string", "", "SQLManager config string (format: yaml). Prefer over file if set")
}

func (s *sqlManagerServiceImpl) loadConfig() (*sqlDatabasesConfig, error) {
	var data []byte
	var err error

	if s.configString == "" {
		data, err = ioutil.ReadFile(s.configFile)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"%s is not exist and sqlman-config-string is not set!",
				s.cfg.DefaultConfigFile)
		} else if err != nil {
			return nil, err
		}
	} else {
		data = []byte(s.configString)
	}

	return loadSqlManagerConfig(data)
}

func (s *sqlManagerServiceImpl) Configure() error {
	log := s.logger()

	if s.cfg.SqlUriTemplate == "" {
		s.cfg.SqlUriTemplate = "mysql://__USER__:__PWD__@tcp(__HOST__)/__DBNAME__?parseTime=true&charset=utf8mb4,utf8"
	}

	conf, err := s.loadConfig()
	if err != nil {
		log.Error(err)
		return err
	}
	s.dbsCfg = conf

	if err = conf.expandMembers(); err != nil {
		log.Error(err)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.dbPools = make(map[DatabaseID]*DatabaseInfo, len(conf.Members))

	log.Infof("Connect to %s servers(%d)...", strings.SplitN(s.cfg.SqlUriTemplate, "://", 2)[0], len(conf.Members))
	for _, m := range conf.Members {
		// configure sql service without default db
		sqlSvc := s.createSqlService(conf, m, false)
		if err = sqlSvc.Configure(); err != nil {
			log.Error(err)
			return err
		}

		id := m.GetDBRange()[0]
		exhausted := m.GetExhaustedDBRange().Contains(id)
		info := DatabaseInfo{
			ID:          id,
			DBName:      m.GetDBName(id),
			Exhausted:   exhausted,
			Description: m.Description,
			pool:        sqlSvc,
		}

		if info.pool.DriverName() != "sqlite3" {
			err := s.checkAndCreateDatabase(&info)
			if err != nil {
				log.Error(err)
				return err
			}

			// reconfigure sql service with a selected db
			sqlSvc.Cleanup()
			sqlSvc = s.createSqlService(conf, m, true)
			if err = sqlSvc.Configure(); err != nil {
				log.Error(err)
				return err
			}
			info.pool = sqlSvc
		}

		s.dbPools[id] = &info
	}

	return nil
}

func (s *sqlManagerServiceImpl) createSqlService(conf *sqlDatabasesConfig, m *sqlManagerMember, withDb bool) SqlService {
	uri := s.cfg.SqlUriTemplate
	uri = strings.Replace(uri, "__USER__", conf.User, 1)
	uri = strings.Replace(uri, "__PWD__", conf.Pass, 1)
	uri = strings.Replace(uri, "__HOST__", m.Servers[0], 1)

	if strings.HasPrefix(uri, "sqlite3") || withDb {
		uri = strings.Replace(uri, "__DBNAME__", m.GetDBName(m.GetDBRange()[0]), 1)
	} else {
		uri = strings.Replace(uri, "__DBNAME__", "", 1)
	}

	cfg := SqlConfig{
		App:      s.cfg.App,
		Embedded: true,
	}
	return &sqlServiceImpl{
		cfg:             cfg,
		sqlUri:          uri,
		connMaxLifetime: conf.ConnMaxLifetime,
		maxIdleConns:    conf.MaxIdleConns,
		maxOpenConns:    conf.MaxOpenConns,
	}
}

// Check if database exists, if not, try to create it
func (s *sqlManagerServiceImpl) checkAndCreateDatabase(info *DatabaseInfo) error {
	if info.pool.DriverName() == "sqlite3" {
		return nil
	}

	db := info.pool.DB()

	var dbName string
	row := db.QueryRow(fmt.Sprintf("SHOW DATABASES LIKE '%s'", info.DBName))
	if err := row.Scan(&dbName); err == nil {
		return nil
	} else if err != sql.ErrNoRows {
		return err
	}

	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", info.DBName))
	if err != nil {
		return fmt.Errorf(`You do not have permission to view or create database "%s": %s`, info.DBName, err.Error())
	}

	return nil
}

func (s *sqlManagerServiceImpl) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, i := range s.dbPools {
		i.pool.Cleanup()
	}

	s.dbPools = nil
}

func (s *sqlManagerServiceImpl) DB(dbId DatabaseID) (*sql.DB, error) {
	info, found := s.dbPools[dbId]
	if !found {
		return nil, DatabaseIDNotFound
	}
	return info.pool.DB(), nil
}

func (s *sqlManagerServiceImpl) DriverName(dbId DatabaseID) (string, error) {
	info, found := s.dbPools[dbId]
	if !found {
		return "", DatabaseIDNotFound
	}
	return info.pool.DriverName(), nil
}

type tmpListDbId []DatabaseInfo

func (t tmpListDbId) Len() int {
	return len(t)
}

func (t tmpListDbId) Less(i, j int) bool {
	return t[i].ID < t[j].ID
}

func (t tmpListDbId) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (s *sqlManagerServiceImpl) GetAllDatabaseInfo() []DatabaseInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	infos := []DatabaseInfo{}
	for _, info := range s.dbPools {
		infos = append(infos, *info)
	}
	sort.Sort(tmpListDbId(infos))
	return infos
}
