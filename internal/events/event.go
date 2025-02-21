package events

import (
	"context"
	"go-template/internal/dto"

	"go-template/internal/infra/conf"
	"go-template/internal/infra/logger"
	"go-template/internal/services/stream"
)

type Event struct {
	stream *stream.Service
	config *conf.Config
	logger *logger.Logger
}

func NewEvent(stream *stream.Service, config *conf.Config, logger *logger.Logger) *Event {
	return &Event{
		stream: stream,
		config: config,
		logger: logger,
	}
}

type eventMessage struct {
	Name string                 `json:"name"`
	Data dto.PublishableMessage `json:"data,omitempty"`
}

func (e *Event) Fire(ctx context.Context, event dto.PublishableMessage) {
	go func() {
		//em := eventMessage{
		//	Name: event.Name(),
		//	Data: event,
		//}

		//err := e.stream.SendEvent(ctx, e.config.Kafka.Topics.InternalEvents, em)
		//if err != nil {
		//	e.logger.Sugar.Errorf("send event to %s failed: %e", e.config.Kafka.Topics.InternalEvents, err)
		//}
	}()
}
