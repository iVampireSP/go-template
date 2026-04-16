package route

import (
	"github.com/iVampireSP/go-template/internal/api/user/v1/handler"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/go-chi/chi/v5"
)

type Auth struct {
	controller *handler.AuthHandler
}

func NewAuth(controller *handler.AuthHandler) *Auth {
	return &Auth{controller: controller}
}

func (r *Auth) Register(public httpserver.API, auth httpserver.API, _ chi.Router, security httpserver.Security) {
	httpserver.POST(public, "/v1/auth/login", httpserver.Operation{
		ID:      "login",
		Summary: "用户登录",
		Tags:    []string{"Auth"},
	}, r.controller.Login)

	httpserver.POST(public, "/v1/auth/register", httpserver.Operation{
		ID:      "register",
		Summary: "用户注册",
		Tags:    []string{"Auth"},
	}, r.controller.Register)
}
