package logger

import "go.uber.org/zap"

type Logger struct {
	Sugar  *zap.SugaredLogger
	Logger *zap.Logger
}

func NewZapLogger() *Logger {
	logger, err := zap.NewProduction()
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
