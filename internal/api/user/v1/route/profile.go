package route

import (
	"net/http"

	"github.com/iVampireSP/go-template/internal/api/user/v1/handler"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/go-chi/chi/v5"
)

type Profile struct {
	controller *handler.ProfileHandler
}

func NewProfile(controller *handler.ProfileHandler) *Profile {
	return &Profile{
		controller: controller,
	}
}

func (r *Profile) Register(auth httpserver.API, verified httpserver.API, _ chi.Router, security httpserver.Security) {
	httpserver.GET(auth, "/v1/users/me", httpserver.Operation{
		ID:       "getCurrentUser",
		Summary:  "获取当前用户信息",
		Tags:     []string{"Profile"},
		Security: security,
	}, r.controller.Me)

	httpserver.PATCH(verified, "/v1/users/me", httpserver.Operation{
		ID:       "updateCurrentUser",
		Summary:  "更新当前用户资料",
		Tags:     []string{"Profile"},
		Security: security,
	}, r.controller.UpdateMe)

	httpserver.POST(verified, "/v1/users/me/password", httpserver.Operation{
		ID:            "updateCurrentUserPassword",
		Summary:       "修改当前用户密码",
		Tags:          []string{"Profile"},
		DefaultStatus: http.StatusNoContent,
		Security:      security,
	}, r.controller.UpdatePassword)
}
