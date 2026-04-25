package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iVampireSP/go-template/pkg/foundation/bus"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
	"github.com/iVampireSP/go-template/pkg/foundation/tracing"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/spf13/cobra"
)

// EventBus holds EventBus service dependencies.
type EventBus struct {
	app       *container.Application
	bus       *bus.Bus
	listeners []bus.Listener
	metrics   httpserver.MetricsConfig
}

// NewEventBus declares EventBus command dependencies.
func NewEventBus(app *container.Application) *EventBus {
	return &EventBus{app: app}
}

// Command constructs the eventbus cobra command.
func (e *EventBus) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "eventbus", Short: "Start event bus consumer",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return e.app.Invoke(func(b *bus.Bus, l []bus.Listener, m httpserver.MetricsConfig) {
				e.bus = b
				e.listeners = l
				e.metrics = m
			})
		},
		RunE: func(c *cobra.Command, _ []string) error { return e.Handle(c) },
	}

	topic := &cobra.Command{Use: "topic", Short: "Event bus topic management"}
	create := &cobra.Command{Use: "create", Short: "Create configured event bus topics",
		RunE: func(c *cobra.Command, args []string) error { return e.CreateTopics(c, args) }}
	topic.AddCommand(create)
	cmd.AddCommand(topic)

	return cmd
}

// Handle runs the eventbus service.
func (e *EventBus) Handle(cmd *cobra.Command) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	tp, err := tracing.GetService("foundation-eventbus")
	if err != nil {
		return err
	}
	defer tracing.ShutdownWithTimeout(tp)

	server := httpserver.New("eventbus", "1.0.0", httpserver.WithMetrics(e.metrics))
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
