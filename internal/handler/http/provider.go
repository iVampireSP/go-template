package http

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	v1 "go-template/internal/handler/http/controller/v1"
	"go-template/internal/handler/http/middleware"
	"go-template/internal/service/auth"
)

type IMiddleware interface {
	Handler() echo.MiddlewareFunc
}

type Middleware struct {
	Logger       IMiddleware
	Auth         IMiddleware
	JSONResponse IMiddleware
}

type Handlers struct {
	User *v1.UserController
}

func NewHandler(
	user *v1.UserController,
) *Handlers {
	return &Handlers{
		User: user,
	}
}

func NewMiddleware(config *conf.Config, logger *logger.Logger, authService *auth.Service) *Middleware {
	return &Middleware{
		Logger:       middleware.NewLoggerMiddleware(logger.Logger),
		Auth:         middleware.NewAuthMiddleware(config, authService),
		JSONResponse: middleware.NewJSONResponseMiddleware(),
	}
}

var ProviderSet = wire.NewSet(
	middleware.NewAuthMiddleware,
	middleware.NewLoggerMiddleware,
	middleware.NewJSONResponseMiddleware,
	NewMiddleware,
	v1.NewUserController,
	NewHandler,
)
