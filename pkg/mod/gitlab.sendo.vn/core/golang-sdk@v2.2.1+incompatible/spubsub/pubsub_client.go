package spubsub

import (
	"context"
	"sync"
	"time"

	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"

	"gitlab.sendo.vn/protobuf/internal-apis-go/core/pubsub"
)

type PubsubClient interface {
	Publish(*Publishing) (string, error)
	Subscribe(context.Context, *SubscribeOption) (<-chan *Message, error)
}

type Publishing struct {
	Event string
	// event token
	Token string
	Data  []byte
	Delay time.Duration
}

type SubscribeOption struct {
	Event string
	// subscriber token
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
	Created time.Time

	ack       func(string)
	redeliver func(string, time.Duration)

	hadAck bool
}

func (m *Message) Ack(logmsg string) {
	m.ack(logmsg)
	m.hadAck = true
}

func (m *Message) Redeliver(logmsg string, delay time.Duration) {
	m.redeliver(logmsg, delay)
	m.hadAck = true
}

func NewPubsubClient(conn *grpc.ClientConn) PubsubClient {
	return &pubsubClientImpl{
		cli: pubsub.NewSendoPubsubClient(conn),
	}
}

type pubsubClientImpl struct {
	cli pubsub.SendoPubsubClient
}

func (p *pubsubClientImpl) Publish(pub *Publishing) (string, error) {
	d := ptypes.DurationProto(pub.Delay)

	req := pubsub.PublishReq{
		Event: pub.Event,
		Token: pub.Token,
		Data:  pub.Data,
		Delay: d,
	}

	resp, err := p.cli.Publish(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return resp.MessageId, nil
}

func (p *pubsubClientImpl) processSubscribeStream(stream pubsub.SendoPubsub_SubscribeClient, chMsgs chan<- *Message) error {
	var counter uint64
	mu := &sync.Mutex{}

	for {
		m, err := stream.Recv()
		if err != nil {
			return err
		}

		counter++

		t, err := ptypes.TimestampFromProto(m.Created)
		if err != nil {
			return err
		}
		msg := Message{
			Id:             m.Id,
			Data:           m.Data,
			DeliveredCount: m.DeliveredCount,
			Created:        t,
		}
		tag := counter
		msg.ack = func(s string) {
			mu.Lock()
			defer mu.Unlock()

			stream.Send(&pubsub.SubscribeRequest{
				Request: &pubsub.SubscribeRequest_Ack{
					&pubsub.MessageAck{
						Tag:     tag,
						Message: s,
					},
				},
			})
		}
		msg.redeliver = func(s string, delay time.Duration) {
			mu.Lock()
			defer mu.Unlock()

			stream.Send(&pubsub.SubscribeRequest{
				Request: &pubsub.SubscribeRequest_Redeliver{
					&pubsub.MessageRedeliver{
						Tag:     tag,
						Message: s,
						Delay:   ptypes.DurationProto(delay),
					},
				},
			})
		}
		chMsgs <- &msg
	}
}

func (p *pubsubClientImpl) Subscribe(ctx context.Context, opt *SubscribeOption) (<-chan *Message, error) {
	ctx, cancel := context.WithCancel(ctx)

	stream, err := p.cli.Subscribe(ctx)
	if err != nil {
		return nil, err
	}

	optReq := &pubsub.SubscribeRequest{
		Request: &pubsub.SubscribeRequest_Option{
			&pubsub.SubscribeOption{
				Event:         opt.Event,
				Token:         opt.Token,
				MaxConcurrent: opt.MaxConcurrent,
			},
		},
	}
	if err = stream.Send(optReq); err != nil {
		return nil, err
	}

	chMsgs := make(chan *Message)

	// wait a little to check if is there any error
	wait := make(chan struct{}, 1)
	go func() {
		err = p.processSubscribeStream(stream, chMsgs)
		wait <- struct{}{}
		cancel()
		close(chMsgs)
	}()

	t := time.NewTimer(time.Second)
	defer t.Stop()
	select {
	case <-wait:
	case <-t.C:
	}

	if err != nil {
		return nil, err
	}

	return chMsgs, nil
}
