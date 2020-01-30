package sdms

// non-runnable service
type Service interface {
	// config flags here
	InitFlags()

	// configure service
	Configure() error

	// Cleanup service like db connection, remove temp files
	Cleanup()
}

type ReloadableService interface {
	Service

	// Run when receive HUP signal
	// Used to reload config (if read from file)
	// or log rotation
	Reload()
}

// Runnable service
type RunnableService interface {
	Service

	// Start main logic
	Run() error

	// Use to stop from another goroutine
	Stop()
}
