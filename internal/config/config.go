package config

import "fmt"

type Config struct {
	Business *BusinessConfig
	Http     *HttpConfig
	MongoDB  *MongoDBConfig
	Redis    *RedisConfig
}

func New() (*Config, error) {
	businessConfig, err := NewBusinessConfig()
	if err != nil {
		return nil, fmt.Errorf("init business config: %w", err)
	}

	httpConfig, err := NewHTTPConfig()
	if err != nil {
		return nil, fmt.Errorf("init http config: %w", err)
	}

	mongodbConfig, err := NewMongoDBConfig()
	if err != nil {
		return nil, fmt.Errorf("init mongodb config: %w", err)
	}

	redisConfig, err := NewRedisConfig()
	if err != nil {
		return nil, fmt.Errorf("init redis config: %w", err)
	}

	return &Config{
		Business: businessConfig,
		Http:     httpConfig,
		MongoDB:  mongodbConfig,
		Redis:    redisConfig,
	}, nil
}
