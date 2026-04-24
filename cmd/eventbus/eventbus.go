package eventbus

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iVampireSP/go-template/internal/infra/bus"
	"github.com/iVampireSP/go-template/internal/infra/config"
	"github.com/iVampireSP/go-template/internal/infra/tracing"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/spf13/cobra"
)

// EventBus holds EventBus service dependencies.
type EventBus struct {
	bus       *bus.Bus
	listeners []bus.Listener
}

// NewEventBus declares EventBus command dependencies.
func NewEventBus(b *bus.Bus, listeners []bus.Listener) *EventBus {
	return &EventBus{
		bus:       b,
		listeners: listeners,
	}
}

// Command constructs the eventbus cobra command.
func (e *EventBus) Command() *cobra.Command {
	cmd := &cobra.Command{Use: "eventbus", Short: "Start event bus consumer"}

	topic := &cobra.Command{Use: "topic", Short: "Event bus topic management"}
	create := &cobra.Command{Use: "create", Short: "Create configured event bus topics", RunE: e.CreateTopics}
	topic.AddCommand(create)
	cmd.AddCommand(topic)

	return cmd
}

// Handle runs the eventbus service.
func (e *EventBus) Handle(cmd *cobra.Command) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	tp, err := tracing.GetService("app-eventbus")
	if err != nil {
		return err
	}
	defer tracing.ShutdownWithTimeout(tp)

	server := httpserver.New("eventbus", "1.0.0", httpserver.WithMetrics(httpserver.MetricsConfig{
		Enabled:         config.Bool("metrics.enabled", true),
		Host:            config.String("metrics.host", "0.0.0.0"),
		Port:            config.Int("metrics.port", 9090),
		ShutdownTimeout: 30 * time.Second,
	}))
	if err := server.Start(); err != nil {
		return err
	}
	defer server.ShutdownWithTimeout()

	e.bus.Register(e.listeners...)
	e.bus.EnableDLQ(func(topic string) string {
		return topic + ".dlq"
	})
	if err := e.bus.Init("eventbus"); err != nil {
		return err
	}
	logger.Info("eventbus listeners", "registered", len(e.listeners))
	logger.Info("eventbus: starting consumer")

	done := make(chan error, 1)
	go func() {
		done <- e.bus.Run(ctx)
	}()

	logger.Info("eventbus: consumer started, waiting for messages")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("eventbus: received signal, initiating graceful shutdown", "signal", sig)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			logger.Error("eventbus: consumer shutdown", "error", err)
		} else {
			logger.Info("eventbus: consumer stopped gracefully")
		}
	case <-time.After(30 * time.Second):
		logger.Warn("eventbus: consumer shutdown timed out after 30s, forcing exit")
	}
	logger.Info("eventbus: service stopped")
	return nil
}

func (e *EventBus) CreateTopics(cmd *cobra.Command, _ []string) error {
	if err := e.bus.CreateAllTopics(cmd.Context(), true); err != nil {
		return err
	}

	fmt.Println("eventbus topics created")
	return nil
}
