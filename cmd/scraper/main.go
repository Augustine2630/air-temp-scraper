package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cherepovskiy/air-temp-scraper/internal/config"
	"github.com/cherepovskiy/air-temp-scraper/internal/httpserver"
	"github.com/cherepovskiy/air-temp-scraper/internal/metrics"
	"github.com/cherepovskiy/air-temp-scraper/internal/scraper"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	m := metrics.New()

	sc, err := scraper.New(m)
	if err != nil {
		log.Fatalf("failed to open SMC: %v", err)
	}
	defer sc.Close()

	srv := httpserver.New(":"+cfg.Port, m.Registry)

	go func() {
		log.Printf("metrics server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http server error: %v", err)
			os.Exit(1)
		}
	}()

	// Initial scrape immediately on startup.
	sc.Scrape()

	ticker := time.NewTicker(cfg.ScrapeInterval)
	defer ticker.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			sc.Scrape()
		case sig := <-stop:
			log.Printf("received signal %s, shutting down", sig)
			ctx, cancel := context.WithTimeout(context.Background(), 5*cfg.ScrapeInterval)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("graceful shutdown error: %v", err)
			}
			return
		}
	}
}
