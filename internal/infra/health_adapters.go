package infra

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoHealthAdapter struct {
	client *mongo.Client
}

func NewMongoHealthAdapter(client *mongo.Client) *MongoHealthAdapter {
	return &MongoHealthAdapter{client: client}
}

func (a *MongoHealthAdapter) Name() string { return "mongo" }

func (a *MongoHealthAdapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx, nil)
}

type RedisHealthAdapter struct {
	client *redis.Client
}

func NewRedisHealthAdapter(client *redis.Client) *RedisHealthAdapter {
	return &RedisHealthAdapter{client: client}
}

func (a *RedisHealthAdapter) Name() string { return "redis" }

func (a *RedisHealthAdapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx).Err()
}
