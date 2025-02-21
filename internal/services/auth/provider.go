package auth

import (
	"go-template/internal/infra/conf"
	"go-template/internal/infra/logger"
	"go-template/internal/services/jwks"
)

type Service struct {
	config *conf.Config
	jwks   *jwks.JWKS
	logger *logger.Logger
}

func NewService(
	config *conf.Config,
	jwks *jwks.JWKS,
	logger *logger.Logger,
) *Service {
	return &Service{
		config: config,
		jwks:   jwks,
		logger: logger,
	}
}
