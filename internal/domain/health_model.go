package domain

import "time"

type ServiceHealth struct {
	Name          string
	Healthy       bool
	Error         string
	CheckDuration time.Duration
}

type HealthCheckResult struct {
	AllHealthy bool
	Services   []ServiceHealth
}
