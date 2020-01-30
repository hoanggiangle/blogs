package registry

import consulAPI "github.com/hashicorp/consul/api"

type Agent interface {
	GetCheckIP() string
	RegisterService(*consulAPI.AgentServiceRegistration) string
	DeregisterService(string)
	// GetValueFromKey(string) ([]byte, error)
	GetKVs() []byte
	IsRunning() bool
}
