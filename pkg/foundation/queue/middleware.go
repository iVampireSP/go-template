package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// envelopeContextKey 用于在 context 中存储 envelope
type envelopeContextKey struct{}

// ContextWithEnvelope 将 envelope 存储到 context
func ContextWithEnvelope(ctx context.Context, env *Envelope) context.Context {
	return context.WithValue(ctx, envelopeContextKey{}, env)
}

// EnvelopeFromContext 从 context 获取 envelope
func EnvelopeFromContext(ctx context.Context) *Envelope {
	if env, ok := ctx.Value(envelopeContextKey{}).(*Envelope); ok {
		return env
	}
	return nil
}

// recovery panic 恢复中间件
func recovery() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, payload []byte) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic recovered: %v", r)
					logger.Error("handler panic", zap.Any("panic", r))
				}
			}()
			return next(ctx, payload)
		}
	}
}

// logging 日志中间件
func logging() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, payload []byte) error {
			env := EnvelopeFromContext(ctx)
			start := time.Now()

			err := next(ctx, payload)

			duration := time.Since(start)

			if env != nil {
				if err != nil {
					logger.Warn("message processing failed",
						zap.String("name", env.Name),
						zap.String("id", env.ID),
						zap.Duration("duration", duration),
						zap.Error(err),
					)
				} else {
					logger.Debug("message processed",
						zap.String("name", env.Name),
						zap.String("id", env.ID),
						zap.Duration("duration", duration),
					)
				}
			}

			return err
		}
	}
}

// Idempotent 幂等中间件（基于 Redis）
func Idempotent(redisClient redis.UniversalClient, ttl time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, payload []byte) error {
			env := EnvelopeFromContext(ctx)
			if env == nil {
				return next(ctx, payload)
			}

			key := fmt.Sprintf("queue:processed:%s", env.ID)

			ok, err := redisClient.SetNX(ctx, key, "1", ttl).Result()
			if err != nil {
				logger.Warn("idempotent check failed", zap.Error(err))
				return next(ctx, payload)
			}

			if !ok {
				logger.Debug("message already processed, skipping",
					zap.String("id", env.ID),
					zap.String("name", env.Name),
				)
				return nil
			}

			return next(ctx, payload)
		}
	}
}

// Timeout 超时中间件
func Timeout(duration time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, payload []byte) error {
			ctx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- next(ctx, payload)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// retryInfo 重试信息中间件
func retryInfo() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, payload []byte) error {
			env := EnvelopeFromContext(ctx)
			if env != nil && env.Attempt > 0 {
				logger.Debug("processing retry",
					zap.String("id", env.ID),
					zap.String("name", env.Name),
					zap.Int("attempt", env.Attempt),
					zap.Int("max_retry", env.MaxRetry),
				)
			}
			return next(ctx, payload)
		}
	}
}
