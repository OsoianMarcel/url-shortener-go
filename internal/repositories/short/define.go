package short

import (
	"context"

	"github.com/OsoianMarcel/url-shortener/internal/entities"
)

type Repository interface {
	InsertOne(ctx context.Context, shortLink entities.ShortLink) (string, error)
	FindOne(ctx context.Context, key string) (entities.ShortLink, error)
	FindOriginalURL(ctx context.Context, key string) (string, error)
	DeleteOne(ctx context.Context, key string) error
	IncreaseHits(ctx context.Context, key string) error
	FindStats(ctx context.Context, key string) (StatsModel, error)
}
