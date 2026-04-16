package route

import (
	adminmw "github.com/iVampireSP/go-template/internal/api/admin/middleware"
	"github.com/iVampireSP/go-template/internal/service/identity/admin"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	adminSvc *admin.Admin
	auth     *Auth
	user     *User
}

func NewRouter(
	adminSvc *admin.Admin,
	auth *Auth,
	user *User,
) *Router {
	return &Router{
		adminSvc: adminSvc,
		auth:     auth,
		user:     user,
	}
}

func (r *Router) Register(api httpserver.API, router chi.Router) {
	publicGroup := api.Group()
	authGroup := api.Group()
	authGroup.UseMiddleware(adminmw.NewHumaAuth(api.Huma(), r.adminSvc))

	security := httpserver.Security{{"bearer": {}}}
	r.auth.Register(publicGroup, authGroup, router, security)
	r.user.Register(authGroup, router, security)
}
