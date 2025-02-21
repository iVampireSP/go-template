package stream

import (
	"go-template/internal/infra/conf"
	"go-template/internal/infra/stream"
)

type Service struct {
	config        *conf.Config
	streamService *stream.Stream
}

func NewService(config *conf.Config, streamService *stream.Stream) *Service {
	return &Service{
		config,
		streamService,
	}
}
