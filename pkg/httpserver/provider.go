package httpserver

import (
	"time"

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
	p.app.Singleton(NewDefaultMetricsConfig)
}

func (p *Provider) Boot() {}

func NewDefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Enabled:         config.Bool("metrics.enabled", true),
		Host:            config.String("metrics.host", "0.0.0.0"),
		Port:            config.Int("metrics.port", 9090),
		ShutdownTimeout: 30 * time.Second,
	}
}
