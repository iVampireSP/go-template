package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "go-template/docs"
)

type SwaggerRouter struct {
	//config *conf.Config
}

func NewSwaggerRoute() *SwaggerRouter {
	return &SwaggerRouter{}
}

func (a *SwaggerRouter) Register(r fiber.Router) {
	r.Get("/swagger/*", swagger.HandlerDefault)
}
