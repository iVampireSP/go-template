package auth

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"go-template/internal/types/auth"
	"go-template/internal/types/constants"
)

func (a *Service) AuthFromToken(tokenType auth.JWTTokenTypes, token string) (*auth.User, error) {
	if a.config.Debug.Enabled {
		return a.parseUserJWT(tokenType, "")
	}

	return a.parseUserJWT(tokenType, token)
}

func (a *Service) GetUserFromIdToken(idToken string) (*auth.User, error) {
	return a.parseUserJWT(auth.JWTIDToken, idToken)
}

func (a *Service) GetUser(ctx *fiber.Ctx) *auth.User {
	userCtx := ctx.Locals(constants.AuthMiddlewareKey)

	u, ok := userCtx.(*auth.User)
	u.Id = u.Token.Sub

	if !ok {
		panic("User context is not valid")
	}

	return u
}

func (a *Service) GetCtx(ctx context.Context) *auth.User {
	userCtx := ctx.Value(constants.AuthMiddlewareKey)

	u, ok := userCtx.(*auth.User)
	u.Id = u.Token.Sub

	if !ok {
		panic("User context is not valid")
	}

	return u
}

func (a *Service) GetUserSafe(ctx *fiber.Ctx) (*auth.User, bool) {
	userCtx := ctx.Locals(constants.AuthMiddlewareKey)

	u, ok := userCtx.(*auth.User)
	u.Id = u.Token.Sub

	return u, ok
}

func (a *Service) GetCtxSafe(ctx context.Context) (*auth.User, bool) {
	userCtx := ctx.Value(constants.AuthMiddlewareKey)

	u, ok := userCtx.(*auth.User)
	u.Id = u.Token.Sub

	return u, ok
}

func (a *Service) SetUser(ctx context.Context, user *auth.User) context.Context {
	return context.WithValue(ctx, constants.AuthMiddlewareKey, user)
}

func (a *Service) parseUserJWT(tokenType auth.JWTTokenTypes, jwtToken string) (*auth.User, error) {
	var sub = constants.AnonymousUser
	var jwtIdToken = new(auth.User)

	if a.config.Debug.Enabled {
		jwtIdToken.Token.Sub = sub
		jwtIdToken.Valid = true
		return jwtIdToken, nil
	} else {
		token, err := a.jwks.ParseJWT(jwtToken)
		if err != nil {
			return nil, constants.ErrNotValidToken
		}

		subStr, err := token.Claims.GetSubject()
		if err != nil {
			return nil, constants.ErrNotValidToken
		}

		sub = auth.Id(subStr)

		// 如果 token.Header 中没有 typ
		if token.Header["typ"] == "" {
			return nil, constants.ErrEmptyResponse
		}

		// 验证 token 类型
		if tokenType != "" && tokenType.String() != token.Header["typ"] {
			return nil, constants.ErrTokenError
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
