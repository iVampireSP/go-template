package container

import (
	"fmt"

	"go.uber.org/dig"
)

// Container wraps dig.Container with a Laravel-inspired API.
type Container struct {
	scope *dig.Container
}

// NewContainer creates a new dependency injection container.
func NewContainer(opts ...dig.Option) *Container {
	return &Container{scope: dig.New(opts...)}
}

// Singleton registers a constructor function. The constructor is called at most once
// (on first need), and the result is cached for subsequent requests.
// This is equivalent to Laravel's $app->singleton().
func (c *Container) Singleton(constructor any, opts ...dig.ProvideOption) error {
	return c.scope.Provide(constructor, opts...)
}

// Instance registers a pre-built value into the container.
// This is equivalent to Laravel's $app->instance().
// The value's concrete type is used as the binding key.
func (c *Container) Instance(value any) error {
	return c.scope.Provide(func() any { return value })
}

// Invoke resolves dependencies from the container and calls the given function.
// All function parameters are injected from the container.
// This is equivalent to Laravel's $app->call().
func (c *Container) Invoke(fn any, opts ...dig.InvokeOption) error {
	return c.scope.Invoke(fn, opts...)
}

// Scope returns the underlying dig.Container for advanced use cases.
func (c *Container) Scope() *dig.Container {
	return c.scope
}

// Make resolves a value of type T from the container (Go 1.26 generics).
// This is equivalent to Laravel's $app->make(T::class).
func Make[T any](c *Container) (T, error) {
	var result T
	err := c.scope.Invoke(func(v T) {
		result = v
	})
	return result, err
}

// MustMake resolves a value of type T from the container, panicking on failure.
func MustMake[T any](c *Container) T {
	v, err := Make[T](c)
	if err != nil {
		var zero T
		panic(fmt.Sprintf("container: failed to resolve %T: %v", zero, err))
	}
	return v
}
