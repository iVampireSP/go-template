package providers

import (
	servecmd "github.com/iVampireSP/go-template/cmd/serve"
	adminhandler "github.com/iVampireSP/go-template/internal/api/admin/v1/handler"
	adminroute "github.com/iVampireSP/go-template/internal/api/admin/v1/route"
	userhandler "github.com/iVampireSP/go-template/internal/api/user/v1/handler"
	userroute "github.com/iVampireSP/go-template/internal/api/user/v1/route"
	wellknownhandler "github.com/iVampireSP/go-template/internal/api/wellknown/handler"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
)

// RouteServiceProvider registers API handlers, routes, and the serve command.
type RouteServiceProvider struct {
	app *container.Application
}

func NewRouteServiceProvider(app *container.Application) *RouteServiceProvider {
	return &RouteServiceProvider{app: app}
}

func (p *RouteServiceProvider) Register() {
	// Admin API handlers + routes
	p.app.Singleton(adminhandler.NewAuthHandler)
	p.app.Singleton(adminhandler.NewUserHandler)
	p.app.Singleton(adminroute.NewAuth)
	p.app.Singleton(adminroute.NewUser)
	p.app.Singleton(adminroute.NewRouter)

	// User API handlers + routes
	p.app.Singleton(userhandler.NewAuthHandler)
	p.app.Singleton(userhandler.NewProfileHandler)
	p.app.Singleton(userroute.NewAuth)
	p.app.Singleton(userroute.NewProfile)
	p.app.Singleton(userroute.NewRouter)

	// Well-known handlers
	p.app.Singleton(wellknownhandler.NewJWKSHandler)
	p.app.Singleton(wellknownhandler.NewDiscoveryHandler)

	// Serve command
	p.app.AddCommand(servecmd.NewServe(p.app))
}

func (p *RouteServiceProvider) Boot() {}
