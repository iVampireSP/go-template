package middleware

import (
	"github.com/labstack/echo/v4"
	"go-template/internal/base/conf"
	"go-template/internal/handler/http/response"
	"go-template/internal/schema"
	"go-template/internal/service/auth"
	"go-template/pkg/consts"
	"net/http"
	"slices"
	"strings"
)

type AuthMiddleware struct {
	config      *conf.Config
	authService *auth.Service
}

var audienceLength int

func NewAuthMiddleware(config *conf.Config, authService *auth.Service) *AuthMiddleware {
	audienceLength = len(config.App.AllowedAudiences)

	return &AuthMiddleware{
		config,
		authService,
	}
}

func (a *AuthMiddleware) Handler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var r = response.Ctx(c)
			var err error
			var token *schema.User

			if a.config.Debug.Enabled {
				token, err = a.authService.AuthFromToken(schema.JWTAccessToken, "")
				if err != nil {
					return r.Error(err).Send()
				}

				c.Set(consts.AuthMiddlewareKey, token)

				return next(c)
			}

			authorization := c.Request().Header.Get(consts.AuthHeader)

			if authorization == "" {
				return r.Error(consts.ErrJWTFormatError).Send()
			}

			authSplit := strings.Split(authorization, " ")
			if len(authSplit) != 2 {
				return r.Error(consts.ErrJWTFormatError).Send()
			}

			if authSplit[0] != consts.AuthPrefix {
				return r.Error(consts.ErrNotBearerType).Send()
			}

			token, err = a.authService.AuthFromToken(schema.JWTIDToken, authSplit[1])

			if err != nil {
				return r.Error(err).Status(http.StatusUnauthorized).Send()
			}

			if token == nil {
				return r.Error(err).Status(http.StatusUnauthorized).Send()
			}

			if audienceLength > 0 {
				// 检测 aud
				if !slices.Contains(a.config.App.AllowedAudiences, token.Token.Aud) {
					return r.Error(consts.ErrNotValidToken).Send()
				}
			}

			c.Set(consts.AuthMiddlewareKey, token)

			return next(c)
		}
	}

}
