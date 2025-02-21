package auth

import (
	"go-template/internal/infra/conf"
	"go-template/internal/infra/logger"
)

type Service struct {
	config *conf.Config
	logger *logger.Logger
}

func NewService(
	config *conf.Config,
	logger *logger.Logger,
) *Service {
	return &Service{
		config: config,
		logger: logger,
	}
}
