package bus

import (
	"context"
	"time"

	"github.com/iVampireSP/go-template/pkg/cerr"
	"github.com/iVampireSP/go-template/pkg/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var (
	ErrEncodeMessage = cerr.Internal("failed to encode event").WithCode("BUS_ENCODE_EVENT")
	ErrPublish       = cerr.ServiceUnavailable("failed to publish event").WithCode("BUS_PUBLISH_EVENT")
	ErrEventTopic    = cerr.BadRequest("event topic is required").WithCode("BUS_EVENT_TOPIC_REQUIRED")
)

type publisher struct {
	kafka *kafkaClient
}

func newPublisher(kafka *kafkaClient) *publisher {
	return &publisher{
		kafka: kafka,
	}
}

func (p *publisher) publish(ctx context.Context, event Event, topic string) error {
	env, err := newEnvelope(event.Name(), event.Key(), event)
	if err != nil {
		return ErrEncodeMessage.WithCause(err)
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(env.Metadata))

	if topic == "" {
		return ErrEventTopic
	}

	data, err := env.Encode()
	if err != nil {
		return ErrEncodeMessage.WithCause(err)
	}

	msg := kafkaMessage{Value: data}
	if env.Key != "" {
		msg.Key = []byte(env.Key)
	}
	// Detach from caller's cancel (e.g. HTTP request) but keep values (OTel trace).
	writeCtx, writeCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer writeCancel()

	if err := p.kafka.WriteTo(writeCtx, topic, msg); err != nil {
		return ErrPublish.WithCause(err)
	}
	return nil
}

func (p *publisher) publishDLQ(ctx context.Context, topic string, key string, dlq *DLQEnvelope) error {
	data, err := json.Marshal(dlq)
	if err != nil {
		return ErrEncodeMessage.WithCause(err)
	}
	msg := kafkaMessage{Value: data}
	if key != "" {
		msg.Key = []byte(key)
	}
	dlqCtx, dlqCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer dlqCancel()

	if err := p.kafka.WriteTo(dlqCtx, topic, msg); err != nil {
		return ErrPublish.WithCause(err)
	}
	return nil
}
