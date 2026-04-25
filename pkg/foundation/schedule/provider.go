package schedule

import "github.com/iVampireSP/go-template/pkg/foundation/container"

type Provider struct {
	app *container.Application
}

func NewProvider(app *container.Application) *Provider {
	return &Provider{app: app}
}

func (p *Provider) Register() {}

func (p *Provider) Boot() {}
