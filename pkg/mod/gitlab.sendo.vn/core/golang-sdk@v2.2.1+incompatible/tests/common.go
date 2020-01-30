package tests

import (
	"fmt"
	"sync/atomic"
	"time"

	sdms "gitlab.sendo.vn/core/golang-sdk"
)

type nullService struct {
	RunFunc func() error
	inRun   int32
}

func (n *nullService) InitFlags()       {}
func (n *nullService) Configure() error { return nil }
func (n *nullService) Stop()            {}
func (n *nullService) Run() error {
	atomic.StoreInt32(&n.inRun, 1)
	if n.RunFunc != nil {
		return n.RunFunc()
	} else {
		for {
			time.Sleep(time.Second)
		}
	}
	return nil
}
func (n *nullService) Cleanup() {}

func (n *nullService) WaitRun() {
	for atomic.LoadInt32(&n.inRun) == 0 {
		time.Sleep(time.Millisecond)
	}
}

type producerService struct {
	nullService
	log sdms.Logger
	ch  chan string
}

func newproducerService(log sdms.Logger, ch chan string) sdms.RunnableService {
	return &producerService{
		log: log,
		ch:  ch,
	}
}

func (c *producerService) Run() error {
	for i := 0; i < 5; i++ {
		c.ch <- fmt.Sprintf("message %d", i)
	}
	close(c.ch)
	return nil
}

type consumeService struct {
	nullService
	log      sdms.Logger
	ch       chan string
	chQuit   chan bool
	chQuitOk chan bool
}

func newConsumeService(log sdms.Logger, ch chan string) sdms.RunnableService {
	return &consumeService{
		log:      log,
		ch:       ch,
		chQuit:   make(chan bool, 1),
		chQuitOk: make(chan bool, 1),
	}
}

func (c *consumeService) Run() error {
	for {
		select {
		case s, ok := <-c.ch:
			if !ok {
				goto SHUTDOWN
			}
			c.log.Info(s)
		case <-c.chQuit:
			goto SHUTDOWN
		}
	}

SHUTDOWN:
	c.chQuitOk <- true
	return nil
}

func (c *consumeService) Stop() {
	c.chQuit <- true
	<-c.chQuitOk
}

func executeApp(app sdms.Application) func() {
	go app.Run()
	return func() {
		<-app.Shutdown()
	}
}
