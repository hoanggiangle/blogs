// +build pubsub

package tests

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/slog"
	"gitlab.sendo.vn/core/golang-sdk/spubsub"
)

const (
	PUBSUB_SERVER = "localhost:10000"
	PUBSUB_EVENT  = "abc"
	PUBSUB_TOKEN  = "Z40zpnnoEOKlvV9K"
)

func pubsubPublishWorker(t *testing.T, cli spubsub.PubsubClient, n int) {
	for i := 0; i < n; i++ {
		if _, err := cli.Publish(&spubsub.Publishing{
			Event: PUBSUB_EVENT,
			Token: "1",
		}); err != nil {
			t.Error(err)
		}
		// time.Sleep(time.Second)
	}
}

func pubsubConsumeWorker(t *testing.T, cli spubsub.PubsubClient, n int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	msgs, err := cli.Subscribe(ctx, &spubsub.SubscribeOption{
		Event: PUBSUB_EVENT,
		Token: PUBSUB_TOKEN,
	})
	if err != nil {
		t.Fatal(err)
	}

	for m := range msgs {
		go func(m *spubsub.Message) {
			// time.Sleep(time.Second)
			// log.Println(m)
			m.Ack("")
		}(m)

		n--
		if n == 0 {
			time.Sleep(time.Millisecond * 300)
			break
		}
	}
}

func TestPubsubClient(t *testing.T) {
	c, err := grpc.Dial(PUBSUB_SERVER, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	cli := spubsub.NewPubsubClient(c)

	go pubsubPublishWorker(t, cli, 5)
	pubsubConsumeWorker(t, cli, 5)
}

func TestPubsubClientService(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{},
		LogConfig: &slog.LoggerConfig{
			DefaultLevel: "warn",
		},
		UseNewFlagSet: true,
	})

	pss := spubsub.NewPubsubClientService(&spubsub.ClientConfig{
		App:             app,
		PublishEvents:   []string{"event.a", "event.a-1"},
		SubscribeEvents: []string{"event.a", "event.a-1"},
	})
	app.RegService(pss)

	nullS := &nullService{}
	app.RegMainService(nullS)

	defer executeApp(app)()

	nullS.WaitRun()

	cli := pss.GetPubsubClient()

	go pubsubPublishWorker(t, cli, 5)
	pubsubConsumeWorker(t, cli, 5)
}
