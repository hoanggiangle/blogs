package broker

import (
	"flag"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"gitlab.sendo.vn/protobuf/internal-apis-go/core/pubsub"
	"google.golang.org/grpc"
	"sync"
)

type Config struct {
	// which events broker would publish
	PublishEvents []string
	// which events broker would subscribe
	SubscribeEvents []string
}

func New(cfg *Config) *broker {
	return &broker{
		logger:      logger.GetCurrent().GetLogger("broker"),
		cfg:         *cfg,
		pubTokenMap: make(map[string]*string),
		subTokenMap: make(map[string]*string),
	}
}

type broker struct {
	pubSubClient pubsub.SendoPubsubClient
	logger       logger.Logger
	cfg          Config
	initOnce     sync.Once
	clientConn   *grpc.ClientConn
	pubsubAddr   string
	pubTokenMap  map[string]*string
	subTokenMap  map[string]*string
}

func (b *broker) Name() string {
	return "broker"
}

func (b *broker) InitFlags() {
	flag.StringVar(&b.pubsubAddr, "grpc-endpoint-pubsub", "", "pubsub server endpoint")

	log := b.logger

	for _, e := range b.cfg.PublishEvents {
		if _, found := b.pubTokenMap[e]; found {
			log.Panicf("Config.PublishEvents must be unique (duplicated %s)", e)
		}
		b.pubTokenMap[e] = flag.String("pubsub-pub-token-"+e, "", "publish token for event "+e)
	}
	for _, e := range b.cfg.SubscribeEvents {
		if _, found := b.subTokenMap[e]; found {
			log.Panicf("Config.SubscribeEvents must be unique (duplicated %s)", e)
		}
		b.subTokenMap[e] = flag.String("pubsub-sub-token-"+e, "", "subscribe token for event "+e)
	}
}

func (b *broker) Configure() error {
	if b.isDisabled() {
		return nil
	}
	return nil
}

func (b *broker) Run() error {
	if b.isDisabled() {
		return nil
	}

	b.logger.Info("Connecting to pubsub grpc at ", b.pubsubAddr)
	cc, err := grpc.Dial(b.pubsubAddr, grpc.WithInsecure())
	b.clientConn = cc
	b.pubSubClient = pubsub.NewSendoPubsubClient(cc)
	return err
}

func (b *broker) Stop() <-chan bool {
	closeChan := make(chan bool)

	if b.clientConn != nil {
		b.clientConn.Close()
	}

	go func() { closeChan <- true }()
	return closeChan
}

func (b *broker) isDisabled() bool {
	return b.pubsubAddr == ""
}

func (b *broker) GetPubToken(e string) string {
	t, found := b.pubTokenMap[e]
	if !found {
		b.logger.Panicf("Config.PublishEvents do not contain event %s", e)
	}
	return *t
}

func (b *broker) GetSubToken(e string) string {
	t, found := b.subTokenMap[e]
	if !found {
		b.logger.Panicf("Config.SubscribeEvents do not contain event %s", e)
	}
	return *t
}

func (b *broker) IsConnected() bool {
	return b.clientConn != nil
}
