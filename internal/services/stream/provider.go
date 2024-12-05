package stream

import (
	"go-template/internal/base/conf"
)

type Service struct {
	config *conf.Config
}

func NewService(config *conf.Config) *Service {
	return &Service{
		config,
	}
}
