package service

import (
	"go-template/internal/base/logger"
	"go-template/internal/service/auth"
	"go-template/internal/service/jwks"
	"go-template/internal/service/stream"

	"github.com/google/wire"
)

type Service struct {
	logger *logger.Logger
	Jwks   *jwks.JWKS
	Auth   *auth.Service
	Stream *stream.Service
}

var Provider = wire.NewSet(
	jwks.NewJWKS,
	auth.NewAuthService,
	stream.NewService,
	NewService,
)

func NewService(
	logger *logger.Logger,
	jwks *jwks.JWKS,
	auth *auth.Service,
	stream *stream.Service,
) *Service {
	return &Service{
		logger,
		jwks,
		auth,
		stream,
	}
}
