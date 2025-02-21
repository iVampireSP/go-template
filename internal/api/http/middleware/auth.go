package middleware

import (
	"go-template/internal/infra/conf"
	authService "go-template/internal/services/auth"
	"go-template/internal/types/constants"
	"go-template/internal/types/dto"
	"go-template/internal/types/errs"
	authType "go-template/internal/types/user"
	"net/http"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
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
		var r = dto.Ctx(c)
		var err error
		var token = new(authType.User)

		if a.config.Debug.Enabled {
			token, err = a.authService.AuthFromToken(constants.JwtTokenTypeAccessToken, "")
			if err != nil {
				return r.Error(err).Send()
			}

			c.Locals(constants.AuthMiddlewareKey, token)

			return c.Next()
		}

		authorization := c.Get(constants.AuthHeader)

		if authorization == "" {
			return r.Error(errs.JWTFormatError).Send()
		}

		authSplit := strings.Split(authorization, " ")
		if len(authSplit) != 2 {
			return r.Error(errs.JWTFormatError).Send()
		}

		if authSplit[0] != constants.AuthPrefix {
			return r.Error(errs.NotBearerType).Send()
		}

		token, err = a.authService.AuthFromToken(constants.JwtTokenTypeIDToken, authSplit[1])

		if err != nil {
			return r.Error(err).Status(http.StatusUnauthorized).Send()
		}

		if token == nil {
			return r.Error(err).Status(http.StatusUnauthorized).Send()
		}

		if audienceLength > 0 {
			// 检测 aud
			if !slices.Contains(a.config.App.AllowedAudiences, token.Token.Aud) {
				return r.Error(errs.NotValidToken).Send()
			}
		}

		c.Locals(constants.AuthMiddlewareKey, token)

		return c.Next()
	}

}
