package usecase

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/OsoianMarcel/url-shortener/internal/domain"
)

var _ domain.HealthUsecase = (*healthUsecase)(nil)

type healthUsecase struct {
	logger       *slog.Logger
	dependencies []domain.HealthDependency
}

func NewHealthUsecase(
	logger *slog.Logger,
	dependencies ...domain.HealthDependency,
) *healthUsecase {
	return &healthUsecase{
		logger:       logger,
		dependencies: dependencies,
	}
}

func (u *healthUsecase) CheckHealth(ctx context.Context) domain.HealthCheckResult {
	services := make([]domain.ServiceHealth, 0, len(u.dependencies))
	rc := make(chan domain.ServiceHealth)

	// run health checks in concurrently
	wg := new(sync.WaitGroup)
	for _, dependency := range u.dependencies {
		dep := dependency
		wg.Go(func() { checkDependencyHealth(ctx, dep, rc) })
	}

	// close channel when goroutines are done
	go func() {
		wg.Wait()
		close(rc)
	}()

	// receive the results from the channel
	for r := range rc {
		services = append(services, r)
	}

	allHealthy := true
	for _, service := range services {
		if !service.Healthy {
			allHealthy = false
			break
		}
	}

	if !allHealthy {
		u.logger.Warn("Usecase.CheckHealth: unhealthy service(s)", slog.Any("services", services))
	}

	output := domain.HealthCheckResult{
		AllHealthy: allHealthy,
		Services:   services,
	}

	return output
}

func checkDependencyHealth(ctx context.Context, dependency domain.HealthDependency, rc chan<- domain.ServiceHealth) {
	model := domain.ServiceHealth{
		Name:    dependency.Name(),
		Healthy: true,
	}

	start := time.Now()
	err := dependency.Ping(ctx)
	model.CheckDuration = time.Since(start)

	if err != nil {
		model.Healthy = false
		model.Error = err.Error()
	}

	rc <- model
}
