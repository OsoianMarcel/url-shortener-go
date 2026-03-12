package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/OsoianMarcel/url-shortener/internal/domain"
	"github.com/OsoianMarcel/url-shortener/pkg/randlinkkey"
)

var _ domain.ShortLinkUsecase = (*shortLinkUsecase)(nil)

const (
	createCircuitBreaker = 10
	linkKeyLength        = 6
)

type repository interface {
	InsertOne(ctx context.Context, shortLink domain.ShortLink) (string, error)
	FindOne(ctx context.Context, key string) (domain.ShortLink, error)
	FindOriginalURL(ctx context.Context, key string) (string, error)
	DeleteOne(ctx context.Context, key string) error
	IncreaseHits(ctx context.Context, key string) error
	FindStats(ctx context.Context, key string) (domain.StatsResult, error)
}

type shortLinkUsecase struct {
	logger        *slog.Logger
	shortLinkRepo repository
	baseURL       string
}

func NewShortLinkUsecase(logger *slog.Logger, shortLinkRepository repository, baseURL string) *shortLinkUsecase {
	return &shortLinkUsecase{
		logger:        logger,
		shortLinkRepo: shortLinkRepository,
		baseURL:       baseURL,
	}
}

func (u *shortLinkUsecase) Create(ctx context.Context, createInput domain.CreateAction) (domain.CreateResult, error) {
	// validate URL
	if _, err := url.ParseRequestURI(createInput.OriginalURL); err != nil {
		return domain.CreateResult{}, domain.ErrInvalidURL
	}

	for i := range createCircuitBreaker {
		key := randlinkkey.GenLinkKey(linkKeyLength)
		ent := domain.ShortLink{
			Key:         key,
			OriginalURL: createInput.OriginalURL,
			ShortURL:    createFullShortUrl(u.baseURL, key),
			CreatedAt:   time.Now(),
		}

		id, err := u.shortLinkRepo.InsertOne(ctx, ent)
		if err != nil {
			// retry if the key already exists
			if err == domain.ErrShortLinkKeyExists {
				u.logger.Warn("link key already exists",
					slog.String("key", key),
					slog.Int("attempt", i),
				)
				continue
			}
			// otherwise return the error
			return domain.CreateResult{}, fmt.Errorf("Usecase.Create: insert: %w", err)
		}
		// after inserting, set the entity ID
		ent.ID = id

		return domain.CreateResult{Key: ent.Key, ShortURL: ent.ShortURL}, nil
	}

	return domain.CreateResult{}, errors.New("Usecase.Create: circuit breaker")
}

func (u *shortLinkUsecase) Delete(ctx context.Context, key string) error {
	err := u.shortLinkRepo.DeleteOne(ctx, key)
	if err != nil {
		return fmt.Errorf("Usecase.Delete (key: %s): %w", key, err)
	}

	return nil
}

func (u *shortLinkUsecase) Expand(ctx context.Context, key string) (domain.ShortLink, error) {
	shortURL, err := u.shortLinkRepo.FindOne(ctx, key)

	if err != nil {
		if err == domain.ErrShortLinkNotFound {
			return domain.ShortLink{}, domain.ErrShortLinkNotFound
		}

		return domain.ShortLink{}, fmt.Errorf("Usecase.Expand (key: %s): %w", key, err)
	}

	return shortURL, nil
}

func (u *shortLinkUsecase) OriginalURL(ctx context.Context, key string) (string, error) {
	originalURL, err := u.shortLinkRepo.FindOriginalURL(ctx, key)

	if err != nil {
		if err == domain.ErrShortLinkNotFound {
			return "", domain.ErrShortLinkNotFound
		}

		return "", fmt.Errorf("Usecase.OriginalURL (key: %s): %w", key, err)
	}

	err = u.shortLinkRepo.IncreaseHits(ctx, key)
	if err != nil {
		u.logger.Warn("failed to increase the link hits, continue", slog.Any("err", err))
	}

	return originalURL, nil
}

func (u *shortLinkUsecase) Stats(ctx context.Context, key string) (domain.StatsResult, error) {
	statsModel, err := u.shortLinkRepo.FindStats(ctx, key)

	if err != nil {
		if err == domain.ErrShortLinkNotFound {
			return domain.StatsResult{}, domain.ErrShortLinkNotFound
		}

		return domain.StatsResult{}, fmt.Errorf("Usecase.Stats (key: %s): %w", key, err)
	}

	return domain.StatsResult{
		Hits:      statsModel.Hits,
		CreatedAt: statsModel.CreatedAt,
	}, nil
}

func createFullShortUrl(baseURL string, key string) string {
	return fmt.Sprintf("%s/api/shortener/%s/redirect", baseURL, key)
}
