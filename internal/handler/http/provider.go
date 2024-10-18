package http

import (
	"github.com/google/wire"
	v1 "go-template/internal/handler/http/controller/v1"
	"go-template/internal/handler/http/middleware"
)

var ProviderSet = wire.NewSet(
	middleware.NewAuthMiddleware,
	middleware.NewGinLoggerMiddleware,
	middleware.NewJSONResponseMiddleware,
	NewMiddleware,
	v1.NewUserController,
	NewHandler,
)

type Middleware struct {
	GinLogger    *middleware.GinLoggerMiddleware
	Auth         *middleware.AuthMiddleware
	JSONResponse *middleware.JSONResponseMiddleware
}

func NewMiddleware(
	GinLogger *middleware.GinLoggerMiddleware,
	Auth *middleware.AuthMiddleware,
	JSONResponse *middleware.JSONResponseMiddleware,
) *Middleware {
	return &Middleware{
		Auth:         Auth,
		GinLogger:    GinLogger,
		JSONResponse: JSONResponse,
	}
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
