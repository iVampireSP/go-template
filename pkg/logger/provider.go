package logger

import (
	"github.com/iVampireSP/go-template/internal/infra/config"
	"go.uber.org/zap"
)

// NewLogger 从配置创建 Logger 实例用于依赖注入
func NewLogger() (*Logger, *zap.Logger, *zap.SugaredLogger) {
	logConfig := Config{
		Level: config.String("log.level", "info"),
		Debug: config.Bool("app.debug", false),
	}

	logger, sugar := New(logConfig)
	return &Logger{
		Sugar:  sugar,
		Logger: logger,
	}, logger, sugar
}
