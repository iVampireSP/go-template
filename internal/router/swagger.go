package router

import (
	_ "go-template/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type SwaggerRouter struct {
	//config *conf.Config
}

func NewSwaggerRoute() *SwaggerRouter {
	return &SwaggerRouter{}
}

func (a *SwaggerRouter) Register(r *gin.RouterGroup) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
