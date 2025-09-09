package config

import (
	"fmt"
	"os"
)

type MongoDBConfig struct {
	URI string
}

func NewMongoDBConfig() (*MongoDBConfig, error) {
	URI, ok := os.LookupEnv("MONGODB_URI")
	if !ok {
		return nil, fmt.Errorf("MONGODB_URI env variable is not defined")
	}

	return &MongoDBConfig{
		URI: URI,
	}, nil
}
