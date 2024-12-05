package auth

import (
	"context"
	"go-template/internal/types/constants"
	"go-template/internal/types/errs"
	"go-template/internal/types/user"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

func (a *Service) AuthFromToken(tokenType constants.JwtTokenTypes, token string) (*user.User, error) {
	if a.config.Debug.Enabled {
		return a.parseUserJWT(tokenType, "")
	}

	return a.parseUserJWT(tokenType, token)
}

func (a *Service) GetUserFromIdToken(idToken string) (*user.User, error) {
	return a.parseUserJWT(constants.JwtTokenTypeIDToken, idToken)
}

func (a *Service) GetUser(ctx *fiber.Ctx) *user.User {
	userCtx := ctx.Locals(constants.AuthMiddlewareKey)

	u, ok := userCtx.(*user.User)
	u.Id = u.Token.Sub

	if !ok {
		panic("User context is not valid")
	}

	return u
}

func (a *Service) GetCtx(ctx context.Context) *user.User {
	userCtx := ctx.Value(constants.AuthMiddlewareKey)

	u, ok := userCtx.(*user.User)
	u.Id = u.Token.Sub

	if !ok {
		panic("User context is not valid")
	}

	return u
}

func (a *Service) GetUserSafe(ctx *fiber.Ctx) (*user.User, bool) {
	userCtx := ctx.Locals(constants.AuthMiddlewareKey)

	u, ok := userCtx.(*user.User)
	u.Id = u.Token.Sub

	return u, ok
}

func (a *Service) GetCtxSafe(ctx context.Context) (*user.User, bool) {
	userCtx := ctx.Value(constants.AuthMiddlewareKey)

	u, ok := userCtx.(*user.User)
	u.Id = u.Token.Sub

	return u, ok
}

func (a *Service) SetUser(ctx context.Context, user *user.User) context.Context {
	return context.WithValue(ctx, constants.AuthMiddlewareKey, user)
}

func (a *Service) parseUserJWT(tokenType constants.JwtTokenTypes, jwtToken string) (*user.User, error) {
	var sub = user.AnonymousUser
	var jwtIdToken = new(user.User)

	if a.config.Debug.Enabled {
		jwtIdToken.Token.Sub = sub
		jwtIdToken.Valid = true
		return jwtIdToken, nil
	} else {
		token, err := a.jwks.ParseJWT(jwtToken)
		if err != nil {
			return nil, errs.NotValidToken
		}

		subStr, err := token.Claims.GetSubject()
		if err != nil {
			return nil, errs.NotValidToken
		}

		sub = user.Id(subStr)

		// 如果 token.Header 中没有 typ
		if token.Header["typ"] == "" {
			return nil, errs.EmptyResponse
		}

		// 验证 token 类型
		if tokenType != "" && tokenType.String() != token.Header["typ"] {
			return nil, errs.TokenError
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
