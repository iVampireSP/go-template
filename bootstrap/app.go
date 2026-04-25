package bootstrap

import "github.com/iVampireSP/go-template/pkg/foundation/container"

// CreateApplication creates and configures the application with all providers.
func CreateApplication() *container.Application {
	app := container.NewApplication()
	app.Register(Providers()...)
	return app
}
