package auth

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"go-template/internal/consts"
	"go-template/internal/pkg/user"
	"go-template/internal/schema"
)

func (a *Service) AuthFromToken(tokenType schema.JWTTokenTypes, token string) (*user.User, error) {
	if a.config.Debug.Enabled {
		return a.parseUserJWT(tokenType, "")
	}

	return a.parseUserJWT(tokenType, token)
}

func (a *Service) GetUserFromIdToken(idToken string) (*user.User, error) {
	return a.parseUserJWT(schema.JWTIDToken, idToken)
}

func (a *Service) GetUser(ctx *fiber.Ctx) *user.User {
	userCtx := ctx.Locals(consts.AuthMiddlewareKey)

	u, ok := userCtx.(*user.User)
	u.Id = u.Token.Sub

	if !ok {
		panic("User context is not valid")
	}

	return u
}

func (a *Service) GetCtx(ctx context.Context) *user.User {
	userCtx := ctx.Value(consts.AuthMiddlewareKey)

	u, ok := userCtx.(*user.User)
	u.Id = u.Token.Sub

	if !ok {
		panic("User context is not valid")
	}

	return u
}

func (a *Service) GetUserSafe(ctx *fiber.Ctx) (*user.User, bool) {
	userCtx := ctx.Locals(consts.AuthMiddlewareKey)

	u, ok := userCtx.(*user.User)
	u.Id = u.Token.Sub

	return u, ok
}

func (a *Service) GetCtxSafe(ctx context.Context) (*user.User, bool) {
	userCtx := ctx.Value(consts.AuthMiddlewareKey)

	u, ok := userCtx.(*user.User)
	u.Id = u.Token.Sub

	return u, ok
}

func (a *Service) SetUser(ctx context.Context, user *user.User) context.Context {
	return context.WithValue(ctx, consts.AuthMiddlewareKey, user)
}

func (a *Service) parseUserJWT(tokenType schema.JWTTokenTypes, jwtToken string) (*user.User, error) {
	var sub = consts.AnonymousUser
	var jwtIdToken = new(user.User)

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

		sub = user.Id(subStr)

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
