package base

import (
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"go-template/internal/api"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"go-template/internal/base/redis"
	"go-template/internal/base/s3"
	"go-template/internal/base/server"
	"go-template/internal/batch"
	"go-template/internal/dao"
	"go-template/internal/service"
	"gorm.io/gorm"
)

type Application struct {
	Config     *conf.Config
	Logger     *logger.Logger
	Api        *api.Api
	HttpServer *server.HttpServer
	GORM       *gorm.DB
	DAO        *dao.Query
	Service    *service.Service
	Redis      *redis.Redis
	Batch      *batch.Batch
	S3         *s3.S3
	Milvus     client.Client
}

func NewApplication(
	config *conf.Config,
	httpServer *server.HttpServer,
	api *api.Api,
	logger *logger.Logger,
	services *service.Service,
	redis *redis.Redis,
	batch *batch.Batch,
	s3 *s3.S3,
	gorm *gorm.DB,
	dao *dao.Query,
	milvus client.Client,
) *Application {
	return &Application{
		Config:     config,
		HttpServer: httpServer,
		Api:        api,
		Logger:     logger,
		Service:    services,
		Redis:      redis,
		Batch:      batch,
		S3:         s3,
		GORM:       gorm,
		DAO:        dao,
		Milvus:     milvus,
	}
}
