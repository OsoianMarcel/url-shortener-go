package domain

import (
	"context"
)

type ShortLinkRepo interface {
	InsertOne(ctx context.Context, shortLink ShortLink) (string, error)
	FindOne(ctx context.Context, key string) (ShortLink, error)
	FindOriginalURL(ctx context.Context, key string) (string, error)
	DeleteOne(ctx context.Context, key string) error
	IncreaseHits(ctx context.Context, key string) error
	FindStats(ctx context.Context, key string) (StatsResult, error)
}
