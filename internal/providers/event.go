package providers

import (
	examplelistener "github.com/iVampireSP/go-template/internal/listener/example"
	"github.com/iVampireSP/go-template/pkg/foundation/bus"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
)

// EventServiceProvider registers event listeners.
type EventServiceProvider struct {
	app *container.Application
}

func NewEventServiceProvider(app *container.Application) *EventServiceProvider {
	return &EventServiceProvider{app: app}
}

func (p *EventServiceProvider) Register() {
	p.app.Singleton(examplelistener.NewListener)
	p.app.Singleton(func(l *examplelistener.Listener) []bus.Listener {
		return []bus.Listener{l}
	})
}

func (p *EventServiceProvider) Boot() {}
