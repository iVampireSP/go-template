package interceptor

import (
	"go-template/internal/infra/conf"
	"go-template/internal/infra/logger"
	authService "go-template/internal/services/auth"
)

type Auth struct {
	authService *authService.Service
	logger      *logger.Logger
	config      *conf.Config
}

func NewAuth(
	authService *authService.Service,
	logger *logger.Logger,
	config *conf.Config,
) *Auth {
	return &Auth{
		authService: authService,
		logger:      logger,
		config:      config,
	}
}
