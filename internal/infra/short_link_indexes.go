package infra

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsureShortLinkIndexes(ctx context.Context, logger *slog.Logger, mongoClient *mongo.Client) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := mongoClient.Database(shortenerDBName).Collection(shortLinksCollectionName)
	indexesCursor, err := collection.Indexes().List(timeoutCtx)
	if err != nil {
		return fmt.Errorf("list indexes: %w", err)
	}
	defer indexesCursor.Close(timeoutCtx)

	for indexesCursor.Next(timeoutCtx) {
		var indexDoc bson.M
		if err := indexesCursor.Decode(&indexDoc); err != nil {
			return fmt.Errorf("decode index document: %w", err)
		}

		if isShortLinkKeyUniqueIndex(indexDoc) {
			logger.Debug("mongodb index already present; skipping creation",
				slog.String("collection", shortLinksCollectionName),
				slog.String("index", shortLinksUniqueKeyIndexName),
			)

			return nil
		}
	}

	if err := indexesCursor.Err(); err != nil {
		return fmt.Errorf("iterate indexes: %w", err)
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "key", Value: 1}},
		Options: options.Index().SetName(shortLinksUniqueKeyIndexName).SetUnique(true),
	}

	if _, err := collection.Indexes().CreateOne(timeoutCtx, indexModel); err != nil {
		return fmt.Errorf("create index %q: %w", shortLinksUniqueKeyIndexName, err)
	}

	logger.Info("mongodb index created",
		slog.String("collection", shortLinksCollectionName),
		slog.String("index", shortLinksUniqueKeyIndexName),
	)

	return nil
}

func isShortLinkKeyUniqueIndex(indexDoc bson.M) bool {
	indexName, hasName := indexDoc["name"].(string)
	if hasName && indexName == shortLinksUniqueKeyIndexName {
		return true
	}

	unique, _ := indexDoc["unique"].(bool)
	if !unique {
		return false
	}

	keyOrder, ok := getIndexKeyOrder(indexDoc["key"], "key")
	if !ok {
		return false
	}

	return keyOrder == 1
}

func getIndexKeyOrder(indexKeys any, key string) (int64, bool) {
	switch typed := indexKeys.(type) {
	case bson.M:
		return castIndexOrder(typed[key])
	case bson.D:
		for _, kv := range typed {
			if kv.Key == key {
				return castIndexOrder(kv.Value)
			}
		}
	}

	return 0, false
}

func castIndexOrder(v any) (int64, bool) {
	switch value := v.(type) {
	case int:
		return int64(value), true
	case int32:
		return int64(value), true
	case int64:
		return value, true
	case float64:
		return int64(value), true
	default:
		return 0, false
	}
}
