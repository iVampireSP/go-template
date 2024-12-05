package interceptor

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go-template/internal/base/logger"
	"go.uber.org/zap"
)

type Logger struct {
	Logger *logger.Logger
}

func NewLogger(logger *logger.Logger) *Logger {
	return &Logger{
		Logger: logger,
	}
}

func (l *Logger) ZapLogInterceptor() logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger2 := l.Logger.Logger.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger2.Debug(msg)
		case logging.LevelInfo:
			logger2.Info(msg)
		case logging.LevelWarn:
			logger2.Warn(msg)
		case logging.LevelError:
			logger2.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
