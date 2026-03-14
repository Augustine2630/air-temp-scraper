//go:build darwin

package scraper

import (
	"log"
	"time"

	"github.com/cherepovskiy/air-temp-scraper/internal/metrics"
	"github.com/cherepovskiy/air-temp-scraper/internal/smc"
)

// Scraper holds a persistent SMC connection and updates Prometheus metrics
// on each scrape cycle without spawning processes or re-opening the connection.
type Scraper struct {
	reader  *smc.Reader
	metrics *metrics.Metrics
}

// New opens the SMC connection and returns a ready Scraper.
// Returns an error if the SMC cannot be opened (e.g. non-macOS or no permission).
func New(m *metrics.Metrics) (*Scraper, error) {
	r, err := smc.Open()
	if err != nil {
		return nil, err
	}
	return &Scraper{reader: r, metrics: m}, nil
}

// Close releases the underlying SMC connection.
func (s *Scraper) Close() {
	s.reader.Close()
}

// Scrape performs one read-parse cycle and updates all Prometheus metrics.
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
	parseStart := time.Now()

	readings := s.reader.ReadTemperatures()

	s.metrics.ParseDuration.Observe(time.Since(parseStart).Seconds())
	s.metrics.ParsedSensors.Set(float64(len(readings)))

	for _, rd := range readings {
		s.metrics.SensorTemperature.
			WithLabelValues(rd.Desc, rd.Key, rd.Type).
			Set(float64(rd.Value))
	}

	return nil
}
