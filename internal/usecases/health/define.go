package health

import (
	"context"
)

type Usecase interface {
	CheckHealth(ctx context.Context) HealthCheckOutput
}
