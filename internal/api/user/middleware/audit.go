package middleware

import (
	"github.com/danielgtaylor/huma/v2"
)

// AuditRequestInfoMiddleware is a placeholder for request audit logging.
func AuditRequestInfoMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		next(ctx)
	}
}
