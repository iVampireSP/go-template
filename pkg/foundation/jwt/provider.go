package jwt

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
	p.app.Singleton(NewJWT)
}

func (p *Provider) Boot() {}

// NewDefaultConfig returns a JWT Config populated from the application config.
func NewDefaultConfig() Config {
	return Config{
		KeyName:   config.String("jwt.key", "rsa"),
		Issuer:    config.String("discovery.issuer", "http://localhost"),
		ExpiresIn: config.Int("jwt.expires_in", 86400),
	}
}
