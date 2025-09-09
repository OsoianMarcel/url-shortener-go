package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type BusinessConfig struct {
	BaseURL                 string
	LinkNotFoundRedirectURL string
	LinkExpiresAfter        time.Duration
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

	linkExpiresAfter := time.Hour * 365 * 24
	if durStr, ok := os.LookupEnv("LINK_EXPIRES_AFTER"); ok {
		var err error
		if linkExpiresAfter, err = time.ParseDuration(durStr); err != nil {
			return nil, fmt.Errorf("parse LINK_EXPIRES_AFTER: %w", err)
		}
	}

	return &BusinessConfig{
		BaseURL:                 baseURL,
		LinkExpiresAfter:        linkExpiresAfter,
		LinkNotFoundRedirectURL: linkNotFoundRedirectURL,
	}, nil
}
