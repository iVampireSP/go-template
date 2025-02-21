package infra

import (
	"database/sql"
	"go-template/ent"
	"go-template/internal/api"
	"go-template/internal/batch"
	"go-template/internal/infra/conf"
	"go-template/internal/infra/logger"
	"go-template/internal/infra/redis"
	"go-template/internal/infra/s3"
	"go-template/internal/services"
)

type Application struct {
	Config  *conf.Config
	Logger  *logger.Logger
	Api     *api.Api
	Service *services.Service
	Redis   *redis.Redis
	Batch   *batch.Batch
	S3      *s3.S3
	Ent     *ent.Client
	DB      *sql.DB
}

func NewApplication(
	config *conf.Config,
	api *api.Api,
	logger *logger.Logger,
	services *services.Service,
	redis *redis.Redis,
	batch *batch.Batch,
	s3 *s3.S3,
	ent *ent.Client,
	db *sql.DB,
) *Application {
	return &Application{
		Config:  config,
		Api:     api,
		Logger:  logger,
		Service: services,
		Redis:   redis,
		Batch:   batch,
		S3:      s3,
		Ent:     ent,
		DB:      db,
	}
}
