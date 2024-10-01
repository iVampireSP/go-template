package router

import "github.com/google/wire"

// ProviderSetRouter is providers.
var ProviderSetRouter = wire.NewSet(
	NewApiRoute,
	NewSwaggerRoute,
)
