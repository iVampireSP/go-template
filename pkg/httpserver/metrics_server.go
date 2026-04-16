package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/iVampireSP/go-template/pkg/logger"
)

// MetricsConfig holds metrics server configuration.
type MetricsConfig struct {
	Name            string
	Enabled         bool
	Host            string
	Port            int
	ShutdownTimeout time.Duration
}

// MetricsServer exposes /metrics and /healthz endpoints.
type MetricsServer struct {
	cfg    MetricsConfig
	server *http.Server
}

// newMetricsServer creates a new metrics server.
func newMetricsServer(cfg MetricsConfig) *MetricsServer {
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 30 * time.Second
	}
	return &MetricsServer{
		cfg: cfg,
	}
}

// Start starts the metrics server in a goroutine.
func (s *MetricsServer) Start() {
	if !s.cfg.Enabled {
		return
	}

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Handle("/metrics", promhttp.Handler())

	addr := s.Addr()
	s.server = &http.Server{Addr: addr, Handler: r}

	go func() {
		logger.Info("metrics server listening", "name", s.cfg.Name, "on", addr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("metrics server", "name", s.cfg.Name, "error", err)
		}
	}()
}

// Shutdown gracefully shuts down the metrics server.
func (s *MetricsServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// ShutdownWithTimeout gracefully shuts down the metrics server.
func (s *MetricsServer) ShutdownWithTimeout() {
	if s.server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("metrics server shutdown", "name", s.cfg.Name, "error", err)
	}
}

// Addr returns the metrics server address.
func (s *MetricsServer) Addr() string {
	return fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
}
