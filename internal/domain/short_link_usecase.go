package domain

import "context"

type ShortLinkUsecase interface {
	Create(ctx context.Context, createAction CreateAction) (CreateResult, error)
	Expand(ctx context.Context, key string) (ShortLink, error)
	OriginalURL(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Stats(ctx context.Context, key string) (StatsResult, error)
}
