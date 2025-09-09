package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/OsoianMarcel/url-shortener/internal/config"
	commonHTTPHandler "github.com/OsoianMarcel/url-shortener/internal/delivery/http/handlers/common"
	healthHTTPHandler "github.com/OsoianMarcel/url-shortener/internal/delivery/http/handlers/health"
	shortHTTPHandler "github.com/OsoianMarcel/url-shortener/internal/delivery/http/handlers/short"
	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/middlewares"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type app struct {
	logger          *slog.Logger
	mongoClient     *mongo.Client
	redisClient     *redis.Client
	serviceProvider *serviceProvider
	httpServer      *http.Server
}

func New() (*app, error) {
	a := &app{}

	return a, nil
}

func (a *app) Serve(ctx context.Context) error {
	var err error

	a.logger = initLogger()

	conf, err := config.New()
	if err != nil {
		return fmt.Errorf("init config: %w", err)
	}

	a.mongoClient, err = initMongoDB(ctx)
	if err != nil {
		return fmt.Errorf("init mongodb: %w", err)
	}

	a.redisClient, err = initRedis(ctx)
	if err != nil {
		return fmt.Errorf("init redis: %w", err)
	}

	a.serviceProvider = newServiceProvider(
		a.logger,
		conf,
		a.mongoClient,
		a.redisClient,
	)

	a.httpServer, err = initHTTPServer(a.serviceProvider)
	if err != nil {
		return fmt.Errorf("init http server: %w", err)
	}

	a.logger.Info("starting the HTTP server", slog.String("addr", a.serviceProvider.config.Http.Address()))

	if err := a.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("listen HTTP server: %w", err)
	}

	return nil
}

func (a *app) Shutdown(ctx context.Context) error {
	var allErr error
	var err error

	err = a.httpServer.Shutdown(ctx)
	if err != nil {
		allErr = errors.Join(allErr, err)
	}

	err = a.serviceProvider.mongoClient.Disconnect(ctx)
	if err != nil {
		allErr = errors.Join(allErr, err)
	}

	err = a.serviceProvider.redisClient.Close()
	if err != nil {
		allErr = errors.Join(allErr, err)
	}

	if allErr != nil {
		return fmt.Errorf("graceful shutdown: %w", allErr)
	}

	a.logger.Info("shutdown completed successfully")

	return nil
}

func initHTTPServer(sp *serviceProvider) (*http.Server, error) {
	httpConfig, err := config.NewHTTPConfig()
	if err != nil {
		return nil, fmt.Errorf("load http config: %w", err)
	}
	businessConfig, err := config.NewBusinessConfig()
	if err != nil {
		return nil, fmt.Errorf("load business config: %w", err)
	}

	mux := http.NewServeMux()

	// Shortener handlers.
	shortHTTPHandler.RegisterHandler(
		mux,
		sp.logger,
		sp.getShortUsecase(),
		httpConfig.APISecret,
		businessConfig.LinkNotFoundRedirectURL,
	)

	// Health handlers.
	healthHTTPHandler.RegisterHandler(mux, sp.logger, sp.getHealthUsecase())

	// Common handlers.
	commonHTTPHandler.RegisterHandler(mux, sp.logger, "./api/openapi-spec.yaml")

	httpServer := &http.Server{
		Addr: httpConfig.Address(),
		Handler: middlewares.Chain(
			commonHTTPHandler.PreflightHandler(mux),
			middlewares.LoggingMiddleware(sp.logger),
		),
	}

	return httpServer, nil
}

func initLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: false}))
}

func initRedis(ctx context.Context) (*redis.Client, error) {
	redisConfig, err := config.NewRedisConfig()
	if err != nil {
		return nil, fmt.Errorf("load redis config: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisConfig.Address(),
	})

	timeoutCtx, cancelCtx := context.WithTimeout(ctx, 3*time.Second)
	defer cancelCtx()

	if pong := client.Ping(timeoutCtx); pong.Err() != nil {
		return nil, fmt.Errorf("ping redis: %w", pong.Err())
	}

	return client, nil
}

func initMongoDB(ctx context.Context) (*mongo.Client, error) {
	mongoDBConfig, err := config.NewMongoDBConfig()
	if err != nil {
		return nil, fmt.Errorf("load mongodb config: %w", err)
	}

	timeoutCtx, cancelCtx := context.WithTimeout(ctx, 5*time.Second)
	defer cancelCtx()

	client, err := mongo.Connect(timeoutCtx, options.Client().
		ApplyURI(mongoDBConfig.URI).SetConnectTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("connect to MongoDB: %w", err)
	}

	if err := client.Ping(timeoutCtx, nil); err != nil {
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	return client, nil
}
