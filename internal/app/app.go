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
	"github.com/OsoianMarcel/url-shortener/internal/delivery/cli"
	"github.com/OsoianMarcel/url-shortener/internal/delivery/cli/command"
	commonHTTPHandler "github.com/OsoianMarcel/url-shortener/internal/delivery/http/handler/common"
	healthHTTPHandler "github.com/OsoianMarcel/url-shortener/internal/delivery/http/handler/health"
	shortHTTPHandler "github.com/OsoianMarcel/url-shortener/internal/delivery/http/handler/short"
	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/middleware"
	"github.com/OsoianMarcel/url-shortener/internal/infra"
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
	initialized     bool
}

func New() (*app, error) {
	a := &app{}

	return a, nil
}

func (a *app) Init(ctx context.Context) error {
	if a.initialized {
		return nil
	}

	var err error

	a.logger = initLogger()

	conf, err := config.New()
	if err != nil {
		return fmt.Errorf("init config: %w", err)
	}

	a.mongoClient, err = initMongoDB(ctx, conf.MongoDB)
	if err != nil {
		return fmt.Errorf("init mongodb: %w", err)
	}

	if err := infra.EnsureShortLinkIndexes(ctx, a.logger, a.mongoClient); err != nil {
		return fmt.Errorf("ensure mongodb indexes: %w", err)
	}

	a.redisClient, err = initRedis(ctx, conf.Redis)
	if err != nil {
		return fmt.Errorf("init redis: %w", err)
	}

	a.serviceProvider = newServiceProvider(
		a.logger,
		conf,
		a.mongoClient,
		a.redisClient,
	)

	a.httpServer = initHTTPServer(a.serviceProvider)
	a.initialized = true

	return nil
}

// ServeHTTP inits the app and starts the HTTP server.
func (a *app) ServeHTTP(ctx context.Context) error {
	if err := a.Init(ctx); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	a.logger.Info("starting the HTTP server", slog.String("addr", a.serviceProvider.config.Http.Address()))

	if err := a.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("listen HTTP server: %w", err)
	}

	return nil
}

// ServeCLI inits the app, registers CLI commands with their dependencies
// from the DI container, and runs the CLI against the provided args.
func (a *app) ServeCLI(ctx context.Context, args []string) error {
	if err := a.Init(ctx); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	shortCmd := command.NewShortCommand(a.serviceProvider.getShortLinkUsecase())

	return cli.Run(ctx, args, os.Stdout, os.Stderr, shortCmd)
}

func (a *app) Shutdown(ctx context.Context) error {
	var allErr error
	var err error

	if a.httpServer != nil {
		err = a.httpServer.Shutdown(ctx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			allErr = errors.Join(allErr, err)
		}
	}

	if a.mongoClient != nil {
		err = a.mongoClient.Disconnect(ctx)
		if err != nil {
			allErr = errors.Join(allErr, err)
		}
	}

	if a.redisClient != nil {
		err = a.redisClient.Close()
		if err != nil {
			allErr = errors.Join(allErr, err)
		}
	}

	if allErr != nil {
		return fmt.Errorf("graceful shutdown: %w", allErr)
	}

	return nil
}

func initHTTPServer(sp *serviceProvider) *http.Server {
	mux := http.NewServeMux()

	// Shortener handlers.
	shortHTTPHandler.RegisterHandler(
		mux,
		sp.logger,
		sp.getShortLinkUsecase(),
		sp.config.Http.APISecret,
		sp.config.Business.LinkNotFoundRedirectURL,
	)

	// Health handlers.
	healthHTTPHandler.RegisterHandler(mux, sp.logger, sp.getHealthUsecase())

	// Common handlers.
	commonHTTPHandler.RegisterHandler(mux, sp.logger, "./api/openapi-spec.yaml")

	httpServer := &http.Server{
		Addr: sp.config.Http.Address(),
		Handler: middleware.Chain(
			commonHTTPHandler.PreflightHandler(mux),
			middleware.LoggingMiddleware(sp.logger),
			middleware.RecoverMiddleware(sp.logger),
		),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return httpServer
}

func initLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: false}))
}

func initRedis(ctx context.Context, redisConfig *config.RedisConfig) (*redis.Client, error) {
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

func initMongoDB(ctx context.Context, mongoDBConfig *config.MongoDBConfig) (*mongo.Client, error) {
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
