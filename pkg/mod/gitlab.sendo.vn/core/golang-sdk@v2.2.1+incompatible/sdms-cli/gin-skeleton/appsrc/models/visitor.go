package models

{{- if .Resources.mgo }}

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"{{ .ImportPath }}/appsrc/resprovider"
)

type Visitor struct {
	Addr string    `json:"addr"`
	UA   string    `json:"ua"`
	Date time.Time `json:"date"`
}

type VisitorModel struct {
}

// get collection of this model
func (m *VisitorModel) c() (coll *mgo.Collection, close func()) {
	return resprovider.GetInstance().C("visitor")
}

func (m *VisitorModel) logger(prefix string) resprovider.Logger {
	return resprovider.GetInstance().Logger(prefix)
}

func (m *VisitorModel) Insert(v *Visitor) error {
	c, close := m.c()
	defer close()

	log := m.logger("visitor")
	log.With("ip", v.Addr).Debug("new visitor")
	return c.Insert(v)
}

func (m *VisitorModel) Last10(visitors *[]Visitor) error {
	c, close := m.c()
	defer close()

	return c.Find(bson.M{}).Sort("-_id").Limit(10).All(visitors)
}
{{- else }}

// this file used with mongo
{{- end }}
