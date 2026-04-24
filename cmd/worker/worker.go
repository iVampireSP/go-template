package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iVampireSP/go-template/internal/infra/config"
	"github.com/iVampireSP/go-template/internal/infra/tracing"
	"github.com/iVampireSP/go-template/internal/job"
	jobqueue "github.com/iVampireSP/go-template/pkg/foundation/queue"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/iVampireSP/go-template/pkg/logger"

	"github.com/spf13/cobra"
)

// Worker holds Worker service dependencies.
type Worker struct {
	queue    *jobqueue.Queue
	handlers []job.Handler
}

// NewWorker declares Worker command dependencies.
func NewWorker(queue *jobqueue.Queue, handlers []job.Handler) *Worker {
	return &Worker{
		queue:    queue,
		handlers: handlers,
	}
}

// Command constructs the worker cobra command.
func (w *Worker) Command() *cobra.Command {
	return &cobra.Command{Use: "worker", Short: "Start background worker"}
}

// Handle runs the worker service.
func (w *Worker) Handle(cmd *cobra.Command) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	tp, err := tracing.GetService("foundation-worker")
	if err != nil {
		return err
	}
	defer tracing.ShutdownWithTimeout(tp)

	server := httpserver.New("worker", "1.0.0", httpserver.WithMetrics(httpserver.MetricsConfig{
		Enabled:         config.Bool("metrics.enabled", true),
		Host:            config.String("metrics.host", "0.0.0.0"),
		Port:            config.Int("metrics.port", 9090),
		ShutdownTimeout: 30 * time.Second,
	}))
	if err := server.Start(); err != nil {
		return err
	}
	defer server.ShutdownWithTimeout()

	// Register all handlers
	for _, h := range w.handlers {
		w.queue.Process(h.JobName(), h.Handle)
		logger.Debug("registered queue handler", "for", h.JobName())
	}
	logger.Info("worker queue handlers", "registered", len(w.handlers))

	logger.Info("worker: starting queue consumer")

	done := make(chan error, 1)
	go func() {
		for {
			runErr := func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("queue consumer panic: %v", r)
					}
				}()
				return w.queue.Run(ctx)
			}()

			if ctx.Err() != nil {
				done <- ctx.Err()
				return
			}
			if runErr == nil {
				done <- nil
				return
			}

			logger.Error("worker: queue consumer crashed, restarting in", "3s", runErr)
			select {
			case <-ctx.Done():
				done <- ctx.Err()
				return
			case <-time.After(3 * time.Second):
			}
		}
	}()

	logger.Info("worker: queue consumer started, waiting for jobs")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-quit:
		logger.Info("worker: received signal, initiating graceful shutdown", "signal", sig)
		cancel()

		logger.Info("worker: waiting for in-flight jobs to complete")
		select {
		case err := <-done:
			if err != nil && err != context.Canceled {
				logger.Error("worker: queue consumer shutdown", "error", err)
			} else {
				logger.Info("worker: queue consumer stopped gracefully")
			}
		case <-time.After(30 * time.Second):
			logger.Warn("worker: queue consumer shutdown timed out after 30s, forcing exit")
		}
	case err := <-done:
		if err != nil && err != context.Canceled {
			return err
		}
		logger.Warn("worker: queue consumer exited unexpectedly")
	}
	logger.Info("worker: service stopped")
	return nil
}
