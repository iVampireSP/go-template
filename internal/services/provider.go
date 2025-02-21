package services

import (
	"go-template/internal/infra/logger"
	"go-template/internal/services/auth"
	"go-template/internal/services/stream"

	"github.com/google/wire"
)

type Service struct {
	logger *logger.Logger
	Auth   *auth.Service
	//Stream *stream.Service
}

var Provide = wire.NewSet(
	auth.NewService,
	stream.NewService,
	NewService,
)

func NewService(
	logger *logger.Logger,
	auth *auth.Service,
	// stream *stream.Service,
) *Service {
	return &Service{
		logger,
		auth,
		//stream,
	}
}
