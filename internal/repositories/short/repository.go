package short

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/OsoianMarcel/url-shortener/internal/entities"
	"github.com/redis/go-redis/v9"
)

var _ Repository = (*repository)(nil)

type repository struct {
	logger     *slog.Logger
	collection *mongo.Collection
	redis      *redis.Client
}

func New(logger *slog.Logger, collection *mongo.Collection, redis *redis.Client) Repository {
	return &repository{
		logger:     logger,
		collection: collection,
		redis:      redis,
	}
}

func (r *repository) InsertOne(ctx context.Context, shortLink entities.ShortLink) (string, error) {
	// insert into DB
	doc := fromEntityToDocument(shortLink)
	res, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", ErrShortLinkKeyExists
		}

		return "", err
	}
	insertID := res.InsertedID.(primitive.ObjectID).Hex()

	// set entity ID
	shortLink.ID = insertID

	// trying to cache the entity
	if err := r.setEntityCache(ctx, shortLink); err != nil {
		r.logger.Warn("unable to cache short URL entity",
			slog.String("key", shortLink.Key),
			slog.String("originalURL", shortLink.OriginalURL),
			slog.Any("error", err),
		)
	}
	// trying to cache the original URL
	if err := r.setOriginalURLCache(ctx, shortLink.Key, shortLink.OriginalURL); err != nil {
		r.logger.Warn("unable to cache original URL for short link",
			slog.String("key", shortLink.Key),
			slog.String("originalURL", shortLink.OriginalURL),
			slog.Any("error", err),
		)
	}

	return insertID, nil
}

func (r *repository) FindOne(ctx context.Context, key string) (entities.ShortLink, error) {
	cachedEntity, err := r.getEntityCache(ctx, key)
	if err != nil {
		r.logger.Warn("unable to fetch cache for short URL entity",
			slog.String("key", key),
			slog.Any("error", err),
		)
	} else if cachedEntity != nil {
		return *cachedEntity, nil
	}

	doc := new(shortLinkDoc)
	err = r.collection.FindOne(ctx, bson.M{"key": key}).Decode(doc)
	if err == mongo.ErrNoDocuments {
		return entities.ShortLink{}, ErrShortLinkNotFound
	}
	if err != nil {
		return entities.ShortLink{}, err
	}

	ent := fromDocumentToEntity(*doc)

	// trying to cache the entity
	if err := r.setEntityCache(ctx, ent); err != nil {
		r.logger.Warn("unable to set cache for short URL entity",
			slog.String("key", key),
			slog.Any("error", err),
		)
	}

	return ent, nil
}

func (r *repository) FindOriginalURL(ctx context.Context, key string) (string, error) {
	// trying to fetch original URL form cache
	cachedOriginalURL, err := r.getOriginalURLCache(ctx, key)
	if err != nil {
		r.logger.Warn("unable to fetch cache for original URL",
			slog.String("key", key),
			slog.Any("err", err),
		)
	} else if cachedOriginalURL != "" {
		return cachedOriginalURL, nil
	}

	// find original URL form DB
	var result bson.M
	filter := bson.M{"key": key}
	projection := bson.M{"originalURL": 1, "_id": 0}
	err = r.collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return "", ErrShortLinkNotFound
	}
	if err != nil {
		return "", err
	}

	originalURL, ok := result["originalURL"].(string)
	if !ok {
		return "", fmt.Errorf("unable to find the property %q", originalURL)
	}

	err = r.setOriginalURLCache(ctx, key, originalURL)
	if err != nil {
		r.logger.Warn("unable to set cache for original URL",
			slog.String("key", key),
			slog.Any("err", err),
		)
	}

	return originalURL, nil
}

func (r *repository) DeleteOne(ctx context.Context, key string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"key": key})
	if err != nil {
		return err
	}

	err = r.deleteEntityCache(ctx, key)
	if err != nil {
		r.logger.Warn("unable to delete short URL entity from cache",
			slog.String("key", key),
			slog.Any("err", err),
		)
	}

	err = r.deleteOriginalURLCache(ctx, key)
	if err != nil {
		r.logger.Warn("unable to delete original URL form cache",
			slog.String("key", key),
			slog.Any("err", err),
		)
	}

	return nil
}

func (r *repository) IncreaseHits(ctx context.Context, key string) error {
	filter := bson.M{"key": key}
	update := bson.M{"$inc": bson.M{"hits": 1}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to increase hits for link with key %q: %w", key, err)
	}

	return nil
}

func (r *repository) FindStats(ctx context.Context, key string) (StatsModel, error) {
	statsDoc := new(statsDoc)
	filter := bson.M{"key": key}
	projection := bson.M{"hits": 1, "createdAt": 1, "_id": 0}
	err := r.collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&statsDoc)
	if err == mongo.ErrNoDocuments {
		return StatsModel{}, ErrShortLinkNotFound
	}
	if err != nil {
		return StatsModel{}, err
	}

	return StatsModel{
		Hits:      statsDoc.Hits,
		CreatedAt: statsDoc.CreatedAt.Time(),
	}, nil
}

func (r *repository) setOriginalURLCache(ctx context.Context, key string, originalURL string) error {
	cacheKey := genCacheKey(key, "originalURL")

	return r.redis.Set(ctx, cacheKey, originalURL, time.Hour*24).Err()
}

func (r *repository) getOriginalURLCache(ctx context.Context, key string) (string, error) {
	cacheKey := genCacheKey(key, "originalURL")

	originalURL, err := r.redis.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func (r *repository) deleteOriginalURLCache(ctx context.Context, key string) error {
	cacheKey := genCacheKey(key, "originalURL")

	return r.redis.Del(ctx, cacheKey).Err()
}

func (r *repository) setEntityCache(ctx context.Context, ent entities.ShortLink) error {
	cacheKey := genCacheKey(ent.Key, "entity")

	entData, err := json.Marshal(ent)
	if err != nil {
		return err
	}

	return r.redis.Set(ctx, cacheKey, entData, time.Hour*24).Err()
}

func (r *repository) getEntityCache(ctx context.Context, key string) (*entities.ShortLink, error) {
	cacheKey := genCacheKey(key, "entity")

	entData, err := r.redis.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	ent := new(entities.ShortLink)
	if err := json.Unmarshal([]byte(entData), ent); err != nil {
		return nil, err
	}

	return ent, nil
}

func (r *repository) deleteEntityCache(ctx context.Context, key string) error {
	cacheKey := genCacheKey(key, "entity")

	return r.redis.Del(ctx, cacheKey).Err()
}

func genCacheKey(dataKey string, cacheName string) string {
	return "shortener:" + cacheName + "#" + dataKey
}

func fromEntityToDocument(entity entities.ShortLink) shortLinkDoc {
	return shortLinkDoc{
		Key:         entity.Key,
		OriginalURL: entity.OriginalURL,
		ShortURL:    entity.ShortURL,
		Hits:        entity.Hits,
		CreatedAt:   primitive.NewDateTimeFromTime(entity.CreatedAt),
	}
}

func fromDocumentToEntity(model shortLinkDoc) entities.ShortLink {
	return entities.ShortLink{
		ID:          model.ID.Hex(),
		Key:         model.Key,
		OriginalURL: model.OriginalURL,
		ShortURL:    model.ShortURL,
		Hits:        model.Hits,
		CreatedAt:   model.CreatedAt.Time(),
	}
}
