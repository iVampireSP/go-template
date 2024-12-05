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

func (a *Api) InitApiRouter(r fiber.Router) {
	//r.GET("/ping", a.User.Test)

	r.Get("/ping", a.Middleware.RBAC.RoutePermission(), a.HttpHandler.User.Test)
}

func (a *Api) InitNoAuthApiRouter(r fiber.Router) {

}
