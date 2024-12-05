package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go-template/internal/api/http/response"
	"go-template/internal/base/conf"
	authService "go-template/internal/services/auth"
	authType "go-template/internal/types/auth"
	"go-template/internal/types/constants"
	"net/http"
	"slices"
	"strings"
)

type Auth struct {
	config      *conf.Config
	authService *authService.Service
}

var audienceLength int

func NewAuth(config *conf.Config, authService *authService.Service) *Auth {
	audienceLength = len(config.App.AllowedAudiences)

	return &Auth{
		config,
		authService,
	}
}

func (a *Auth) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var r = response.Ctx(c)
		var err error
		var token = new(authType.User)

		if a.config.Debug.Enabled {
			token, err = a.authService.AuthFromToken(authType.JWTAccessToken, "")
			if err != nil {
				return r.Error(err).Send()
			}

			c.Locals(constants.AuthMiddlewareKey, token)

			return c.Next()
		}

		authorization := c.Get(constants.AuthHeader)

		if authorization == "" {
			return r.Error(constants.ErrJWTFormatError).Send()
		}

		authSplit := strings.Split(authorization, " ")
		if len(authSplit) != 2 {
			return r.Error(constants.ErrJWTFormatError).Send()
		}

		if authSplit[0] != constants.AuthPrefix {
			return r.Error(constants.ErrNotBearerType).Send()
		}

		token, err = a.authService.AuthFromToken(authType.JWTIDToken, authSplit[1])

		if err != nil {
			return r.Error(err).Status(http.StatusUnauthorized).Send()
		}

		if token == nil {
			return r.Error(err).Status(http.StatusUnauthorized).Send()
		}

		if audienceLength > 0 {
			// 检测 aud
			if !slices.Contains(a.config.App.AllowedAudiences, token.Token.Aud) {
				return r.Error(constants.ErrNotValidToken).Send()
			}
		}

		c.Locals(constants.AuthMiddlewareKey, token)

		return c.Next()
	}

}
