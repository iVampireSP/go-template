package router

import (
	"github.com/labstack/echo/v4"
	"go-template/internal/handler/http"
)

// 两种方法都可以
//type Api struct {
//	User *v1.UserController
//}

type Api struct {
	HttpHandler *http.Handlers
}

func NewApiRoute(
	//User *v1.UserController,
	HttpHandler *http.Handlers,
) *Api {
	//return &Api{
	//	User,
	//}

	return &Api{
		HttpHandler,
	}
}

func (a *Api) InitApiRouter(r *echo.Group) {
	//r.GET("/ping", a.User.Test)

	r.GET("/ping", a.HttpHandler.User.Test)
}

func (a *Api) InitNoAuthApiRouter(r *echo.Group) {

}
