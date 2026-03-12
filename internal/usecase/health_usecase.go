package usecase

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/OsoianMarcel/url-shortener/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ domain.HealthUsecase = (*healthUsecase)(nil)

type healthUsecase struct {
	logger      *slog.Logger
	mongoClient *mongo.Client
	redisClient *redis.Client
}

func NewHealthUsecase(
	logger *slog.Logger,
	mongoClient *mongo.Client,
	redisClient *redis.Client,
) *healthUsecase {
	return &healthUsecase{
		logger:      logger,
		mongoClient: mongoClient,
		redisClient: redisClient,
	}
}

func (u *healthUsecase) CheckHealth(ctx context.Context) domain.HealthCheckResult {
	services := make([]domain.ServiceHealth, 0, 2)
	rc := make(chan domain.ServiceHealth)

	// run health checks in concurrently
	wg := new(sync.WaitGroup)
	wg.Go(func() { getMongoHealth(ctx, u.mongoClient, rc) })
	wg.Go(func() { getRedisHealth(ctx, u.redisClient, rc) })

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

func getMongoHealth(ctx context.Context, mongoClient *mongo.Client, rc chan<- domain.ServiceHealth) {
	model := domain.ServiceHealth{
		Name:    "mongo",
		Healthy: true,
	}

	start := time.Now()
	err := mongoClient.Ping(ctx, nil)
	model.CheckDuration = time.Since(start)

	if err != nil {
		model.Healthy = false
		model.Error = err.Error()
	}

	rc <- model
}

func getRedisHealth(ctx context.Context, redisClient *redis.Client, rc chan<- domain.ServiceHealth) {
	model := domain.ServiceHealth{
		Name:    "redis",
		Healthy: true,
	}

	start := time.Now()
	statusCmd := redisClient.Ping(ctx)
	model.CheckDuration = time.Since(start)

	err := statusCmd.Err()
	if err != nil {
		model.Healthy = false
		model.Error = err.Error()
	}

	rc <- model
}
