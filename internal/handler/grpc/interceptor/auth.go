package interceptor

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go-template/internal/schema"
	auth2 "go-template/internal/service/auth"
	"go-template/pkg/consts"
)

type Auth struct {
	authService *auth2.Service
}

func NewAuth(authService *auth2.Service) *Auth {
	return &Auth{
		authService: authService,
	}
}

func (a *Auth) JwtAuth(ctx context.Context) (context.Context, error) {
	tokenString, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	token, err := a.authService.AuthFromToken(schema.JWTIDToken, tokenString)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, consts.ErrNotValidToken
	}

	ctx = logging.InjectFields(ctx, logging.Fields{"auth.sub", token.Token.Sub})

	return a.authService.SetUser(ctx, token), nil
}
