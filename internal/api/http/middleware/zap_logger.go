package middleware

import (
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
}

func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Handler() fiber.Handler {
	var config = fiberzap.Config{
		Logger: l.logger,
	}

	return fiberzap.New(config)
}
