package bus

import "context"

type TopicID string

const (
	TopicDefault TopicID = "default"
)

type Event interface {
	Name() string
	Key() string
	Topic() TopicID
}

type Handler func(ctx context.Context, payload []byte) error

type Middleware func(next Handler) Handler
