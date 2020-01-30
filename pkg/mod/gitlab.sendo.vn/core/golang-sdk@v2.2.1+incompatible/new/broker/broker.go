package broker

import (
	"context"
	"errors"
	"time"
)

type Broker interface {
	Publish(*Publishing) (string, error)
	Subscribe(context.Context, *SubscribeOption) (<-chan *Message, error)

	IsConnected() bool
	GetPubToken(e string) string
	GetSubToken(e string) string
}

type notSetBroker struct{}

func NewNotSetBroker() *notSetBroker {
	return &notSetBroker{}
}

func (nsb *notSetBroker) Publish(*Publishing) (string, error) {
	return "", errors.New("not set broker")
}

func (nsb *notSetBroker) Subscribe(context.Context, *SubscribeOption) (<-chan *Message, error) {
	return nil, errors.New("not set broker")
}

func (nsb *notSetBroker) IsConnected() bool {
	return false
}

func (nsb *notSetBroker) GetPubToken(e string) string {
	return ""
}

func (nsb *notSetBroker) GetSubToken(e string) string {
	return ""
}

type Publishing struct {
	Event string
	Token string
	Data  []byte
	Delay time.Duration
}

type SubscribeOption struct {
	Event string
	Token string
	// default: max server-side allowed number
	MaxConcurrent int32
}

type Message struct {
	Id   string
	Data []byte
	// number of time this message has redelivered (but not acknowledge yet)
	DeliveredCount int32
	// message created time
	Created   time.Time
	ack       func(string)
	redeliver func(string, time.Duration)
	hadAck    bool
}

func (m *Message) Ack(logMsg string) {
	m.ack(logMsg)
	m.hadAck = true
}

func (m *Message) Redeliver(logMsg string, delay time.Duration) {
	m.redeliver(logMsg, delay)
	m.hadAck = true
}
