package middleware

import (
	"go-template/internal/schema"
	"go-template/internal/service/auth"
	"go-template/pkg/consts"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *auth.Service
}

func NewAuthMiddleware(authService *auth.Service) *AuthMiddleware {
	return &AuthMiddleware{
		authService,
	}
}

func (a AuthMiddleware) RequireJWTIDToken(c *gin.Context) {
	user, err := a.authService.GinMiddlewareAuth(schema.JWTIDToken, c)

	if err != nil {
		c.Abort()
		schema.NewResponse(c).Error(err).Status(http.StatusUnauthorized).Send()
		return
	}

	c.Set(consts.AuthMiddlewareKey, user)
	c.Next()
}

func (a AuthMiddleware) RequireJWTAccessToken(c *gin.Context) {
	user, err := a.authService.GinMiddlewareAuth(schema.JWTAccessToken, c)
	if err != nil {
		c.Abort()
		schema.NewResponse(c).Error(err).Status(http.StatusUnauthorized).Send()
		return
	}

	c.Set(consts.AuthMiddlewareKey, user)
	c.Next()
}
