package queue

import (
	"time"

	"github.com/iVampireSP/go-template/internal/infra/config"
	foundationqueue "github.com/iVampireSP/go-template/pkg/foundation/queue"
	"github.com/redis/go-redis/v9"
)

// Re-export foundation types.
type Queue = foundationqueue.Queue
type Job = foundationqueue.Job
type Handler = foundationqueue.Handler
type Middleware = foundationqueue.Middleware
type Dispatcher = foundationqueue.Dispatcher
type Envelope = foundationqueue.Envelope

var (
	ContextWithEnvelope = foundationqueue.ContextWithEnvelope
	EnvelopeFromContext = foundationqueue.EnvelopeFromContext
	Idempotent          = foundationqueue.Idempotent
	Timeout             = foundationqueue.Timeout
	WithProcessAt       = foundationqueue.WithProcessAt
	WithProcessIn       = foundationqueue.WithProcessIn
)

func NewQueue(redisClient redis.UniversalClient) *Queue {
	cfg := foundationqueue.DefaultConfig()
	if mr := config.Int("job.defaults.max_retry"); mr > 0 {
		cfg.MaxRetry = mr
	}
	if delays := config.Strings("job.defaults.retry_delays"); len(delays) > 0 {
		cfg.RetryDelays = make([]time.Duration, 0, len(delays))
		for _, d := range delays {
			if duration, err := time.ParseDuration(d); err == nil {
				cfg.RetryDelays = append(cfg.RetryDelays, duration)
			}
		}
	}

	redisCfg := foundationqueue.RedisConfig{
		Host:         config.String("redis.host", "localhost"),
		Port:         config.Int("redis.port", 6379),
		Password:     config.String("redis.password"),
		DB:           config.Int("redis.db", 0),
		ClusterAddrs: config.String("redis.cluster_addrs"),
	}

	return foundationqueue.NewQueue(redisCfg, cfg, redisClient)
}
