package scraper

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/cherepovskiy/air-temp-scraper/internal/metrics"
	"github.com/cherepovskiy/air-temp-scraper/internal/parser"
)

// Scraper runs iSMC and parses its JSON output into Prometheus metrics.
type Scraper struct {
	ismcPath string
	metrics  *metrics.Metrics
}

func New(ismcPath string, m *metrics.Metrics) *Scraper {
	return &Scraper{
		ismcPath: ismcPath,
		metrics:  m,
	}
}

// Scrape performs one exec-parse cycle and updates all metrics.
func (s *Scraper) Scrape() {
	s.metrics.ScrapeTotal.Inc()

	start := time.Now()
	err := s.scrapeOnce()
	elapsed := time.Since(start).Seconds()

	s.metrics.ScrapeDuration.Observe(elapsed)
	s.metrics.LastScrapeDuration.Set(elapsed)

	if err != nil {
		s.metrics.ScrapeErrorsTotal.Inc()
		log.Printf("scrape error: %v", err)
		return
	}
	s.metrics.ScrapeSuccessTotal.Inc()
}

func (s *Scraper) scrapeOnce() error {
	// #nosec G204 — path comes from trusted config, not user input
	cmd := exec.Command(s.ismcPath, "temp", "-o", "json")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("iSMC exec: %w (stderr: %s)", err, stderr.String())
	}

	parseStart := time.Now()
	var count float64

	err := parser.Parse(&stdout, func(reading parser.SensorReading) {
		if reading.Quantity == nil {
			return
		}
		s.metrics.SensorTemperature.
			WithLabelValues(reading.Name, reading.Key, reading.Type).
			Set(*reading.Quantity)
		count++
	})

	s.metrics.ParseDuration.Observe(time.Since(parseStart).Seconds())

	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	s.metrics.ParsedSensors.Set(count)
	return nil
}
