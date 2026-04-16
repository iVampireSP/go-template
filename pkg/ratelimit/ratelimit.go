package ratelimit

import (
	"context"
	"strconv"
	"time"

	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// luaSlidingWindow 滑动窗口限流 Lua 脚本。
// 仅操作 KEYS[1]（单 key），兼容 Redis Cluster。
var luaSlidingWindow = redis.NewScript(`
local key = KEYS[1]
local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

redis.call("ZREMRANGEBYSCORE", key, 0, now - window)
local count = redis.call("ZCARD", key)
if count >= limit then
    return 0
end
redis.call("ZADD", key, now, now .. ":" .. math.random(1000000))
redis.call("PEXPIRE", key, window)
return 1
`)

// Limiter 基于 Redis 的分布式滑动窗口限流器。
type Limiter struct {
	client redis.UniversalClient
	prefix string
}

// New 创建限流器实例。prefix 用于 key 前缀（如 "rl:oauth:"）。
func New(client redis.UniversalClient, prefix string) *Limiter {
	return &Limiter{
		client: client,
		prefix: prefix,
	}
}

// Allow 检查 key 在 window 窗口内是否未超过 rate 次。
// Redis 不可用时 fail-open 放行。
func (l *Limiter) Allow(ctx context.Context, key string, rate int, window time.Duration) bool {
	if l == nil || l.client == nil {
		return true
	}
	nowMs := time.Now().UnixMilli()
	windowMs := window.Milliseconds()

	result, err := luaSlidingWindow.Run(ctx, l.client, []string{l.prefix + key},
		strconv.FormatInt(windowMs, 10),
		strconv.Itoa(rate),
		strconv.FormatInt(nowMs, 10),
	).Int64()
	if err != nil {
		logger.Warn("ratelimit: redis error (fail-open)", "err", err)
		return true
	}
	return result == 1
}
