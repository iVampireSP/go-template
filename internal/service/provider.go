package service

import (
	"go-template/internal/base/logger"
	"go-template/internal/service/auth"
	"go-template/internal/service/jwks"

	"github.com/google/wire"
)

type Service struct {
	logger *logger.Logger
	Jwks   *jwks.JWKS
	Auth   *auth.Service
}

var Provider = wire.NewSet(
	jwks.NewJWKS,
	auth.NewAuthService,
	NewService,
)

func NewService(
	logger *logger.Logger,
	jwks *jwks.JWKS,
	auth *auth.Service,

) *Service {
	return &Service{
		logger,
		jwks,
		auth,
	}
}
