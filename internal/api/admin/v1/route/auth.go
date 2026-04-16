package route

import (
	"github.com/iVampireSP/go-template/internal/api/admin/v1/handler"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/go-chi/chi/v5"
)

type Auth struct {
	controller *handler.AuthHandler
}

func NewAuth(controller *handler.AuthHandler) *Auth {
	return &Auth{
		controller: controller,
	}
}

func (r *Auth) Register(public httpserver.API, auth httpserver.API, _ chi.Router, security httpserver.Security) {
	httpserver.POST(public, "/v1/auth/login", httpserver.Operation{
		ID:      "adminLogin",
		Summary: "管理员登录",
		Tags:    []string{"Admin Auth"},
	}, r.controller.Login)

	httpserver.GET(auth, "/v1/auth/me", httpserver.Operation{
		ID:       "adminMe",
		Summary:  "获取当前管理员信息",
		Tags:     []string{"Admin Auth"},
		Security: security,
	}, r.controller.Me)
}
