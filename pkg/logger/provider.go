package logger

import (
	"go.uber.org/zap"
)

// NewLogger 从配置创建 Logger 实例用于依赖注入。
// 由外部传入 Config，不再读取全局 config。
func NewLogger(logConfig Config) (*Logger, *zap.Logger, *zap.SugaredLogger) {
	logger, sugar := New(logConfig)
	return &Logger{
		Sugar:  sugar,
		Logger: logger,
	}, logger, sugar
}
