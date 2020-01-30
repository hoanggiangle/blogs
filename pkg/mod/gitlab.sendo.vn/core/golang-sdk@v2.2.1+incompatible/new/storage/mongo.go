package storage

import (
	"flag"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"

	"github.com/globalsign/mgo"
)

type MongoDBOpt struct {
	MgoUri string
	Prefix string
}

type mongoDB struct {
	name      string
	logger    logger.Logger
	session   *mgo.Session
	isRunning bool
	*MongoDBOpt
}

func NewMongoDB(name, prefix string) *mongoDB {
	return &mongoDB{
		MongoDBOpt: &MongoDBOpt{
			Prefix: prefix,
		},
		name:      name,
		logger:    logger.GetCurrent().GetLogger(name),
		isRunning: false,
	}
}

func (mgDB *mongoDB) GetPrefix() string {
	return mgDB.Prefix
}

func (mgDB *mongoDB) Name() string {
	return mgDB.name
}

func (mgDB *mongoDB) InitFlags() {
	prefix := mgDB.Prefix
	if mgDB.Prefix != "" {
		prefix += "-"
	}

	flag.StringVar(&mgDB.MgoUri, prefix+"mgo-uri", "", "MongoDB connection-string. Ex: mongodb://...")
}

func (mgDB *mongoDB) isDisabled() bool {
	return mgDB.MgoUri == ""
}

func (mgDB *mongoDB) Configure() error {
	if mgDB.isDisabled() || mgDB.isRunning {
		return nil
	}

	mgDB.logger.Info("Connect to Mongodb at ", mgDB.MgoUri, " ...")

	var err error
	mgDB.session, err = mgo.Dial(mgDB.MgoUri)
	if err != nil {
		mgDB.logger.Error("Error connect to mongodb at ", mgDB.MgoUri, ". ", err.Error())
		return err
	}
	mgDB.isRunning = true

	return nil
}

func (mgDB *mongoDB) Cleanup() {
	if mgDB.isDisabled() {
		return
	}

	if mgDB.session != nil {
		mgDB.session.Close()
	}
}

func (mgDB *mongoDB) Run() error {
	return mgDB.Configure()
}

func (mgDB *mongoDB) Stop() <-chan bool {
	if mgDB.session != nil {
		mgDB.session.Close()
	}
	mgDB.isRunning = false

	c := make(chan bool)
	go func() { c <- true }()
	return c
}

func (mgDB *mongoDB) Mgo() *mgo.Session {
	return mgDB.session
}
