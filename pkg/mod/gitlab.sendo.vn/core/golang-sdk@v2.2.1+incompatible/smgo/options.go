package smgo

import (
	"github.com/globalsign/mgo"
)

type Option interface {
}

// make new connection (use session.Copy)
type OptionNewConn struct {
}

type OptionSafe struct {
	Value *mgo.Safe
}
