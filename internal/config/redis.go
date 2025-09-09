package config

import (
	"net"
	"os"
)

type RedisConfig struct {
	Host string
	Port string
}

func NewRedisConfig() (*RedisConfig, error) {
	host, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		host = "localhost"
	}

	port, ok := os.LookupEnv("REDIS_PORT")
	if !ok {
		port = "6379"
	}

	return &RedisConfig{
		Host: host,
		Port: port,
	}, nil
}

func (c *RedisConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
