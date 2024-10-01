package middleware

import (
	"go-template/internal/base/logger"
	"go-template/internal/service/auth"

	"github.com/google/wire"
)

type Middleware struct {
	GinLogger    *GinLoggerMiddleware
	Auth         *AuthMiddleware
	JSONResponse *JSONResponseMiddleware
}

func NewMiddleware(logger *logger.Logger, authService *auth.Service) *Middleware {
	return &Middleware{
		GinLogger:    NewGinLoggerMiddleware(logger.Logger),
		Auth:         NewAuthMiddleware(authService),
		JSONResponse: NewJSONResponseMiddleware(),
	}
}

var Provider = wire.NewSet(
	NewMiddleware,
)
