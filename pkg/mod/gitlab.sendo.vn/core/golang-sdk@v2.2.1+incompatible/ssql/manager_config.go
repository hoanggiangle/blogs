package ssql

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var (
	DBRangeErrorAll       = errors.New(`DBRange can't be "all"`)
	DBRangeErrorEmpty     = errors.New(`DBRange must have at least one member`)
	DBRangeErrorNotSubSet = errors.New(`ExhaustedDBRange must be subset of DBRange`)
)

type DatabaseID uint

type ListDatabaseID []DatabaseID

func (l ListDatabaseID) Contains(id DatabaseID) bool {
	for _, i := range l {
		if id == i {
			return true
		}
	}
	return false
}

func (l ListDatabaseID) IsSubset(l2 *ListDatabaseID) bool {
	for _, i := range l {
		if !l2.Contains(i) {
			return false
		}
	}
	return true
}

func expandRange(s string, dbIds ListDatabaseID) (ListDatabaseID, error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, "-", 2)

	begin, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}
	if len(parts) < 2 {
		dbIds = append(dbIds, DatabaseID(begin))
		return dbIds, nil
	}
	end, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	for i := begin; i <= end; i++ {
		dbIds = append(dbIds, DatabaseID(i))
	}

	return dbIds, nil
}

func ParseListDatabaseID(s string) (dbIds ListDatabaseID, isAll bool, err error) {
	if strings.ToLower(s) == "all" {
		return dbIds, true, nil
	}
	for _, p := range strings.Split(s, ",") {
		if p == "" {
			continue
		}
		dbIds, err = expandRange(p, dbIds)
		if err != nil {
			return nil, false, err
		}
	}
	return
}

type sqlDatabasesConfig struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`

	ConnMaxLifetime int `yaml:"conn_max_life_time"`
	MaxIdleConns    int `yaml:"max_idle_conns"`
	MaxOpenConns    int `yaml:"max_open_conns"`

	Members []*sqlManagerMember `yaml:"members"`
}

// turn each shard to a separate pool
func (s *sqlDatabasesConfig) expandMembers() error {
	newMem := []*sqlManagerMember{}
	for _, m := range s.Members {
		exrange := m.GetExhaustedDBRange()
		for _, id := range m.GetDBRange() {
			nm := *m
			nm.DBName = nm.GetDBName(id)
			nm.DBRange = fmt.Sprintf("%d", id)
			if exrange.Contains(id) {
				nm.ExhaustedDBRange = nm.DBRange
			} else {
				nm.ExhaustedDBRange = ""
			}
			err := nm.parse()
			if err != nil {
				return err
			}
			newMem = append(newMem, &nm)
		}
	}
	s.Members = newMem

	return nil
}

type sqlManagerMember struct {
	Servers          []string `yaml:"servers"`
	DBRange          string   `yaml:"dbrange"`
	ExhaustedDBRange string   `yaml:"dbrange_exhausted,omitempty"`
	Description      string   `yaml:"desc,omitempty"`
	DBName           string   `yaml:"dbname"`

	// parsed dbrange*
	dbrange           ListDatabaseID
	dbrange_exhausted ListDatabaseID
}

func (s *sqlManagerMember) parse() error {
	var (
		isAll bool
		err   error
	)

	s.dbrange, isAll, err = ParseListDatabaseID(s.DBRange)
	if err != nil {
		return err
	}

	if isAll {
		return DBRangeErrorAll
	}

	if len(s.dbrange) == 0 {
		return DBRangeErrorEmpty
	}

	s.dbrange_exhausted, isAll, err = ParseListDatabaseID(s.ExhaustedDBRange)
	if err != nil {
		return err
	}
	if isAll {
		s.dbrange_exhausted = s.dbrange
	} else if !s.dbrange_exhausted.IsSubset(&s.dbrange) {
		return DBRangeErrorNotSubSet
	}

	if len(s.Servers) == 0 {
		return errors.New("Atleast one server must be defined")
	} else if len(s.Servers) > 1 {
		return errors.New("Multiple servers is not supported yet!")
	}
	return nil
}

func (s *sqlManagerMember) GetDBRange() ListDatabaseID {
	return s.dbrange
}

func (s *sqlManagerMember) GetExhaustedDBRange() ListDatabaseID {
	return s.dbrange_exhausted
}

func (s *sqlManagerMember) GetDBName(id DatabaseID) string {
	if strings.Contains(s.DBName, "%") {
		return fmt.Sprintf(s.DBName, id)
	} else {
		return s.DBName
	}
}

func loadSqlManagerConfig(b []byte) (*sqlDatabasesConfig, error) {
	conf := sqlDatabasesConfig{}
	err := yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, err
	}

	for _, m := range conf.Members {
		if err := m.parse(); err != nil {
			return nil, err
		}
	}

	return &conf, err
}
