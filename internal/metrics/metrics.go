package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const namespace = "temp_scraper"

// Metrics holds all Prometheus metrics for the scraper service.
type Metrics struct {
	// Scrape runtime
	ScrapeTotal        prometheus.Counter
	ScrapeErrorsTotal  prometheus.Counter
	ScrapeSuccessTotal prometheus.Counter
	ScrapeDuration     prometheus.Histogram
	LastScrapeDuration prometheus.Gauge

	// Parse runtime
	ParseDuration prometheus.Histogram
	ParsedSensors prometheus.Gauge

	// Per-sensor temperature gauge (labels: sensor_name, sensor_key, sensor_type)
	SensorTemperature *prometheus.GaugeVec

	Registry *prometheus.Registry
}

func New() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	m := &Metrics{
		Registry: reg,

		ScrapeTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "scrape",
			Name:      "attempts_total",
			Help:      "Total number of scrape attempts.",
		}),
		ScrapeErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "scrape",
			Name:      "errors_total",
			Help:      "Total number of failed scrape attempts.",
		}),
		ScrapeSuccessTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "scrape",
			Name:      "success_total",
			Help:      "Total number of successful scrape attempts.",
		}),
		ScrapeDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "scrape",
			Name:      "duration_seconds",
			Help:      "Histogram of scrape latency in seconds.",
			Buckets:   prometheus.ExponentialBuckets(0.005, 2, 10),
		}),
		LastScrapeDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "scrape",
			Name:      "last_duration_seconds",
			Help:      "Duration of the most recent scrape in seconds.",
		}),
		ParseDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "parse",
			Name:      "duration_seconds",
			Help:      "Histogram of JSON parse latency in seconds.",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
		}),
		ParsedSensors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "parse",
			Name:      "sensors_total",
			Help:      "Number of sensor readings parsed in the last scrape.",
		}),
		SensorTemperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "sensor",
			Name:      "temperature_celsius",
			Help:      "Current temperature reading per sensor in degrees Celsius.",
		}, []string{"sensor_name", "sensor_key", "sensor_type"}),
	}

	reg.MustRegister(
		m.ScrapeTotal,
		m.ScrapeErrorsTotal,
		m.ScrapeSuccessTotal,
		m.ScrapeDuration,
		m.LastScrapeDuration,
		m.ParseDuration,
		m.ParsedSensors,
		m.SensorTemperature,
	)

	return m
}
