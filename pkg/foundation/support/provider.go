package support

// ServiceProvider defines a service provider that registers bindings and boots services.
// This mirrors Laravel's Illuminate\Support\ServiceProvider.
//
// Register() is called during the registration phase to bind services into the container.
// Boot() is called after all providers have been registered.
type ServiceProvider interface {
	Register()
	Boot()
}
