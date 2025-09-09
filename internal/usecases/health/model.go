package health

import "time"

type ServiceHealth struct {
	Name          string
	Healthy       bool
	Error         string
	CheckDuration time.Duration
}

type HealthCheckOutput struct {
	AllHealthy bool
	Services   []ServiceHealth
}
