package stream

import (
	"go-template/internal/infra/conf"
)

type Service struct {
	config *conf.Config
}

func NewService(config *conf.Config) *Service {
	return &Service{
		config,
	}
}
