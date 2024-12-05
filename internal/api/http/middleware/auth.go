package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go-template/internal/api/http/response"
	"go-template/internal/base/conf"
	"go-template/internal/consts"
	"go-template/internal/pkg/user"
	"go-template/internal/schema"
	"go-template/internal/service/auth"
	"net/http"
	"slices"
	"strings"
)

type Auth struct {
	config      *conf.Config
	authService *auth.Service
}

var audienceLength int

func NewAuth(config *conf.Config, authService *auth.Service) *Auth {
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
		var token = new(user.User)

		if a.config.Debug.Enabled {
			token, err = a.authService.AuthFromToken(schema.JWTAccessToken, "")
			if err != nil {
				return r.Error(err).Send()
			}

			c.Locals(consts.AuthMiddlewareKey, token)

			return c.Next()
		}

		authorization := c.Get(consts.AuthHeader)

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

		c.Locals(consts.AuthMiddlewareKey, token)

		return c.Next()
	}

}
