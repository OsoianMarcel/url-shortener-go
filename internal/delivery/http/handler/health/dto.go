package health

import "time"

type serviceHealthDTO struct {
	Name          string `json:"name"`
	Healthy       bool   `json:"healthy"`
	Error         string `json:"error,omitempty"`
	CheckDuration string `json:"check_duration"`
}

type healthResponseDTO struct {
	AllHealthy bool               `json:"all_healthy"`
	Services   []serviceHealthDTO `json:"services"`
	ServerTime time.Time          `json:"server_time"`
}
