// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package cmd

import (
	"github.com/google/wire"
	"go-template/internal/api"
	"go-template/internal/api/grpc"
	"go-template/internal/api/grpc/interceptor"
	"go-template/internal/api/grpc/v1/documents"
	"go-template/internal/api/http"
	"go-template/internal/api/http/v1"
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
	"go-template/internal/services/auth"
	"go-template/internal/services/jwks"
	"go-template/internal/services/stream"
)

import (
	_ "github.com/lib/pq"
)

// Injectors from wire.go:

func CreateApp() (*infra.Application, error) {
	config := conf.NewConfig()
	loggerLogger := logger.NewZapLogger(config)
	jwksJWKS := jwks.NewJWKS(config, loggerLogger)
	service := auth.NewService(config, jwksJWKS, loggerLogger)
	userController := v1.NewUserController(service)
	handlers := http.NewHandler(userController)
	middleware := http.NewMiddleware(config, loggerLogger, service)
	routerApi := router.NewApiRoute(handlers, middleware)
	swaggerRouter := router.NewSwaggerRoute()
	httpServer := server.NewHTTPServer(config, routerApi, swaggerRouter, middleware, loggerLogger)
	client := orm.NewEnt(config)
	handler := documents.NewHandler(client)
	interceptorAuth := interceptor.NewAuth(service, loggerLogger, config)
	interceptorLogger := interceptor.NewLogger(loggerLogger)
	grpcInterceptor := grpc.NewInterceptor(interceptorAuth, interceptorLogger)
	grpcHandlers := grpc.NewHandler(handler, grpcInterceptor)
	apiApi := api.NewApi(grpcHandlers, handlers)
	streamService := stream.NewService(config)
	servicesService := services.NewService(loggerLogger, jwksJWKS, service, streamService)
	redisRedis := redis.NewRedis(config)
	batchBatch := batch.NewBatch(loggerLogger)
	s3S3 := s3.NewS3(config)
	db := orm.NewSqlDB(config)
	application := infra.NewApplication(config, httpServer, apiApi, loggerLogger, servicesService, redisRedis, batchBatch, s3S3, client, db)
	return application, nil
}

// wire.go:

var ProviderSet = wire.NewSet(conf.NewConfig, logger.NewZapLogger, orm.ProviderSet, redis.NewRedis, s3.NewS3, milvus.NewService, batch.NewBatch, services.Provide, api.Provide, router.Provide, server.NewHTTPServer, infra.NewApplication)
