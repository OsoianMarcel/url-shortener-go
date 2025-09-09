package short

import (
	"context"

	"github.com/OsoianMarcel/url-shortener/internal/entities"
)

type Usecase interface {
	Create(ctx context.Context, createInput CreateInput) (CreateOutput, error)
	Expand(ctx context.Context, key string) (entities.ShortLink, error)
	OriginalURL(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Stats(ctx context.Context, key string) (StatsOutput, error)
}
