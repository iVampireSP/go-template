package orm

import (
	"fmt"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"gorm.io/driver/postgres"

	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

func NewGORM(
	config *conf.Config,
	logger *logger.Logger,
) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Database.Host, config.Database.User, config.Database.Password, config.Database.Name, config.Database.Port, config.Database.SSLMode, config.Database.TimeZone)

	gormConfig := &gorm.Config{}

	if !config.Debug.Enabled {
		zapGormLogger := zapgorm2.New(logger.Logger)
		zapGormLogger.SetAsDefault()
		gormConfig.Logger = zapGormLogger
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		panic(err)
	}

	return db

}
