package logger

import (
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 封装 zap logger，提供结构化日志 API。
// 所有日志方法签名为 (msg string, keysAndValues ...any)，
// 底层调用 zap SugaredLogger 的 Infow/Warnw/Errorw 等结构化方法。
type Logger struct {
	Sugar  *zap.SugaredLogger
	Logger *zap.Logger
}

// Config logger 配置
type Config struct {
	Level string // debug, info, warn, error
	Debug bool   // 是否开启调试模式 (APP_DEBUG)
}

var (
	globalLogger *Logger
	once         sync.Once
)

// init 初始化全局 logger（从环境变量读取配置）
func init() {
	once.Do(func() {
		_ = godotenv.Load()

		config := Config{
			Level: getEnv("LOG_LEVEL", "info"),
			Debug: getEnvBool("APP_DEBUG", false),
		}
		logger, sugar := New(config)
		globalLogger = &Logger{
			Sugar:  sugar,
			Logger: logger,
		}
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}

// New 创建新的 logger 实例
func New(config Config) (*zap.Logger, *zap.SugaredLogger) {
	var encoderConfig zapcore.EncoderConfig

	if config.Debug {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	core := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		getLogLevel(config.Level),
	)

	var logger *zap.Logger
	if config.Debug {
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.Development())
	} else {
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	return logger, logger.Sugar()
}

func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// --- 实例方法：结构化日志 ---

func (l *Logger) Debug(msg string, keysAndValues ...any) {
	l.Sugar.Debugw(msg, keysAndValues...)
}

func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.Sugar.Infow(msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.Sugar.Warnw(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.Sugar.Errorw(msg, keysAndValues...)
}

func (l *Logger) Fatal(msg string, keysAndValues ...any) {
	l.Sugar.Fatalw(msg, keysAndValues...)
}

// With 添加结构化上下文字段
func (l *Logger) With(keysAndValues ...any) *Logger {
	return &Logger{
		Sugar:  l.Sugar.With(keysAndValues...),
		Logger: l.Logger,
	}
}

// Sync 刷新缓冲区
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// --- 全局函数 ---

func Debug(msg string, keysAndValues ...any) {
	globalLogger.Sugar.Debugw(msg, keysAndValues...)
}

func Info(msg string, keysAndValues ...any) {
	globalLogger.Sugar.Infow(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...any) {
	globalLogger.Sugar.Warnw(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...any) {
	globalLogger.Sugar.Errorw(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...any) {
	globalLogger.Sugar.Fatalw(msg, keysAndValues...)
}

func With(keysAndValues ...any) *Logger {
	return &Logger{
		Sugar:  globalLogger.Sugar.With(keysAndValues...),
		Logger: globalLogger.Logger,
	}
}

func Sync() error {
	return globalLogger.Logger.Sync()
}

func GetLogger() *Logger {
	return globalLogger
}
