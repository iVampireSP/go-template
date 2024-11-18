package router

import (
	"github.com/labstack/echo/v4"
	_ "go-template/docs"

	echoSwagger "github.com/swaggo/echo-swagger"
)

type SwaggerRouter struct {
	//config *conf.Config
}

func NewSwaggerRoute() *SwaggerRouter {
	return &SwaggerRouter{}
}

func (a *SwaggerRouter) Register(e *echo.Group) {
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}
