package app

import (
	"log/slog"

	"github.com/OsoianMarcel/url-shortener/internal/config"
	"github.com/OsoianMarcel/url-shortener/internal/domain"
	"github.com/OsoianMarcel/url-shortener/internal/infra"
	"github.com/OsoianMarcel/url-shortener/internal/usecase"
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
	shortRepository domain.ShortLinkRepo
	shortUsecase    domain.ShortLinkUsecase
	healthUsecase   domain.HealthUsecase
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

func (sp *serviceProvider) getShortRepository() domain.ShortLinkRepo {
	if sp.shortRepository != nil {
		return sp.shortRepository
	}

	sp.shortRepository = infra.NewShortLinkRepository(
		sp.logger,
		sp.mongoClient.Database("shortener").Collection("short_links"),
		sp.redisClient,
	)

	return sp.shortRepository
}

func (sp *serviceProvider) getShortUsecase() domain.ShortLinkUsecase {
	if sp.shortUsecase != nil {
		return sp.shortUsecase
	}

	sp.shortUsecase = usecase.NewShortLinkUsecase(
		sp.logger,
		sp.getShortRepository(),
		sp.config.Business.BaseURL,
	)

	return sp.shortUsecase
}

func (sp *serviceProvider) getHealthUsecase() domain.HealthUsecase {
	if sp.healthUsecase != nil {
		return sp.healthUsecase
	}

	sp.healthUsecase = usecase.NewHealthUsecase(sp.logger, sp.mongoClient, sp.redisClient)

	return sp.healthUsecase
}
