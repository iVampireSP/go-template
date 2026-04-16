package middleware

import (
	"net/http"

	"github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/danielgtaylor/huma/v2"
)

// NewEmailVerified 创建邮箱验证检查中间件
func NewEmailVerified(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		op := ctx.Operation()
		if op == nil || !requiresBearerAuth(op) {
			next(ctx)
			return
		}

		model, ok := user.Authenticated(ctx.Context())
		if !ok {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing authorization context")
			return
		}

		if !model.EmailVerified {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "email is not verified",
				&huma.ErrorDetail{
					Location: "header",
					Message:  "email verification required",
					Value:    "EMAIL_NOT_VERIFIED",
				},
			)
			return
		}

		next(ctx)
	}
}
