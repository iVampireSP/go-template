package bootstrap

import (
	"time"

	"github.com/iVampireSP/go-template/pkg/foundation/bus"
	"github.com/iVampireSP/go-template/pkg/foundation/cache"
	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/email"
	"github.com/iVampireSP/go-template/pkg/foundation/queue"
	"github.com/iVampireSP/go-template/pkg/logger"
)

func NewLoggerConfig() logger.Config {
	return logger.Config{
		Level: config.String("log.level", "info"),
		Debug: config.Bool("app.debug", false),
	}
}

func NewRedisConfig() cache.RedisConfig {
	return cache.RedisConfig{
		Host:         config.String("redis.host", "localhost"),
		Port:         config.Int("redis.port", 6379),
		Password:     config.String("redis.password"),
		DB:           config.Int("redis.db", 0),
		ClusterAddrs: config.String("redis.cluster_addrs"),
	}
}

func NewQueueRedisConfig() queue.RedisConfig {
	return queue.RedisConfig{
		Host:         config.String("redis.host", "localhost"),
		Port:         config.Int("redis.port", 6379),
		Password:     config.String("redis.password"),
		DB:           config.Int("redis.db", 0),
		ClusterAddrs: config.String("redis.cluster_addrs"),
	}
}

func NewQueueConfig() queue.Config {
	cfg := queue.DefaultConfig()
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

func NewEmailConfig() email.Config {
	return email.Config{
		Host:       config.String("mail.host"),
		Port:       config.Int("mail.port", 587),
		Username:   config.String("mail.username"),
		Password:   config.String("mail.password"),
		From:       config.String("mail.from"),
		Encryption: config.String("mail.encryption", "tls"),
	}
}

func NewBusConfig() bus.Config {
	var cfg bus.Config
	if err := config.Unmarshal("bus", &cfg); err != nil {
		panic(err)
	}
	return cfg
}
