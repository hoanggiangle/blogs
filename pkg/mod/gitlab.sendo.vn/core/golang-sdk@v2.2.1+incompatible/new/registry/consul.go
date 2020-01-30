package registry

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	consulAPI "github.com/hashicorp/consul/api"
	"gitlab.sendo.vn/core/golang-sdk/new/logger"
	"gitlab.sendo.vn/core/golang-sdk/new/util"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	SERVICE_CONFIG_PREFIX = "service-configs"
	fetchKVInterVal       = time.Second * 15
)

type ServiceConfig struct {
	Name     string `json:"-"`
	Prefix   string `json:"prefix,omitempty"`
	Regex    string `json:"regex,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type serviceStatus struct {
	svc    *consulAPI.AgentServiceRegistration
	status bool
}

// A Consul that implement a simple TCP consul service
type consul struct {
	// Main service name to generate config path on Consul
	mainServiceName string
	logger          logger.Logger
	client          *consulAPI.Client
	currentKVs      []byte
	services        map[string]*serviceStatus
	svcConfigs      map[string]*ServiceConfig
	consulUri       string
	serviceIp       string
	servicePrefix   string
	tags            string
	checkIp         string
	mu              *sync.Mutex
	stopFunc        func()
	syncChan        chan bool
}

func NewConsul(mainServiceName string) *consul {
	return &consul{
		mainServiceName: mainServiceName,
		logger:          logger.GetCurrent().GetLogger("consul"),
		services:        make(map[string]*serviceStatus),
		svcConfigs:      make(map[string]*ServiceConfig),
		mu:              &sync.Mutex{},
		syncChan:        make(chan bool),
	}
}

func (s *consul) Name() string {
	return "Consul"
}

func (s *consul) InitFlags() {
	flag.StringVar(&s.consulUri, "ssd-uri", "0", "consul address. Set to 0 or false to disabled")
	flag.StringVar(&s.serviceIp, "ssd-service-ip", "auto", "ip to register, empty to use consul setting")
	flag.StringVar(&s.servicePrefix, "ssd-service-prefix", "", "service id prefix (default: <hostname>)")
	flag.StringVar(&s.tags, "ssd-service-tags", "", "service tags (example: pilot,pc2)")
}

func (s *consul) Configure() error {
	if s.isDisabled() {
		return nil
	}
	var err error

	u, err := url.Parse(s.consulUri)
	if err != nil {
		s.logger.Error(err)
		return err
	}

	if u.Host == "" {
		return errors.New("invalid consul uri")
	}

	config := consulAPI.DefaultConfig()
	config.Scheme = u.Scheme
	config.Address = u.Host

	s.logger.Infof("Connecting to Consul at %s", s.consulUri)
	s.mu.Lock()
	s.client, err = consulAPI.NewClient(config)
	s.mu.Unlock()
	if err != nil {
		s.logger.Error(err)
		return err
	}

	if s.servicePrefix == "" {
		s.servicePrefix, _ = os.Hostname()
	}

	s.setCheckIP(u.Host)

	if s.serviceIp == "auto" {
		ip := net.ParseIP(s.GetCheckIP())
		if ip != nil && !ip.IsLoopback() {
			s.serviceIp = ip.String()
		} else {
			s.serviceIp = ""
		}
	}

	return nil
}

func (s *consul) isDisabled() bool {
	return s.consulUri == "" || s.consulUri == "false" || s.consulUri == "0"
}

func (s *consul) register(inf *serviceStatus) {
	agent := s.client.Agent()
	asr := inf.svc
	s.logger.Infof("registering service %s %s (%s:%d)", asr.ID, asr.Name, asr.Address, asr.Port)
	err := agent.ServiceRegister(asr)
	if err != nil {
		s.logger.Error("register service %s: ", asr.ID, err)
		inf.status = false
	} else {
		inf.status = true
	}
}

func (s *consul) Run() error {
	if s.isDisabled() {
		return nil
	}

	if err := s.Configure(); err != nil {
		return err
	}

	s.startTimerGetKVs()
	return nil
}

func (s *consul) startTimerGetKVs() {
	if s.isDisabled() || s.client == nil {
		return
	}

	// First load
	s.fetchKV()

	timer := time.NewTicker(fetchKVInterVal)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				s.fetchKV()
			}
		}
	}()

	s.stopFunc = cancel
}

func (s *consul) fetchKV() {
	s.mu.Lock()
	defer s.mu.Unlock()

	firstTimeLoading := s.currentKVs == nil

	clientKV := s.client.KV()
	// kvs, _, err := clientKV.List("", &consulAPI.QueryOptions{
	// 	WaitTime: time.Second * 5,
	// })
	// log.Print(kvs, err)

	key := fmt.Sprintf("%s/%s/%s", SERVICE_CONFIG_PREFIX, s.mainServiceName, "default")
	keyPair, _, err := clientKV.Get(key, &consulAPI.QueryOptions{
		WaitTime: time.Second * 5,
	})

	if err != nil || keyPair == nil {
		return
	}

	if !firstTimeLoading {
		if bytes.Compare(s.currentKVs, keyPair.Value) == 0 {
			s.currentKVs = keyPair.Value
			go s.notifyConfigChanged()
		}
	}

	s.currentKVs = keyPair.Value
}

func (s *consul) notifyConfigChanged() {
	s.syncChan <- true
}

func (s *consul) Stop() <-chan bool {
	s.mu.Lock()
	if s.stopFunc != nil {
		s.stopFunc()
	}
	s.mu.Unlock()

	// make sure all services is removed
	s.RemoveAll()

	c := make(chan bool)
	go func() { c <- true }()
	return c
}

func (s *consul) Cleanup() {}

// Add ID, address if not exists, append default tags
func (s *consul) normalizeASR(asr *consulAPI.AgentServiceRegistration) {
	if asr.ID == "" {
		asr.ID = fmt.Sprintf("%s_%s_%d", s.servicePrefix, asr.Name, asr.Port)
	}

	if asr.Address == "" && s.serviceIp != "" {
		asr.Address = s.serviceIp
	}

	tags := strings.Split(s.tags, ",")
	for i := 0; i < len(tags); i++ {
		tag := strings.TrimSpace(tags[i])
		if tag == "" {
			continue
		}
		asr.Tags = append(asr.Tags, tag)
	}

	if inf := util.GetBuildInfo(); inf != nil {
		if inf.Date != "" {
			asr.Tags = append(asr.Tags, inf.GetDate().Format("20060102-150405MST"))
		}
		if inf.GetBranch() != "" && len(inf.Revision) >= 40 {
			asr.Tags = append(asr.Tags, inf.GetBranch()+"_"+inf.Revision[:7])
		}
	}
}

func (s *consul) RegisterService(asr *consulAPI.AgentServiceRegistration) string {
	if s.isDisabled() {
		return ""
	}

	s.normalizeASR(asr)
	inf := serviceStatus{svc: asr, status: false}
	s.register(&inf)

	s.mu.Lock()
	s.services[asr.ID] = &inf
	s.mu.Unlock()

	return asr.ID
}

func (s *consul) DeregisterService(serviceID string) {
	if serviceID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, serviceID)

	s.logger.Infof("deregister service %s", serviceID)
	err := s.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		s.logger.Warn(err)
	}
}

func (s *consul) RemoveAll() {
	agent := s.client.Agent()

	s.mu.Lock()
	defer s.mu.Unlock()
	for k, inf := range s.services {
		s.logger.Debugf("deregister service %s", inf.svc.ID)
		agent.ServiceDeregister(inf.svc.ID)
		delete(s.services, k)
	}
}

func (s *consul) setCheckIP(remoteHost string) {
	conn, err := net.Dial("tcp", remoteHost)
	if err != nil {
		s.logger.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	s.checkIp = localAddr.IP.String()
	if strings.Contains(s.checkIp, ":") {
		s.checkIp = "[" + s.checkIp + "]"
	}
}

func (s *consul) GetCheckIP() string {
	return s.checkIp
}

func (s *consul) GetKVs() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentKVs
}

func (s *consul) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.client != nil
}

func (s *consul) SyncChan() <-chan bool {
	return s.syncChan
}
