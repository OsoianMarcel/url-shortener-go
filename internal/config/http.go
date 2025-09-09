package config

import (
	"fmt"
	"net"
	"os"
)

type HttpConfig struct {
	Host      string
	Port      string
	APISecret string
}

func NewHTTPConfig() (*HttpConfig, error) {
	host := os.Getenv("HTTP_HOST")

	port, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		port = "3000"
	}

	apiSecret, ok := os.LookupEnv("API_SECRET")
	if !ok {
		return nil, fmt.Errorf("API_SECRET env variable is not defined")
	}

	return &HttpConfig{
		Host:      host,
		Port:      port,
		APISecret: apiSecret,
	}, nil
}

func (c *HttpConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
