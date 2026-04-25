package queue

import (
	"time"

	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
)

type Provider struct {
	app *container.Application
}

func NewProvider(app *container.Application) *Provider {
	return &Provider{app: app}
}

func (p *Provider) Register() {
	p.app.Singleton(NewDefaultConfig)
	p.app.Singleton(NewDefaultRedisConfig)
	p.app.Singleton(NewQueue)
}

func (p *Provider) Boot() {}

// NewDefaultConfig returns a queue Config populated from the application config.
func NewDefaultConfig() Config {
	cfg := DefaultConfig()
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
	return cfg
}

// NewDefaultRedisConfig returns queue-specific Redis config from the application config.
func NewDefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Host:         config.String("redis.host", "localhost"),
		Port:         config.Int("redis.port", 6379),
		Password:     config.String("redis.password"),
		DB:           config.Int("redis.db", 0),
		ClusterAddrs: config.String("redis.cluster_addrs"),
	}
}
