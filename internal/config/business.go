package config

import (
	"fmt"
	"os"
	"strings"
)

type BusinessConfig struct {
	BaseURL                 string
	LinkNotFoundRedirectURL string
}

func NewBusinessConfig() (*BusinessConfig, error) {
	baseURL, ok := os.LookupEnv("BASE_URL")
	if !ok {
		return nil, fmt.Errorf("BASE_URL env variable is not defined")
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	linkNotFoundRedirectURL, ok := os.LookupEnv("LINK_NOT_FOUND_REDIRECT_URL")
	if !ok {
		return nil, fmt.Errorf("LINK_NOT_FOUND_REDIRECT_URL env variable is not defined")
	}

	return &BusinessConfig{
		BaseURL:                 baseURL,
		LinkNotFoundRedirectURL: linkNotFoundRedirectURL,
	}, nil
}
