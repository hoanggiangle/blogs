// RabbitMQ package
package samqp

import (
	"flag"
	"math/rand"
	"strings"
	"time"

	"github.com/streadway/amqp"

	sdms "gitlab.sendo.vn/core/golang-sdk"
)

type AmqpConfig struct {
	App sdms.Application

	// prefix to flag, used to difference multi instance
	FlagPrefix string
}

type AmqpService interface {
	sdms.Service
	// Open a new connection to RabbitMQ
	NewConn() (c *amqp.Connection, err error)
}

type amqpServiceImpl struct {
	cfg AmqpConfig

	rand *rand.Rand

	// flags
	_amqpUris string
	amqpUris  []string
}

func NewAmqp(cfg *AmqpConfig) AmqpService {
	rsrc := rand.NewSource(time.Now().Unix())
	return &amqpServiceImpl{
		cfg:  *cfg,
		rand: rand.New(rsrc),
	}
}

func (s *amqpServiceImpl) logger() sdms.Logger {
	return s.cfg.App.(sdms.SdkApplication).GetLog("amqp")
}

func (s *amqpServiceImpl) InitFlags() {
	flag.StringVar(&s._amqpUris, s.cfg.FlagPrefix+"amqp-uri", "amqp://localhost",
		"RabbitMQ connection-string. Many uris can be added (split by comma)")
}

func (s *amqpServiceImpl) Configure() error {
	s.amqpUris = strings.Split(s._amqpUris, ",")
	// test connection
	c, err := s.NewConn()
	if err == nil {
		defer c.Close()
	}
	return err
}

func (s *amqpServiceImpl) Cleanup() {
}

func (s *amqpServiceImpl) randomUris() []string {
	randUris := append([]string{}, s.amqpUris...)

	for i := range randUris {
		r := s.rand.Intn(len(randUris))
		if i != r {
			x := randUris[i]
			randUris[i] = randUris[r]
			randUris[r] = x
		}
	}
	return randUris
}

func (s *amqpServiceImpl) NewConn() (c *amqp.Connection, err error) {
	log := s.logger()
	for _, u := range s.randomUris() {
		c, err = amqp.Dial(u)
		if err == nil {
			return
		}
	}
	log.Errorf("Open connection to %s: %s", s.amqpUris, err.Error())
	return
}
