package spubsub

import (
	"flag"
	"sync"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/sgrpc"
)

type PubsubClientServiceInterface interface {
	GetPubsubClient() PubsubClient
	GetPubToken(e string) string
	GetSubToken(e string) string
}

type PubsubClientService interface {
	sdms.Service
	PubsubClientServiceInterface
}

type ClientConfig struct {
	App sdms.Application
	// which events want to publish
	PublishEvents []string
	// which events want to subscribe
	SubscribeEvents []string
}

func NewPubsubClientService(cfg *ClientConfig) PubsubClientService {
	return &pubsubClientServiceImpl{
		GrpcClientService: sgrpc.NewGrpcClientService(cfg.App),
		cfg:               *cfg,
		pubTokenMap:       make(map[string]*string),
		subTokenMap:       make(map[string]*string),
	}
}

type pubsubClientServiceImpl struct {
	sgrpc.GrpcClientService
	cfg ClientConfig

	pubTokenMap map[string]*string
	subTokenMap map[string]*string

	initOnce sync.Once

	cli PubsubClient

	pubsubAddr string
}

func (p *pubsubClientServiceImpl) logger() sdms.Logger {
	return p.cfg.App.(sdms.SdkApplication).GetLog("pubsub")
}

func (p *pubsubClientServiceImpl) InitFlags() {
	p.GrpcClientService.InitFlags()
	p.GrpcClientService.EndpointFlag(&p.pubsubAddr, "pubsub", "")

	log := p.logger()

	for _, e := range p.cfg.PublishEvents {
		if _, found := p.pubTokenMap[e]; found {
			log.Panicf("ClientConfig.PublishEvents must be unique (duplicated %s)", e)
		}
		p.pubTokenMap[e] = flag.String("pubsub-pub-token-"+e, "", "publish token for event "+e)
	}
	for _, e := range p.cfg.SubscribeEvents {
		if _, found := p.subTokenMap[e]; found {
			log.Panicf("ClientConfig.SubscribeEvents must be unique (duplicated %s)", e)
		}
		p.subTokenMap[e] = flag.String("pubsub-sub-token-"+e, "", "subscribe token for event "+e)
	}
}

func (p *pubsubClientServiceImpl) initClient() {
	conn, err := p.GetConnection(p.pubsubAddr)
	if err != nil {
		p.logger().Panic(err)
	}
	p.cli = NewPubsubClient(conn)
}

func (p *pubsubClientServiceImpl) GetPubsubClient() PubsubClient {
	p.initOnce.Do(p.initClient)
	return p.cli
}

func (p *pubsubClientServiceImpl) GetPubToken(e string) string {
	t, found := p.pubTokenMap[e]
	if !found {
		p.logger().Panicf("ClientConfig.PublishEvents do not contain event %s", e)
	}
	return *t
}

func (p *pubsubClientServiceImpl) GetSubToken(e string) string {
	t, found := p.subTokenMap[e]
	if !found {
		p.logger().Panicf("ClientConfig.SubscribeEvents do not contain event %s", e)
	}
	return *t
}
