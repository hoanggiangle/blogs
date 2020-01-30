package sendo

// Service watcher is a process help pull configs from Consul and
// restart service if needed.
// We have two types of configs: global and specific service
func (s *service) startServiceWatcher() func() {
	return func() {}
}
