package ssd

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/util"
)

const (
	SERVICE_CONFIG_PREFIX = "service-configs"
)

type ServiceConfig struct {
	// service name
	Name string `json:"-"`
	// prefix to match request
	Prefix string `json:"prefix,omitempty"`
	// regex to match request
	Regex string `json:"regex,omitempty"`

	Protocol string `json:"protocol,omitempty"`
}

type serviceStatus struct {
	svc    *consulapi.AgentServiceRegistration
	status bool
}

// A SD that implement a simple TCP consul service
type consulServiceImpl struct {
	app sdms.Application
	log sdms.Logger

	client *consulapi.Client

	services   map[string]*serviceStatus
	svcConfigs map[string]*ServiceConfig

	stopFunc func()

	consulUri     string
	serviceIp     string
	servicePrefix string
	tags          string

	checkIp string

	mu *sync.Mutex
}

func NewConsul(app sdms.Application) ConsulService {
	return &consulServiceImpl{
		app:        app,
		log:        app.(sdms.SdkApplication).GetLog("ssd"),
		services:   make(map[string]*serviceStatus),
		svcConfigs: make(map[string]*ServiceConfig),
		mu:         &sync.Mutex{},
	}
}

func (s *consulServiceImpl) InitFlags() {
	flag.StringVar(&s.consulUri, "ssd-uri", "0", "consul address. Set to 0 or false to disabled")
	flag.StringVar(&s.serviceIp, "ssd-service-ip", "auto", "ip to register, empty to use consul setting")
	flag.StringVar(&s.servicePrefix, "ssd-service-prefix", "", "service id prefix (default: <hostname>)")
	flag.StringVar(&s.tags, "ssd-service-tags", "", "service tags (example: pilot,pc2)")
}

func (s *consulServiceImpl) Configure() error {
	var err error
	if s.isDisabled() {
		return nil
	}

	u, err := url.Parse(s.consulUri)
	if err != nil {
		s.log.Error(err)
		return err
	}

	if u.Host == "" {
		return errors.New("Invalid ssd uri")
	}

	config := consulapi.DefaultConfig()
	config.Scheme = u.Scheme
	config.Address = u.Host

	s.mu.Lock()
	s.client, err = consulapi.NewClient(config)
	s.mu.Unlock()
	if err != nil {
		s.log.Error(err)
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

func (s *consulServiceImpl) isDisabled() bool {
	return s.consulUri == "" || s.consulUri == "false" || s.consulUri == "0"
}

func (s *consulServiceImpl) register(inf *serviceStatus) {
	agent := s.client.Agent()
	asr := inf.svc
	s.log.Debugf("register service %s %s (%s:%d)", asr.ID, asr.Name, asr.Address, asr.Port)
	err := agent.ServiceRegister(asr)
	if err != nil {
		s.log.Error("register service: ", err)
		inf.status = false
	} else {
		inf.status = true
	}
}

// Background register routine
//
// Service failed to register will be reRegister in 30s.
//
// A success registered service also be reRegister in 15m.
func (s *consulServiceImpl) Run() error {
	if s.isDisabled() {
		return nil
	}

	s.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	s.stopFunc = cancel
	s.mu.Unlock()

	tRegister := time.NewTicker(time.Second * 10)
	defer tRegister.Stop()
	tReregister := time.NewTicker(time.Minute * 2)
	defer tReregister.Stop()
	tReconfig := time.NewTicker(time.Minute * 15)
	defer tReconfig.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-tRegister.C: // register unregistered service
			s.mu.Lock()
			for _, inf := range s.services {
				if inf.status == true {
					continue
				}
				s.register(inf)
			}
			s.mu.Unlock()

		case <-tReregister.C: // force reregister service
			s.mu.Lock()
			for _, inf := range s.services {
				s.register(inf)
			}
			s.mu.Unlock()

		case <-tReconfig.C:
			s.log.Debug("reconfig service...")
			s.mu.Lock()
			for _, cfg := range s.svcConfigs {
				s.writeServiceConfig(cfg)
			}
			s.mu.Unlock()
		}
	}

	return nil
}

func (s *consulServiceImpl) Stop() {
	s.mu.Lock()
	if s.stopFunc != nil {
		s.stopFunc()
	}
	s.mu.Unlock()

	// make sure all services is removed
	s.RemoveAll()
}

func (s *consulServiceImpl) Cleanup() {}

// Add ID, address if not exists, append default tags
func (s *consulServiceImpl) normalizeASR(asr *consulapi.AgentServiceRegistration) {
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

func (s *consulServiceImpl) Add(asr *consulapi.AgentServiceRegistration) {
	if s.isDisabled() {
		return
	}

	s.normalizeASR(asr)

	inf := serviceStatus{svc: asr, status: false}
	s.register(&inf)

	s.mu.Lock()
	s.services[asr.ID] = &inf
	s.mu.Unlock()
}

func (s *consulServiceImpl) Remove(serviceID string) {
	if serviceID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, serviceID)

	s.log.Debugf("deregister service %s", serviceID)
	err := s.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		s.log.Warn(err)
	}
}

func (s *consulServiceImpl) RemoveAll() {
	agent := s.client.Agent()

	s.mu.Lock()
	defer s.mu.Unlock()
	for k, inf := range s.services {
		s.log.Debugf("deregister service %s", inf.svc.ID)
		agent.ServiceDeregister(inf.svc.ID)
		delete(s.services, k)
	}
}

func (s *consulServiceImpl) setCheckIP(remoteHost string) {
	conn, err := net.Dial("tcp", remoteHost)
	if err != nil {
		s.log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	s.checkIp = localAddr.IP.String()
	if strings.Contains(s.checkIp, ":") {
		s.checkIp = "[" + s.checkIp + "]"
	}
}

func (s *consulServiceImpl) GetCheckIP() string {
	return s.checkIp
}

func (s *consulServiceImpl) UpdateServiceConfig(cfg *ServiceConfig) {
	if s.isDisabled() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.svcConfigs[cfg.Name] = cfg

	s.writeServiceConfig(cfg)
}

// write config to consul
func (s *consulServiceImpl) writeServiceConfig(cfg *ServiceConfig) {
	kv := s.client.KV()

	key := fmt.Sprintf(SERVICE_CONFIG_PREFIX+"/%s/default", cfg.Name)
	value, _ := json.Marshal(cfg)

	// check if it is writed & valid
	// rewrite cause xds-api increase a little load
	s.log.Debugf("check service %s config", cfg.Name)
	oldKv, _, err := kv.Get(key, &consulapi.QueryOptions{})
	if err != nil {
		s.log.Error(err)
	}
	if oldKv != nil && string(value) == string(oldKv.Value) {
		return
	}

	p := &consulapi.KVPair{
		Key:   key,
		Value: value,
	}
	s.log.Debugf("update service %s config %s", cfg.Name, value)
	_, err = kv.Put(p, &consulapi.WriteOptions{})
	if err != nil {
		s.log.Error(err)
	}
}

func (s *consulServiceImpl) GetClient() *consulapi.Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client
}
