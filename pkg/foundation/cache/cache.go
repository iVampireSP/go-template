package cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisConfig Redis 连接配置。
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	ClusterAddrs string // 逗号分隔的集群地址，非空时启用集群模式
}

// NewCache 根据配置创建 Redis 客户端。
// 支持单机模式和集群模式，通过 ClusterAddrs 是否为空自动选择。
func NewCache(cfg RedisConfig, logger *zap.SugaredLogger) redis.UniversalClient {
	var client redis.UniversalClient

	if cfg.ClusterAddrs != "" {
		addresses := parseClusterAddresses(cfg.ClusterAddrs)
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addresses,
			Password: cfg.Password,
		})
	} else {
		host := cfg.Host
		if host == "" {
			host = "localhost"
		}
		port := cfg.Port
		if port == 0 {
			port = 6379
		}
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	if err := redisotel.InstrumentTracing(client); err != nil {
		if logger != nil {
			logger.Errorw("failed to instrument Redis tracing", "error", err)
		}
	}
	if err := redisotel.InstrumentMetrics(client); err != nil {
		if logger != nil {
			logger.Errorw("failed to instrument Redis metrics", "error", err)
		}
	}

	return client
}

func parseClusterAddresses(addresses string) []string {
	parts := strings.Split(addresses, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		addr := strings.TrimSpace(part)
		if addr != "" {
			if !strings.Contains(addr, ":") {
				addr = addr + ":6379"
			}
			result = append(result, addr)
		}
	}
	return result
}
