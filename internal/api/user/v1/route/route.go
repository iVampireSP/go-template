package route

import (
	usermw "github.com/iVampireSP/go-template/internal/api/user/middleware"
	usersvc "github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	userSvc *usersvc.User
	auth    *Auth
	profile *Profile
}

func NewRouter(
	userSvc *usersvc.User,
	auth *Auth,
	profile *Profile,
) *Router {
	return &Router{
		userSvc: userSvc,
		auth:    auth,
		profile: profile,
	}
}

func (r *Router) Register(api httpserver.API, router chi.Router) {
	publicGroup := api.Group()
	authGroup := api.Group()
	authGroup.UseMiddleware(usermw.NewHumaAuth(api.Huma(), r.userSvc))
	authGroup.UseMiddleware(usermw.AuditRequestInfoMiddleware(api.Huma()))
	verifiedGroup := authGroup.Group()
	verifiedGroup.UseMiddleware(usermw.NewStatusCheck(api.Huma()))
	verifiedGroup.UseMiddleware(usermw.NewEmailVerified(api.Huma()))

	security := httpserver.Security{{"bearer": {}}}
	r.auth.Register(publicGroup, authGroup, router, security)
	r.profile.Register(authGroup, verifiedGroup, router, security)
}
