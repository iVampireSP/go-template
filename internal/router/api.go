package router

import (
	v1 "go-template/internal/api/v1"

	"github.com/gin-gonic/gin"
)

type Api struct {
	User *v1.UserController
}

func NewApiRoute(
	User *v1.UserController,
) *Api {
	return &Api{
		User,
	}
}

func (a *Api) InitApiRouter(r *gin.RouterGroup) {
	r.GET("/ping", a.User.Test)

}

func (a *Api) InitNoAuthApiRouter(r *gin.RouterGroup) {

}
