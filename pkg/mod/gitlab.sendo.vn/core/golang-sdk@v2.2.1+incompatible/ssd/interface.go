package ssd

import (
	consulapi "github.com/hashicorp/consul/api"

	sdms "gitlab.sendo.vn/core/golang-sdk"
)

type ConsulSD interface {
	// Register service to SSD
	Add(asr *consulapi.AgentServiceRegistration)
	// Deregister service from SSD
	Remove(serviceID string)
	// Deregister all services from SSD
	RemoveAll()
	// get ip used to check service
	GetCheckIP() string
	// SetServiceRoute set which route belong this service
	// The gateway need this information to route an incoming request
	UpdateServiceConfig(cfg *ServiceConfig)
	// Get raw client
	GetClient() *consulapi.Client
}

type ConsulService interface {
	sdms.RunnableService
	ConsulSD
}

// An SD service that do nothing
//
// Used for testing or to disable for some usecase
type nullConsulSD struct{}

func NewNullConsulSD() ConsulSD {
	return &nullConsulSD{}
}

func (n *nullConsulSD) Add(asr *consulapi.AgentServiceRegistration) {}
func (n *nullConsulSD) Remove(serviceID string)                     {}
func (n *nullConsulSD) RemoveAll()                                  {}
func (n *nullConsulSD) GetCheckIP() string {
	return "127.0.0.1"
}
func (n *nullConsulSD) UpdateServiceConfig(cfg *ServiceConfig) {}
func (n *nullConsulSD) GetClient() *consulapi.Client {
	return nil
}
