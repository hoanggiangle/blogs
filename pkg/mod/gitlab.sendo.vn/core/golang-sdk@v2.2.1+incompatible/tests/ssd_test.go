package tests

import (
	"log"
	"sync"
	"testing"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/ssd"
)

type testSsdService struct {
	nullService
	sd      ssd.ConsulSD
	muReady *sync.Mutex
}

func newTestSdService(sd ssd.ConsulSD) *testSsdService {
	mu := &sync.Mutex{}
	mu.Lock()
	return &testSsdService{
		sd:      sd,
		muReady: mu,
	}
}

func (t *testSsdService) Run() error {
	asr := consulapi.AgentServiceRegistration{
		Name: "test-ssd",
		Port: 80,
	}
	t.sd.Add(&asr)

	t.muReady.Unlock()

	time.Sleep(time.Minute)

	return nil
}

func (t *testSsdService) Port() int {
	t.muReady.Lock()
	defer t.muReady.Unlock()
	return 80
}

func TestSsd(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args:          []string{},
		UseNewFlagSet: true,
	})

	sd := ssd.NewConsul(app)
	app.RegService(sd)

	testSvc := newTestSdService(sd)
	app.RegMainService(testSvc)

	defer executeApp(app)()

	testSvc.Port()

	cli := sd.GetClient()

	svcs, _, err := cli.Health().Service("test-ssd", "", false, &consulapi.QueryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(svcs) == 0 {
		t.Fatal("No service is registered")
	}

	svc := svcs[0].Service

	if svc.Port != testSvc.Port() || svc.Service != "test-ssd" {
		t.Fatal("Wrong register info")
	}

	log.Printf(`service "%s" running at %s:%d`, svc.Service, svc.Address, svc.Port)
}
