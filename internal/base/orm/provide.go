package orm

import (
	"fmt"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

//
//func ProviderOrm() *Orm {
//	return NewOrm()
//}

func NewGORM(
	config *conf.Config,
	logger *logger.Logger,
) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name)

	if config.Debug.Enabled {
		db, err := gorm.Open(mysql.Open(dsn))

		if err != nil {
			panic(err)
		}

		return db
	}

	zapGormLogger := zapgorm2.New(logger.Logger)
	zapGormLogger.SetAsDefault()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: zapGormLogger,
	})

	if err != nil {
		panic(err)
	}

	return db

}
