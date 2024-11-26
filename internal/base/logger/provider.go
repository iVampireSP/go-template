package logger

import (
	"go-template/internal/base/conf"
	"go.uber.org/zap"
)

type Logger struct {
	Sugar  *zap.SugaredLogger
	Logger *zap.Logger
}

func NewZapLogger(config *conf.Config) *Logger {
	var logger *zap.Logger
	var err error

	if config.Debug.Enabled {
		logger, err = zap.NewDevelopment(zap.AddCallerSkip(1))
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic(err)
		return nil
	}

	//defer func(logger *zap.Logger) {
	//	err := logger.Sync()
	//	if err != nil {
	//		panic(err)
	//	}
	//}(logger)

	return &Logger{Sugar: logger.Sugar(), Logger: logger}
}
