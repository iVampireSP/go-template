package providers

import (
	usercronjob "github.com/iVampireSP/go-template/internal/cronjob/user"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
	"github.com/iVampireSP/go-template/pkg/foundation/schedule"
)

// ScheduleServiceProvider registers cronjob implementations.
type ScheduleServiceProvider struct {
	app *container.Application
}

func NewScheduleServiceProvider(app *container.Application) *ScheduleServiceProvider {
	return &ScheduleServiceProvider{app: app}
}

func (p *ScheduleServiceProvider) Register() {
	p.app.Singleton(usercronjob.NewCleanUnverified)
	p.app.Singleton(func(c *usercronjob.CleanUnverified) []schedule.CronJob {
		return []schedule.CronJob{c}
	})
}

func (p *ScheduleServiceProvider) Boot() {}
