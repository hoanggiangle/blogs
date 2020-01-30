package storage

// Github: https://github.com/olivere/elastic

import (
	"flag"
	"github.com/olivere/elastic"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"log"
	"os"
)

type ESOpt struct {
	Prefix string
	Uri    string
}

type es struct {
	name   string
	client *elastic.Client
	logger logger.Logger
	*ESOpt
}

func NewES(name, flagPrefix string) *es {
	return &es{
		name:   name,
		logger: logger.GetCurrent().GetLogger(name),
		ESOpt: &ESOpt{
			Prefix: flagPrefix,
		},
	}
}

func (es *es) GetPrefix() string {
	return es.Prefix
}

func (es *es) isDisabled() bool {
	return es.Uri == ""
}

func (es *es) InitFlags() {
	prefix := es.Prefix
	if es.Prefix != "" {
		prefix += "-"
	}

	flag.StringVar(&es.Uri, prefix+"es-uri", "", "Elastic Search connection-string. Ex: http://localhost:9200")
}

func (es *es) Configure() error {
	if es.isDisabled() {
		return nil
	}

	es.logger.Info("Connecting to Elastic Search at ", es.Uri, "...")

	client, err := elastic.NewClient(
		elastic.SetURL(es.Uri),
		elastic.SetInfoLog(log.New(os.Stdout, "ELASTIC ", log.LstdFlags)),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))

	if err != nil {
		es.logger.Error("Cannot connect to Elastic Search. ", err.Error())
		return err
	}

	// Connect successfully, assign client
	es.client = client
	return nil
}

func (es *es) Name() string {
	return es.name
}

func (es *es) ElasticSearch() *elastic.Client {
	return es.client
}

func (es *es) Run() error {
	return es.Configure()
}

func (es *es) Stop() <-chan bool {
	// if es.client != nil {
	// 	es.client.Close()
	// }

	c := make(chan bool)
	go func() { c <- true }()
	return c
}
