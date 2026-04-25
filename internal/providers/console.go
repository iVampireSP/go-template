package providers

import (
	buscommand "github.com/iVampireSP/go-template/pkg/foundation/bus/command"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
	ormcommand "github.com/iVampireSP/go-template/pkg/foundation/orm/command"
	queuecommand "github.com/iVampireSP/go-template/pkg/foundation/queue/command"
	schedulecommand "github.com/iVampireSP/go-template/pkg/foundation/schedule/command"
)

// ConsoleServiceProvider registers foundation framework commands.
// These are registered here (not in foundation providers) to avoid import cycles
// between foundation packages and their command subpackages.
type ConsoleServiceProvider struct {
	app *container.Application
}

func NewConsoleServiceProvider(app *container.Application) *ConsoleServiceProvider {
	return &ConsoleServiceProvider{app: app}
}

func (p *ConsoleServiceProvider) Register() {
	p.app.AddCommand(
		ormcommand.NewMigrate(p.app),
		queuecommand.NewWorker(p.app),
		buscommand.NewEventBus(p.app),
		schedulecommand.NewScheduler(p.app),
	)
}

func (p *ConsoleServiceProvider) Boot() {}
