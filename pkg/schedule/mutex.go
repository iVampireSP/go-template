package schedule

import (
	"context"
	"sync"
	"time"

	"github.com/iVampireSP/go-template/internal/infra/cache"
)

// Mutex 分布式锁接口
type Mutex interface {
	// Acquire 获取锁，返回是否成功
	Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error)
	// Release 释放锁
	Release(ctx context.Context, name string) error
}

// RedisMutex 基于 Redis 的分布式锁实现
type RedisMutex struct {
	locker *cache.Locker
	locks  map[string]*cache.Lock
	mu     sync.Mutex
}

// NewRedisMutex 创建 Redis 分布式锁
func NewRedisMutex(locker *cache.Locker) *RedisMutex {
	return &RedisMutex{
		locker: locker,
		locks:  make(map[string]*cache.Lock),
	}
}

// Acquire 获取锁
func (m *RedisMutex) Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existingLock, exists := m.locks[name]; exists {
		_ = existingLock.Release(ctx)
		delete(m.locks, name)
	}

	lock, err := m.locker.Obtain(ctx, name, ttl, nil)
	if err != nil {
		if err == cache.ErrNotObtained {
			return false, nil
		}
		return false, err
	}

	m.locks[name] = lock
	return true, nil
}

// Release 释放锁
func (m *RedisMutex) Release(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, exists := m.locks[name]
	if !exists {
		return nil
	}

	err := lock.Release(ctx)
	delete(m.locks, name)

	if err == cache.ErrLockNotHeld {
		return nil
	}
	return err
}

// NoopMutex 空操作锁（用于测试或单实例部署）
type NoopMutex struct{}

// NewNoopMutex 创建空操作锁
func NewNoopMutex() *NoopMutex {
	return &NoopMutex{}
}

// Acquire 始终返回成功
func (m *NoopMutex) Acquire(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}

// Release 空操作
func (m *NoopMutex) Release(_ context.Context, _ string) error {
	return nil
}
