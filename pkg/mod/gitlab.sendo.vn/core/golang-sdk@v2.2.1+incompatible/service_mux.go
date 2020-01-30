package sdms

import "sync"

type ServiceMux interface {
	RunnableService
	GetServices() []RunnableService
}

type serviceMux struct {
	services []RunnableService
}

// a service mux that run services in parallel
// If one service is stopped, other services will stop too
func NewServiceMux(services ...RunnableService) ServiceMux {
	return &serviceMux{services}
}

func (m *serviceMux) InitFlags() {
	for _, s := range m.services {
		s.InitFlags()
	}
}

func (m *serviceMux) Configure() error {
	for _, s := range m.services {
		if err := s.Configure(); err != nil {
			return err
		}
	}
	return nil
}

// run all service in parallel. Return when one service die
func (m *serviceMux) Run() error {
	chErr := make(chan error, len(m.services))

	for _, s := range m.services {
		go func(s RunnableService) {
			chErr <- s.Run()
		}(s)
	}

	return <-chErr
}

// stop all services in parallel
func (m *serviceMux) Stop() {
	var wg sync.WaitGroup
	wg.Add(len(m.services))
	for _, s := range m.services {
		go func(s RunnableService) {
			s.Stop()
			wg.Done()
		}(s)
	}
	wg.Wait()
}

func (m *serviceMux) Cleanup() {
	for _, s := range m.services {
		s.Cleanup()
	}
}

func (m *serviceMux) GetServices() []RunnableService {
	return m.services
}
