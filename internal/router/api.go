package router

import (
	"github.com/gofiber/fiber/v2"
	"go-template/internal/api/http"
)

// 两种方法都可以
//type Api struct {
//	User *v1.UserController
//}

type Api struct {
	HttpHandler *http.Handlers
	Middleware  *http.Middleware
}

func NewApiRoute(
	//User *v1.UserController,
	HttpHandler *http.Handlers,
	Middleware *http.Middleware,
) *Api {
	//return &Api{
	//	User,
	//}

	return &Api{
		HttpHandler,
		Middleware,
	}
}

func (a *Api) V1(r fiber.Router) {
	auth := r.Group("/api/v1")
	{
		// 要求认证
		auth.Use(a.Middleware.Auth.Handler())

		// RoutePermission 为权限验证
		//auth.Get("/ping", a.Middleware.RBAC.RoutePermission(), a.HttpHandler.User.Test)

		auth.Get("/ping", a.HttpHandler.User.Test)
	}

	guest := r.Group("/api/v1")
	{
		guest.Get("/guest_ping", a.HttpHandler.User.Test)
	}

}
