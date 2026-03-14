package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server wraps an HTTP server exposing /metrics.
type Server struct {
	srv *http.Server
}

// New creates an HTTP server on the given address.
func New(addr string, reg *prometheus.Registry) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry:            reg,
		EnableOpenMetrics:   true,
		MaxRequestsInFlight: 4,
	}))

	return &Server{
		srv: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}
}

// ListenAndServe starts the server. Blocks until the server stops.
func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
