package middleware

import (
	"net/http"

	entuser "github.com/iVampireSP/go-template/ent/user"
	"github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/danielgtaylor/huma/v2"
)

// NewStatusCheck 创建用户状态检查中间件，拦截被封禁或暂停的用户
func NewStatusCheck(api huma.API) func(huma.Context, func(huma.Context)) {
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

		switch model.Status {
		case entuser.StatusSuspended:
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "account is suspended",
				&huma.ErrorDetail{
					Location: "header",
					Message:  "account suspended",
					Value:    "ACCOUNT_SUSPENDED",
				},
			)
			return
		case entuser.StatusBanned:
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "account is banned",
				&huma.ErrorDetail{
					Location: "header",
					Message:  "account banned",
					Value:    "ACCOUNT_BANNED",
				},
			)
			return
		}

		next(ctx)
	}
}
