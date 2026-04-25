package logger

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
	p.app.Singleton(New)
}

func (p *Provider) Boot() {}

// NewDefaultConfig returns a logger Config populated from the application config.
func NewDefaultConfig() Config {
	return Config{
		Level: config.String("log.level", "info"),
		Debug: config.Bool("app.debug", false),
	}
}
