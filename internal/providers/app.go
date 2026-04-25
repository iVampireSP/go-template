package providers

import (
	usercmd "github.com/iVampireSP/go-template/cmd/user"
	versioncmd "github.com/iVampireSP/go-template/cmd/version"
	"github.com/iVampireSP/go-template/internal/service/identity/admin"
	"github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
)

// AppServiceProvider registers core business services and commands.
type AppServiceProvider struct {
	app *container.Application
}

func NewAppServiceProvider(app *container.Application) *AppServiceProvider {
	return &AppServiceProvider{app: app}
}

func (p *AppServiceProvider) Register() {
	// Business services
	p.app.Singleton(admin.NewAdmin)
	p.app.Singleton(user.NewUser)

	// Business commands
	p.app.AddCommand(
		usercmd.NewUser(p.app),
		versioncmd.NewVersion(),
	)
}

func (p *AppServiceProvider) Boot() {}
