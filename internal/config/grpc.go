package config

import (
	"net"
	"os"
)

type GrpcConfig struct {
	Host string
	Port string
}

func NewGrpcConfig() (*GrpcConfig, error) {
	host := os.Getenv("GRPC_HOST")

	port, ok := os.LookupEnv("GRPC_PORT")
	if !ok {
		port = "50051"
	}

	return &GrpcConfig{
		Host: host,
		Port: port,
	}, nil
}

func (c *GrpcConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
