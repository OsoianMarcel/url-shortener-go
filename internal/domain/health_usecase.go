package domain

import "context"

type HealthUsecase interface {
	CheckHealth(ctx context.Context) HealthCheckResult
}
