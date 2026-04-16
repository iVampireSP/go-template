package middleware

import (
	"net/http"
	"strings"

	"github.com/iVampireSP/go-template/internal/service/identity/admin"
	"github.com/iVampireSP/go-template/pkg/cerr"
	"github.com/danielgtaylor/huma/v2"
)

// NewChiAuth 创建管理员 Chi 认证中间件
func NewChiAuth(svc *admin.Admin) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r.Header.Get("Authorization"))
			if token == "" && isWebSocketUpgrade(r) {
				token = r.URL.Query().Get("access_token")
				if token == "" {
					token = r.URL.Query().Get("token")
				}
			}
			if token != "" {
				ctx := r.Context()
				id, scopes, err := svc.ResolveToken(ctx, token)
				if err == nil {
					a, loadErr := svc.GetByIDCached(ctx, id)
					if loadErr == nil {
						ctx = admin.WithAuth(ctx, a, scopes)
						r = r.WithContext(ctx)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NewHumaAuth 创建管理员 Huma 认证中间件
func NewHumaAuth(api huma.API, svc *admin.Admin) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		op := ctx.Operation()
		if op == nil || !requiresBearerAuth(op) {
			next(ctx)
			return
		}

		token := extractBearerToken(ctx.Header("Authorization"))
		if token == "" {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing authorization header")
			return
		}

		id, scopes, err := svc.ResolveToken(ctx.Context(), token)
		if err != nil {
			status, msg := classifyAuthError(err)
			_ = huma.WriteErr(api, ctx, status, msg)
			return
		}

		a, loadErr := svc.GetByIDCached(ctx.Context(), id)
		if loadErr != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "admin not found")
			return
		}

		enrichedCtx := admin.WithAuth(ctx.Context(), a, scopes)
		next(huma.WithContext(ctx, enrichedCtx))
	}
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}

func isWebSocketUpgrade(r *http.Request) bool {
	if r == nil {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket") {
		return false
	}
	connection := strings.ToLower(r.Header.Get("Connection"))
	return strings.Contains(connection, "upgrade")
}

func requiresBearerAuth(op *huma.Operation) bool {
	for _, sec := range op.Security {
		if _, ok := sec["bearer"]; ok {
			return true
		}
	}
	return false
}

func classifyAuthError(err error) (int, string) {
	if e, ok := cerr.As(err); ok {
		return e.Status, e.Message
	}
	return http.StatusInternalServerError, "internal server error"
}
