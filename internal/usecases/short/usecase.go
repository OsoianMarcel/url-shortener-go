package short

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/OsoianMarcel/url-shortener/internal/entities"
	util "github.com/OsoianMarcel/url-shortener/internal/utils"

	shortRepository "github.com/OsoianMarcel/url-shortener/internal/repositories/short"
)

var _ Usecase = (*usecase)(nil)

const (
	createCircuitBreaker = 10
)

type repository interface {
	InsertOne(ctx context.Context, shortLink entities.ShortLink) (string, error)
	FindOne(ctx context.Context, key string) (entities.ShortLink, error)
	FindOriginalURL(ctx context.Context, key string) (string, error)
	DeleteOne(ctx context.Context, key string) error
	IncreaseHits(ctx context.Context, key string) error
	FindStats(ctx context.Context, key string) (shortRepository.StatsModel, error)
}

type usecase struct {
	logger        *slog.Logger
	shortLinkRepo repository
	baseURL       string
}

func New(logger *slog.Logger, shortLinkRepository repository, baseURL string) Usecase {
	return &usecase{
		logger:        logger,
		shortLinkRepo: shortLinkRepository,
		baseURL:       baseURL,
	}
}

func (u *usecase) Create(ctx context.Context, createInput CreateInput) (CreateOutput, error) {
	// validate URL
	if _, err := url.ParseRequestURI(createInput.OriginalURL); err != nil {
		return CreateOutput{}, ErrInvalidURL
	}

	for i := range createCircuitBreaker {
		key := util.GenLinkKey()
		ent := entities.ShortLink{
			Key:         key,
			OriginalURL: createInput.OriginalURL,
			ShortURL:    createFullShortUrl(u.baseURL, key),
			CreatedAt:   time.Now(),
		}

		id, err := u.shortLinkRepo.InsertOne(ctx, ent)
		if err != nil {
			// retry if the key already exists
			if err == shortRepository.ErrShortLinkKeyExists {
				u.logger.Warn("link key already exists",
					slog.String("key", key),
					slog.Int("attempt", i),
				)
				continue
			}
			// otherwise return the error
			return CreateOutput{}, fmt.Errorf("Usecase.Create: insert: %w", err)
		}
		// after inserting, set the entity ID
		ent.ID = id

		return CreateOutput{Key: ent.Key, ShortURL: ent.ShortURL}, nil
	}

	return CreateOutput{}, errors.New("Usecase.Create: circuit breaker")
}

func (u *usecase) Delete(ctx context.Context, key string) error {
	err := u.shortLinkRepo.DeleteOne(ctx, key)
	if err != nil {
		return fmt.Errorf("Usecase.Delete (key: %s): %w", key, err)
	}

	return nil
}

func (u *usecase) Expand(ctx context.Context, key string) (entities.ShortLink, error) {
	shortURL, err := u.shortLinkRepo.FindOne(ctx, key)

	if err != nil {
		if err == shortRepository.ErrShortLinkNotFound {
			return entities.ShortLink{}, ErrShortLinkNotFound
		}

		return entities.ShortLink{}, fmt.Errorf("Usecase.Expand (key: %s): %w", key, err)
	}

	return shortURL, nil
}

func (u *usecase) OriginalURL(ctx context.Context, key string) (string, error) {
	originalURL, err := u.shortLinkRepo.FindOriginalURL(ctx, key)

	if err != nil {
		if err == shortRepository.ErrShortLinkNotFound {
			return "", ErrShortLinkNotFound
		}

		return "", fmt.Errorf("Usecase.OriginalURL (key: %s): %w", key, err)
	}

	err = u.shortLinkRepo.IncreaseHits(ctx, key)
	if err != nil {
		u.logger.Warn("failed to increase the link hits, continue", slog.Any("err", err))
	}

	return originalURL, nil
}

func (u *usecase) Stats(ctx context.Context, key string) (StatsOutput, error) {
	statsModel, err := u.shortLinkRepo.FindStats(ctx, key)

	if err != nil {
		if err == shortRepository.ErrShortLinkNotFound {
			return StatsOutput{}, ErrShortLinkNotFound
		}

		return StatsOutput{}, fmt.Errorf("Usecase.Stats (key: %s): %w", key, err)
	}

	return StatsOutput{
		Hits:      statsModel.Hits,
		CreatedAt: statsModel.CreatedAt,
	}, nil
}

func createFullShortUrl(baseURL string, key string) string {
	return fmt.Sprintf("%s/api/shortener/%s/redirect", baseURL, key)
}
