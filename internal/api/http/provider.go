package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"go-template/internal/api/http/middleware"
	"go-template/internal/api/http/v1"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"go-template/internal/service/auth"
)

type IMiddleware interface {
	Handler() fiber.Handler
}

type Middleware struct {
	Logger       IMiddleware
	Auth         IMiddleware
	JSONResponse IMiddleware
	RBAC         *middleware.RBAC
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
		Logger:       middleware.NewLogger(logger.Logger),
		Auth:         middleware.NewAuth(config, authService),
		JSONResponse: middleware.NewJSONResponse(),
		RBAC:         middleware.NewRBAC(authService, config),
	}
}

var ProviderSet = wire.NewSet(
	// Init Middleware
	middleware.NewAuth,
	middleware.NewLogger,
	middleware.NewJSONResponse,
	middleware.NewRBAC,
	NewMiddleware,

	// Init Controller
	v1.NewUserController,

	// Init Handler
	NewHandler,
)
