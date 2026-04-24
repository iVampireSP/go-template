package queue

import (
	"time"
)

// Config 任务队列配置。
type Config struct {
	MaxRetry    int
	RetryDelays []time.Duration
}

// DefaultConfig 返回默认配置。
func DefaultConfig() Config {
	return Config{
		MaxRetry: 5,
		RetryDelays: []time.Duration{
			1 * time.Second,
			2 * time.Second,
			4 * time.Second,
			8 * time.Second,
			16 * time.Second,
		},
	}
}

// RedisConfig Asynq Redis 连接配置。
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	ClusterAddrs string // 逗号分隔，非空启用集群模式
}
