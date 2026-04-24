package bus

import (
	"github.com/iVampireSP/go-template/internal/infra/config"
	foundationbus "github.com/iVampireSP/go-template/pkg/foundation/bus"
)

// Re-export foundation types for backward compatibility.
type Bus = foundationbus.Bus
type Event = foundationbus.Event
type Handler = foundationbus.Handler
type Listener = foundationbus.Listener
type Middleware = foundationbus.Middleware
type TopicID = foundationbus.TopicID
type Envelope = foundationbus.Envelope
type ConsumerOption = foundationbus.ConsumerOption

const TopicDefault = foundationbus.TopicDefault

var (
	ConsumerModeShared    = foundationbus.ConsumerModeShared
	ConsumerModeBroadcast = foundationbus.ConsumerModeBroadcast
	WithConsumerMode      = foundationbus.WithConsumerMode
	WithGroupID           = foundationbus.WithGroupID
	ContextWithEnvelope   = foundationbus.ContextWithEnvelope
	EnvelopeFromContext   = foundationbus.EnvelopeFromContext
)

func NewBus() *Bus {
	var cfg foundationbus.Config
	if err := config.Unmarshal("bus", &cfg); err != nil {
		panic(err)
	}
	b, err := foundationbus.NewBus(cfg)
	if err != nil {
		panic(err)
	}
	return b
}
