package config

import (
	"errors"
	"os"
)

type Config struct {
	Port           string
	ISMCPath       string
	ScrapeInterval string
}

func Load() (*Config, error) {
	port := os.Getenv("TEMP_SCRAPER_PORT")
	if port == "" {
		return nil, errors.New("required environment variable TEMP_SCRAPER_PORT is not set")
	}

	ismcPath := os.Getenv("TEMP_SCRAPER_ISMC_PATH")
	if ismcPath == "" {
		ismcPath = "/Users/augustine/go/bin/iSMC"
	}

	interval := os.Getenv("TEMP_SCRAPER_INTERVAL")
	if interval == "" {
		interval = "30s"
	}

	return &Config{
		Port:           port,
		ISMCPath:       ismcPath,
		ScrapeInterval: interval,
	}, nil
}
