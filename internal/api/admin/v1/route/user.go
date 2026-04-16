package route

import (
	"github.com/iVampireSP/go-template/internal/api/admin/v1/handler"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/go-chi/chi/v5"
)

type User struct {
	controller *handler.UserHandler
}

func NewUser(controller *handler.UserHandler) *User {
	return &User{
		controller: controller,
	}
}

func (r *User) Register(api httpserver.API, _ chi.Router, security httpserver.Security) {
	httpserver.GET(api, "/v1/users", httpserver.Operation{
		ID:       "adminListUsers",
		Summary:  "列出用户",
		Tags:     []string{"Admin Users"},
		Security: security,
	}, r.controller.List)

	httpserver.GET(api, "/v1/users/{id}", httpserver.Operation{
		ID:       "adminGetUser",
		Summary:  "获取用户详情",
		Tags:     []string{"Admin Users"},
		Security: security,
	}, r.controller.Get)
}
