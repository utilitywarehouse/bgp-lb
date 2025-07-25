package main

// Result is the result of runing a health check.
type Result struct {
	healthy bool
	err     string
	output  string
}

// Checker is the interface that must be implemented by a healthcheck.
type Checker interface {
	Check() Result
}

// healthCheckSetup return a new healthcheck based on the service config
func healthCheckSetup(serviceConfig serviceConfig) Checker {
	if serviceConfig.HttpHealthCheck != nil {
		return NewHttpCheck(serviceConfig.HttpHealthCheck.Port)
	}
	if serviceConfig.PingHealthCheck != nil {
		return NewPingCheck(serviceConfig.PingHealthCheck.Addresses)
	}
	// Default to pinging well known DNS providers
	return NewPingCheck([]string{"1.1.1.1", "8.8.8.8"})
}
