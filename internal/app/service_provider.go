package app

import (
	"fmt"
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
	shortLinkRep     domain.ShortLinkRepo
	shortLinkUsecase domain.ShortLinkUsecase
	healthUsecase    domain.HealthUsecase
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

func (sp *serviceProvider) getShortRepo() domain.ShortLinkRepo {
	if sp.shortLinkRep != nil {
		return sp.shortLinkRep
	}

	sp.shortLinkRep = infra.NewShortLinkRepository(
		sp.logger,
		sp.mongoClient,
		sp.redisClient,
	)

	return sp.shortLinkRep
}

func (sp *serviceProvider) getShortLinkUsecase() domain.ShortLinkUsecase {
	if sp.shortLinkUsecase != nil {
		return sp.shortLinkUsecase
	}

	buildShortURL := func(key string) string {
		return fmt.Sprintf("%s/api/shortener/%s/redirect", sp.config.Business.BaseURL, key)
	}

	sp.shortLinkUsecase = usecase.NewShortLinkUsecase(
		sp.logger,
		sp.getShortRepo(),
		buildShortURL,
	)

	return sp.shortLinkUsecase
}

func (sp *serviceProvider) getHealthUsecase() domain.HealthUsecase {
	if sp.healthUsecase != nil {
		return sp.healthUsecase
	}

	sp.healthUsecase = usecase.NewHealthUsecase(
		sp.logger,
		infra.NewMongoHealthAdapter(sp.mongoClient),
		infra.NewRedisHealthAdapter(sp.redisClient),
	)

	return sp.healthUsecase
}
