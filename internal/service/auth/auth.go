package auth

import (
	"context"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"go-template/internal/schema"
	"go-template/internal/service/jwks"
	"go-template/pkg/consts"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type Service struct {
	config *conf.Config
	jwks   *jwks.JWKS
	logger *logger.Logger
}

func NewAuthService(config *conf.Config, jwks *jwks.JWKS, logger *logger.Logger) *Service {
	return &Service{
		config: config,
		jwks:   jwks,
		logger: logger,
	}
}

func (a *Service) GinMiddlewareAuth(tokenType schema.JWTTokenTypes, c *gin.Context) (*schema.User, error) {
	if a.config.Debug.Enabled {
		return a.parseUserJWT(tokenType, "")
	}

	authorization := c.Request.Header.Get(consts.AuthHeader)

	if authorization == "" {
		return nil, consts.ErrJWTFormatError
	}

	authSplit := strings.Split(authorization, " ")
	if len(authSplit) != 2 {
		return nil, consts.ErrJWTFormatError
	}

	if authSplit[0] != consts.AuthPrefix {
		return nil, consts.ErrNotBearerType
	}

	return a.parseUserJWT(tokenType, authSplit[1])
}

func (a *Service) GetUserFromIdToken(idToken string) (*schema.User, error) {
	return a.parseUserJWT(schema.JWTIDToken, idToken)
}

func (a *Service) GinUser(c *gin.Context) *schema.User {
	user, _ := c.Get(consts.AuthMiddlewareKey)
	return user.(*schema.User)
}

func (a *Service) GetUserId(ctx context.Context) schema.UserId {
	user := a.GetUser(ctx)

	return user.Token.Sub
}

func (a *Service) GetUser(ctx context.Context) *schema.User {
	user := ctx.Value(consts.AuthMiddlewareKey)

	return user.(*schema.User)
}

func (a *Service) SetUser(ctx context.Context, user *schema.User) context.Context {
	context.WithValue(ctx, consts.AuthMiddlewareKey, user)
	return context.WithValue(ctx, consts.AuthMiddlewareKey, user)
}

func (a *Service) parseUserJWT(tokenType schema.JWTTokenTypes, jwtToken string) (*schema.User, error) {
	var sub = consts.AnonymousUser
	var jwtIdToken = &schema.User{}

	if a.config.Debug.Enabled {
		jwtIdToken.Token.Sub = sub
		jwtIdToken.Valid = true
		return jwtIdToken, nil
	} else {
		token, err := a.jwks.ParseJWT(jwtToken)
		if err != nil {
			return nil, consts.ErrNotValidToken
		}

		subStr, err := token.Claims.GetSubject()
		if err != nil {
			return nil, consts.ErrNotValidToken
		}

		//subInt, err := strconv.Atoi(subStr)
		//if err != nil {
		//	return nil, consts.ErrNotValidToken
		//}

		sub = schema.UserId(subStr)

		// 如果 token.Header 中没有 typ
		if token.Header["typ"] == "" {
			return nil, consts.ErrEmptyResponse
		}

		// 验证 token 类型
		if tokenType != "" && tokenType.String() != token.Header["typ"] {
			return nil, consts.ErrTokenError
		}

		jwtIdToken.Valid = true

		err = mapstructure.Decode(token.Claims, &jwtIdToken.Token)
		if err != nil {
			a.logger.Logger.Error("Failed to map token claims to JwtIDToken struct.\nError: " + err.Error())
			return nil, nil
		}

		// 手动指定，因为 mapstructure 无法转换 UserID 类型
		jwtIdToken.Token.Sub = sub
	}

	return jwtIdToken, nil
}
