package app

import (
	"log/slog"

	"github.com/OsoianMarcel/url-shortener/internal/config"
	shortRepository "github.com/OsoianMarcel/url-shortener/internal/repositories/short"
	healthUsecase "github.com/OsoianMarcel/url-shortener/internal/usecases/health"
	shortUsecase "github.com/OsoianMarcel/url-shortener/internal/usecases/short"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type serviceProvider struct {
	// app dependencies
	logger      *slog.Logger
	config      *config.Config
	mongoClient *mongo.Client
	redisClient *redis.Client
	// initialized providers
	shortRepository shortRepository.Repository
	shortUsecase    shortUsecase.Usecase
	healthUsecase   healthUsecase.Usecase
}

func newServiceProvider(
	logger *slog.Logger,
	config *config.Config,
	mongoClient *mongo.Client,
	redisClient *redis.Client,
) *serviceProvider {
	return &serviceProvider{
		logger:      logger,
		config:      config,
		mongoClient: mongoClient,
		redisClient: redisClient,
	}
}

func (sp *serviceProvider) getShortRepository() shortRepository.Repository {
	if sp.shortRepository != nil {
		return sp.shortRepository
	}

	sp.shortRepository = shortRepository.New(
		sp.logger,
		sp.mongoClient.Database("shortener").Collection("short_links"),
		sp.redisClient,
	)

	return sp.shortRepository
}

func (sp *serviceProvider) getShortUsecase() shortUsecase.Usecase {
	if sp.shortUsecase != nil {
		return sp.shortUsecase
	}

	sp.shortUsecase = shortUsecase.New(
		sp.logger,
		sp.getShortRepository(),
		sp.config.Business.BaseURL,
	)

	return sp.shortUsecase
}

func (sp *serviceProvider) getHealthUsecase() healthUsecase.Usecase {
	if sp.healthUsecase != nil {
		return sp.healthUsecase
	}

	sp.healthUsecase = healthUsecase.New(sp.logger, sp.mongoClient, sp.redisClient)

	return sp.healthUsecase
}
