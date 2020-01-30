package broker

import (
	"context"
	"github.com/pkg/errors"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"gitlab.sendo.vn/protobuf/internal-apis-go/core/pubsub"
)

func (b *broker) processSubscribeStream(stream pubsub.SendoPubsub_SubscribeClient, msgChan chan<- *Message) error {
	var counter uint64
	mu := &sync.Mutex{}

	for {
		m, err := stream.Recv()
		if err != nil {
			return err
		}

		counter++

		t, err := types.TimestampFromProto(m.Created)
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
					Ack: &pubsub.MessageAck{
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
					Redeliver: &pubsub.MessageRedeliver{
						Tag:     tag,
						Message: s,
						Delay:   types.DurationProto(delay),
					},
				},
			})
		}

		msgChan <- &msg
	}
}

func (b *broker) Publish(pub *Publishing) (string, error) {
	if b.isDisabled() {
		return "", errors.New("Broker is not connected")
	}

	d := types.DurationProto(pub.Delay)

	req := pubsub.PublishReq{
		Event: pub.Event,
		Token: pub.Token,
		Data:  pub.Data,
		Delay: d,
	}

	resp, err := b.pubSubClient.Publish(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return resp.MessageId, nil
}

func (b *broker) Subscribe(ctx context.Context, opt *SubscribeOption) (<-chan *Message, error) {
	if b.isDisabled() {
		return nil, errors.New("Broker is not connected")
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := b.pubSubClient.Subscribe(ctx)
	if err != nil {
		return nil, err
	}

	optReq := &pubsub.SubscribeRequest{
		Request: &pubsub.SubscribeRequest_Option{
			Option: &pubsub.SubscribeOption{
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
		err = b.processSubscribeStream(stream, chMsgs)
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
