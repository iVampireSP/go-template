package base

import (
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"go-template/internal/base/redis"
	"go-template/internal/base/s3"
	"go-template/internal/base/server"
	"go-template/internal/batch"
	"go-template/internal/dao"
	"go-template/internal/middleware"
	"go-template/internal/service"
	"gorm.io/gorm"
)

type Application struct {
	Config     *conf.Config
	HttpServer *server.HttpServer
	Logger     *logger.Logger
	GORM       *gorm.DB
	DAO        *dao.Query
	Service    *service.Service
	Middleware *middleware.Middleware
	Redis      *redis.Redis
	Batch      *batch.Batch
	S3         *s3.S3
}

func NewApplication(
	config *conf.Config,
	httpServer *server.HttpServer,
	logger *logger.Logger,
	services *service.Service,
	middleware *middleware.Middleware,
	redis *redis.Redis,
	batch *batch.Batch,
	S3 *s3.S3,
	GORM *gorm.DB,
	DAO *dao.Query,
) *Application {
	return &Application{
		Config:     config,
		HttpServer: httpServer,
		Logger:     logger,
		Service:    services,
		Middleware: middleware,
		Redis:      redis,
		Batch:      batch,
		S3:         S3,
		GORM:       GORM,
		DAO:        DAO,
	}
}
