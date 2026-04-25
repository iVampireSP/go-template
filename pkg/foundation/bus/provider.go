package bus

import (
	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
)

type Provider struct {
	app *container.Application
}

func NewProvider(app *container.Application) *Provider {
	return &Provider{app: app}
}

func (p *Provider) Register() {
	p.app.Singleton(NewDefaultConfig)
	p.app.Singleton(NewBus)
}

func (p *Provider) Boot() {}

// NewDefaultConfig returns a bus Config populated from the application config.
func NewDefaultConfig() Config {
	var cfg Config
	if err := config.Unmarshal("bus", &cfg); err != nil {
		panic(err)
	}
	return cfg
}
