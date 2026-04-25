package container

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/iVampireSP/go-template/pkg/foundation/console"
	"github.com/iVampireSP/go-template/pkg/foundation/support"
	"github.com/spf13/cobra"
)

// Application manages the full lifecycle of service providers and the DI container.
// It follows Laravel's Application pattern: Register → Boot → Run.
type Application struct {
	*Container
	providers   []support.ServiceProvider
	commands    []console.ConsoleCommand
	shutdownFns []func() error
	booted      bool
}

// NewApplication creates a new application with a fresh container.
func NewApplication() *Application {
	return &Application{
		Container: NewContainer(),
	}
}

// Register accepts provider constructor functions and automatically injects *Application.
// Each constructor is called with the Application instance, and the returned provider's
// Register() method is invoked immediately.
//
// Usage:
//
//	app.Register(orm.NewProvider, cache.NewProvider, ...)
//
// Constructor signature: func(app *Application) *XxxProvider
func (app *Application) Register(constructors ...any) {
	for _, ctor := range constructors {
		fn := reflect.ValueOf(ctor)
		if fn.Kind() != reflect.Func {
			panic(fmt.Sprintf("container: Register expects function constructors, got %T", ctor))
		}

		result := fn.Call([]reflect.Value{reflect.ValueOf(app)})
		p, ok := result[0].Interface().(support.ServiceProvider)
		if !ok {
			panic(fmt.Sprintf("container: constructor %T must return support.ServiceProvider", ctor))
		}

		p.Register()
		app.providers = append(app.providers, p)
	}
}

// Boot calls Boot() on all registered providers.
// This is Phase 2 of the lifecycle, called after all providers are registered.
func (app *Application) Boot() error {
	if app.booted {
		return nil
	}
	for _, p := range app.providers {
		p.Boot()
	}
	app.booted = true
	return nil
}

// AddCommand registers console commands with the application.
// Providers call this in their Register() method.
func (app *Application) AddCommand(cmds ...console.ConsoleCommand) {
	app.commands = append(app.commands, cmds...)
}

// OnShutdown registers a cleanup callback to be called during shutdown.
// Callbacks are executed in reverse registration order.
func (app *Application) OnShutdown(fn func() error) {
	app.shutdownFns = append(app.shutdownFns, fn)
}

// Shutdown executes all registered cleanup callbacks in reverse order.
func (app *Application) Shutdown() error {
	var errs []error
	for i := len(app.shutdownFns) - 1; i >= 0; i-- {
		if err := app.shutdownFns[i](); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Run executes the full application lifecycle: Boot → collect commands → cobra.Execute.
func (app *Application) Run(use, short string) error {
	if err := app.Boot(); err != nil {
		return err
	}

	root := &cobra.Command{
		Use:   use,
		Short: short,
	}

	// Add all registered commands
	for _, cc := range app.commands {
		root.AddCommand(cc.Command())
	}

	// Shutdown hook
	root.PersistentPostRunE = func(_ *cobra.Command, _ []string) error {
		return app.Shutdown()
	}

	if err := root.Execute(); err != nil {
		_ = app.Shutdown()
		os.Exit(1)
	}
	return nil
}
