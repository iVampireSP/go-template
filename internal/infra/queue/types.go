package queue

import (
	"context"
	"time"
)

const (
	DefaultQueue = "default"
)

// Dispatcher 提交 Job 到队列的能力接口。
// Queue 实现此接口。消费方应依赖 Dispatcher 而非 *Queue。
type Dispatcher interface {
	Dispatch(ctx context.Context, job Job, dispatchOptions ...DispatchOption) (string, error)
}

type Job interface {
	Name() string
	Key() string
}

// RetryConfig 控制 Job 的重试策略
type RetryConfig struct {
	// MaxRetry 总执行次数上限（含首次），0 表示使用 infra 默认值
	MaxRetry int
	// RetryDelays 每次重试前的等待时间；不足时使用最后一个值，为空时使用 infra 默认值
	RetryDelays []time.Duration
}

// JobWithRetryConfig Job 可实现此接口以声明自己的重试策略
type JobWithRetryConfig interface {
	Job
	RetryConfig() RetryConfig
}

// JobWithMaxRetry 兼容旧接口，新 Job 优先使用 JobWithRetryConfig
type JobWithMaxRetry interface {
	Job
	MaxRetry() int
}

type JobWithQueue interface {
	Job
	Queue() string
}

type DispatchOption func(*DispatchConfig)

type DispatchConfig struct {
	ProcessAt *time.Time
	ProcessIn *time.Duration
}

func WithProcessAt(at time.Time) DispatchOption {
	return func(cfg *DispatchConfig) {
		if at.IsZero() {
			return
		}
		t := at
		cfg.ProcessAt = &t
		cfg.ProcessIn = nil
	}
}

func WithProcessIn(delay time.Duration) DispatchOption {
	return func(cfg *DispatchConfig) {
		if delay <= 0 {
			return
		}
		d := delay
		cfg.ProcessIn = &d
		if cfg.ProcessAt != nil {
			cfg.ProcessAt = nil
		}
	}
}

type Handler func(ctx context.Context, payload []byte) error

type Middleware func(next Handler) Handler
