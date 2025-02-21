//go:build wireinject
// +build wireinject

package cmd

import (
	"github.com/google/wire"
	"go-template/internal/api"
	"go-template/internal/batch"
	"go-template/internal/infra"
	"go-template/internal/infra/conf"
	"go-template/internal/infra/logger"
	"go-template/internal/infra/milvus"
	"go-template/internal/infra/orm"
	"go-template/internal/infra/redis"
	"go-template/internal/infra/s3"
	"go-template/internal/infra/server"
	"go-template/internal/router"
	"go-template/internal/services"
)

var ProviderSet = wire.NewSet(
	// Infra Layer
	conf.NewConfig,
	logger.NewZapLogger,
	orm.ProviderSet,
	redis.NewRedis,
	s3.NewS3,
	//stream.NewStream,
	milvus.NewService,
	batch.NewBatch,

	// Internal Layer
	services.Provide,

	//events.NewEvent,

	// API Layer
	api.Provide,
	router.Provide,
	server.NewHTTPServer,

	// Application
	infra.NewApplication,
)

func CreateApp() (*infra.Application, error) {
	wire.Build(ProviderSet)

	return nil, nil
}
