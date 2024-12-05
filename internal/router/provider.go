package router

import "github.com/google/wire"

// Provide is providers.
var Provide = wire.NewSet(
	NewApiRoute,
	NewSwaggerRoute,
)
