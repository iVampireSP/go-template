//go:build wireinject
// +build wireinject

package cmd

import (
	"go-template/internal/api"
	"go-template/internal/base"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"go-template/internal/base/milvus"
	"go-template/internal/base/orm"
	"go-template/internal/base/redis"
	"go-template/internal/base/s3"
	"go-template/internal/base/server"
	"go-template/internal/batch"
	"go-template/internal/dao"
	"go-template/internal/router"
	"go-template/internal/services"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	conf.NewConfig,
	logger.NewZapLogger,
	orm.NewGORM,
	dao.NewQuery,
	redis.NewRedis,
	s3.NewS3,
	milvus.NewService,
	batch.NewBatch,
	services.Provide,
	api.Provide,
	router.Provide,
	server.NewHTTPServer,
	base.NewApplication,
)

func CreateApp() (*base.Application, error) {
	wire.Build(ProviderSet)

	return nil, nil
}
