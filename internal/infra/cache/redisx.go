package cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/iVampireSP/go-template/internal/infra/config"
	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// New 创建 Redis 客户端与分布式锁
// 支持单机模式和集群模式，通过配置自动选择
func New() (redis.UniversalClient, *Locker) {
	clusterAddresses := config.String("redis.cluster_addrs")
	password := config.String("redis.password")

	var client redis.UniversalClient

	if clusterAddresses != "" {
		// 集群模式
		addresses := parseClusterAddresses(clusterAddresses)
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addresses,
			Password: password,
		})
	} else {
		// 单机模式
		host := config.String("redis.host", "localhost")
		port := config.Int("redis.port", 6379)
		db := config.Int("redis.db", 0)

		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       db,
		})
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	// Enable OTel tracing and metrics for Redis commands
	if err := redisotel.InstrumentTracing(client); err != nil {
		logger.Error("Failed to instrument Redis", "tracing", err)
	}
	if err := redisotel.InstrumentMetrics(client); err != nil {
		logger.Error("Failed to instrument Redis", "metrics", err)
	}

	locker := NewLocker(client)
	return client, locker
}

// parseClusterAddresses 解析集群地址字符串
// 支持格式: "node1:6379,node2:6379,node3:6379" 或 "node1,node2,node3" (自动添加默认端口)
func parseClusterAddresses(addresses string) []string {
	parts := strings.Split(addresses, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		addr := strings.TrimSpace(part)
		if addr != "" {
			// 如果地址不包含端口，添加默认端口 6379
			if !strings.Contains(addr, ":") {
				addr = addr + ":6379"
			}
			result = append(result, addr)
		}
	}
	return result
}
