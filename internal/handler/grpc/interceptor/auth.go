package interceptor

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"go-template/internal/schema"
	auth2 "go-template/internal/service/auth"
	"go-template/pkg/consts"
	"google.golang.org/grpc"
)

var ignoreAuthApis = map[string]bool{
	// 反射
	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": true,

	// 业务 API
	"/api.v1.WorkspaceService/List": true,
}

type Auth struct {
	authService *auth2.Service
	logger      *logger.Logger
	config      *conf.Config
}

func NewAuth(
	authService *auth2.Service,
	logger *logger.Logger,
	config *conf.Config,
) *Auth {
	return &Auth{
		authService: authService,
		logger:      logger,
		config:      config,
	}
}

func (a *Auth) notRequireAuth(fullMethodName string) bool {
	var b = ignoreAuthApis[fullMethodName]

	if a.config.Debug.Enabled {
		if b {
			a.logger.Sugar.Debug("[GRPC Auth] Ignore auth for Method: %s", fullMethodName)
		} else {
			a.logger.Sugar.Debug("[GRPC Auth] Require auth for Method: %s", fullMethodName)
		}
	}

	return b
}

func (a *Auth) authCtx(ctx context.Context) (context.Context, error) {
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

	ctx = logging.InjectFields(ctx, logging.Fields{consts.AuthMiddlewareKey, token.Token.Sub})
	ctx = context.WithValue(ctx, consts.AuthMiddlewareKey, token)

	return ctx, nil
}

func (a *Auth) UnaryJWTAuth() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if a.notRequireAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		ctx, err := a.authCtx(ctx)

		if err != nil {
			return nil, err
		}

		result, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		return result, err
	}
}

func (a *Auth) StreamJWTAuth() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var ctx = ss.Context()

		if a.notRequireAuth(info.FullMethod) {
			return handler(srv, ss)
		}
		ctx, err := a.authCtx(ctx)

		if err != nil {
			return err
		}

		err = handler(srv, ss)
		if err != nil {
			return err
		}

		return nil
	}
}

//
//func (a *Auth) JwtAuth(ctx context.Context) (context.Context, error) {
//	tokenString, err := auth.AuthFromMD(ctx, "bearer")
//	if err != nil {
//		return nil, err
//	}
//
//	token, err := a.authService.AuthFromToken(schema.JWTIDToken, tokenString)
//	if err != nil {
//		return nil, err
//	}
//
//	if !token.Valid {
//		return nil, consts.ErrNotValidToken
//	}
//
//	ctx = logging.InjectFields(ctx, logging.Fields{consts.AuthMiddlewareKey, token.Token.Sub})
//
//	return context.WithValue(ctx, consts.AuthMiddlewareKey, token), nil
//}
