package spubsub

import (
	"context"
	"sync"

	sdms "gitlab.sendo.vn/core/golang-sdk"
)

// easy interface for implement subscribers
type SubscriberService interface {
	sdms.RunnableService
}

type SubscriberConfig struct {
	Event string
	// default: max server-side allowed number
	MaxConcurrent int32

	ClientService PubsubClientServiceInterface
	ProcessFunc   func(*Message)
	// should wait for all ProcessFunc exit before stop app
	WaitForProcessFunc bool
}

func NewSubscriberService(cfg *SubscriberConfig) SubscriberService {
	if cfg.Event == "" {
		panic("SubscriberConfig.Event is not set")
	}
	if cfg.ClientService == nil {
		panic("SubscriberConfig.ClientService is not set")
	}
	if cfg.ProcessFunc == nil {
		panic("SubscriberConfig.ProcessFunc is not set")
	}
	return &subscriberService{
		cfg: *cfg,
	}
}

type subscriberService struct {
	cfg SubscriberConfig

	cancelSub context.CancelFunc

	wg sync.WaitGroup
}

func (s *subscriberService) InitFlags() {}

func (s *subscriberService) Configure() error { return nil }

func (s *subscriberService) Run() error {
	cli := s.cfg.ClientService.GetPubsubClient()

	var ctx context.Context
	ctx, s.cancelSub = context.WithCancel(context.Background())
	defer s.cancelSub()

	opt := SubscribeOption{
		Event:         s.cfg.Event,
		Token:         s.cfg.ClientService.GetSubToken(s.cfg.Event),
		MaxConcurrent: s.cfg.MaxConcurrent,
	}
	msgs, err := cli.Subscribe(ctx, &opt)
	if err != nil {
		panic(err)
	}

	for m := range msgs {
		go s.processMessage(m)
	}

	return nil
}

func (s *subscriberService) processMessage(m *Message) {
	s.wg.Add(1)
	defer s.wg.Done()

	s.cfg.ProcessFunc(m)

	if !m.hadAck {
		panic("ProcessFunc must call Message.Ack or Message.Redeliver!")
	}
}

func (s *subscriberService) Stop() {
	if s.cancelSub != nil {
		s.cancelSub()
	}
	if s.cfg.WaitForProcessFunc {
		s.wg.Wait()
	}
}

func (s *subscriberService) Cleanup() {}
