package middleware

import (
	"net/http"
	"strings"

	"github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/danielgtaylor/huma/v2"
)

// NewHumaAuth creates a Huma authentication middleware (JWT only).
func NewHumaAuth(api huma.API, userSvc *user.User) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		op := ctx.Operation()
		if op == nil || !requiresBearerAuth(op) {
			next(ctx)
			return
		}

		authHeader := ctx.Header("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing or invalid authorization")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)
		if token == "" {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing or invalid authorization")
			return
		}

		userID, err := userSvc.ResolveToken(ctx.Context(), token)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "authentication failed")
			return
		}

		u, err := userSvc.GetByIDCached(ctx.Context(), userID)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "user not found")
			return
		}

		enrichedCtx := user.WithAuth(ctx.Context(), u, nil)
		next(huma.WithContext(ctx, enrichedCtx))
	}
}

// requiresBearerAuth checks if the operation requires bearer auth.
func requiresBearerAuth(op *huma.Operation) bool {
	for _, sec := range op.Security {
		if _, ok := sec["bearer"]; ok {
			return true
		}
	}
	return false
}
