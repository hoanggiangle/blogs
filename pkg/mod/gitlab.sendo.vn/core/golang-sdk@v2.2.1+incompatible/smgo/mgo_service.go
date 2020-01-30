// mgo service
package smgo

import (
	"flag"

	"github.com/globalsign/mgo"

	sdms "gitlab.sendo.vn/core/golang-sdk"
)

type MgoConfig struct {
	App sdms.Application

	// prefix to flag, used to difference multi instance
	FlagPrefix string

	// dbname to use in default config
	DefaultDBName string
}

// mgoServiceImpl Manage a session to mongodb
//
// Each mgoServiceImpl have a config to a server.
//
// If you need connect to many difference server, just use multiple instance of this service
type MgoService interface {
	sdms.Service
	GlobalSession() *mgo.Session
	Session(opts ...Option) *mgo.Session
	DB(opts ...Option) (db *mgo.Database, cleanup func())
	C(name string, opts ...Option) (c *mgo.Collection, cleanup func())
}

type mgoServiceImpl struct {
	cfg  MgoConfig
	log  sdms.Logger
	sess *mgo.Session

	// flags
	mgoUri string
}

func NewMgoService(config *MgoConfig) MgoService {
	if config.DefaultDBName == "" {
		config.DefaultDBName = "test"
	}

	return &mgoServiceImpl{
		cfg: *config,
		log: config.App.(sdms.SdkApplication).GetLog("mgo"),
	}
}

func (s *mgoServiceImpl) InitFlags() {
	flag.StringVar(&s.mgoUri, s.cfg.FlagPrefix+"mgo-uri", "mongodb://localhost/"+s.cfg.DefaultDBName, "MongoDB connection-string")
}

func (s *mgoServiceImpl) Configure() error {
	log := s.log

	var err error
	log.Info("Connect to mongodb at ", s.mgoUri, "...")
	s.sess, err = mgo.Dial(s.mgoUri)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (s *mgoServiceImpl) Cleanup() {
	if s.sess != nil {
		s.sess.Close()
	}
}

// return global session
//
// DO NOT CLOSE IT
func (m *mgoServiceImpl) GlobalSession() *mgo.Session {
	return m.sess
}

// return a new session
//
// Must close session after used
func (m *mgoServiceImpl) Session(opts ...Option) *mgo.Session {
	var s *mgo.Session

	for _, opt := range opts {
		switch opt.(type) {
		case OptionNewConn:
			if s != nil {
				s.Close()
			}
			s = m.sess.Copy()
		}
	}

	// default
	if s == nil {
		s = m.sess.Clone()
	}

	for _, opt := range opts {
		switch opt.(type) {
		case OptionSafe:
			s.SetSafe(opt.(OptionSafe).Value)
		}
	}

	return s
}

// DB return default db
// with a cleanup function (for close session)
//   db, clean := mgoServiceImpl.DB()
//   // db, clean := mgoServiceImpl.DB(smgo.OptionNewConn{})
//   defer clean()
//   // db.Session.Set... // config if need
//   count, err := db.C("test").Find(bson.M{}).Count()
func (m *mgoServiceImpl) DB(opts ...Option) (db *mgo.Database, cleanup func()) {
	s := m.Session(opts...)
	return s.DB(""), s.Close
}

// Return a mgo.Collection
//
// Work like DB() function
//   c, clean := mgoServiceImpl.C("test")
//   defer clean()
//   count, err := c.Find(bson.M{}).Count()
func (m *mgoServiceImpl) C(name string, opts ...Option) (c *mgo.Collection, cleanup func()) {
	db, cleanup := m.DB(opts...)
	return db.C(name), cleanup
}
