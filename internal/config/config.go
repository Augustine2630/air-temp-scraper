package config

import (
	"errors"
	"os"
	"time"
)

// Config holds all runtime configuration for the scraper service.
type Config struct {
	Port           string
	ScrapeInterval time.Duration
}

// Load reads configuration from environment variables.
// Returns an error if any required variable is missing or invalid.
func Load() (*Config, error) {
	port := os.Getenv("TEMP_SCRAPER_PORT")
	if port == "" {
		return nil, errors.New("required environment variable TEMP_SCRAPER_PORT is not set")
	}

	intervalStr := os.Getenv("TEMP_SCRAPER_INTERVAL")
	if intervalStr == "" {
		intervalStr = "30s"
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return nil, errors.New("invalid TEMP_SCRAPER_INTERVAL: " + err.Error())
	}

	return &Config{
		Port:           port,
		ScrapeInterval: interval,
	}, nil
}
