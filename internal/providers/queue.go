package providers

import (
	"github.com/iVampireSP/go-template/pkg/foundation/container"
	"github.com/iVampireSP/go-template/pkg/foundation/queue/job"
	mailhandler "github.com/iVampireSP/go-template/pkg/foundation/queue/job/mail"
)

// QueueServiceProvider registers job handlers.
type QueueServiceProvider struct {
	app *container.Application
}

func NewQueueServiceProvider(app *container.Application) *QueueServiceProvider {
	return &QueueServiceProvider{app: app}
}

func (p *QueueServiceProvider) Register() {
	p.app.Singleton(mailhandler.NewHandler)
	p.app.Singleton(func(h *mailhandler.Handler) []job.Handler {
		return []job.Handler{h}
	})
}

func (p *QueueServiceProvider) Boot() {}
