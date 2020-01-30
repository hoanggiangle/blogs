package dataLayer

import "github.com/globalsign/mgo"

type MgoDataLayer interface {
	Insert(something interface{}) error
	Update(condition, something interface{}) error
	Count(condition interface{}) (int, error)
	Find(condition, result interface{}, offset, limit *int) error
	Delete(something interface{}) error
}

type closeFunc func()

type mgoDataLayer struct {
	s              *mgo.Session
	collectionName string
}

func NewMgoDataLayer(s *mgo.Session, collectionName string) *mgoDataLayer {
	return &mgoDataLayer{s: s, collectionName: collectionName}
}

func (dl *mgoDataLayer) getConn() (*mgo.Collection, closeFunc) {
	s := dl.s.New()
	return s.DB("").C(dl.collectionName), s.Close
}

func (dl *mgoDataLayer) Update(condition, something interface{}) error {
	s, close := dl.getConn()
	defer close()
	return s.Update(condition, something)
}

func (dl *mgoDataLayer) Count(condition interface{}) (int, error) {
	s, close := dl.getConn()
	defer close()
	return s.Find(condition).Count()
}

func (dl *mgoDataLayer) Find(condition, result interface{}, offset, limit *int) error {
	s, close := dl.getConn()
	defer close()

	query := s.Find(condition)
	if offset != nil {
		query = query.Skip(*offset)
	}
	if limit != nil {
		query = query.Limit(*limit)
	}

	return query.All(result)
}

func (dl *mgoDataLayer) Insert(something interface{}) error {
	s, close := dl.getConn()
	defer close()

	return s.Insert(something)
}

func (dl *mgoDataLayer) Delete(condition interface{}) error {
	s, close := dl.getConn()
	defer close()
	return s.Remove(condition)
}
