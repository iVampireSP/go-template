//go:build wireinject
// +build wireinject

package cmd

import (
	v1 "go-template/internal/api/v1"
	"go-template/internal/base"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"go-template/internal/base/orm"
	"go-template/internal/base/redis"
	"go-template/internal/base/s3"
	"go-template/internal/base/server"
	"go-template/internal/batch"
	"go-template/internal/dao"
	"go-template/internal/handler"
	"go-template/internal/handler/grpc"
	"go-template/internal/middleware"
	"go-template/internal/router"
	"go-template/internal/service"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	conf.ProviderConfig,
	logger.NewZapLogger,
	orm.NewGORM,
	dao.NewQuery,
	redis.NewRedis,
	s3.NewS3,
	middleware.Provider,
	batch.NewBatch,
	service.Provider,
	v1.ProviderApiControllerSet,
	grpc.ProviderGrpcHandlerSet,
	grpc.NewGrpcHandlers,
	handler.NewHandler,
	router.ProviderSetRouter,
	server.NewHTTPServer,
	base.NewApplication,
)

func CreateApp() (*base.Application, error) {
	wire.Build(ProviderSet)

	return nil, nil
}
