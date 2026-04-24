package schedule

import (
	"context"
	"sync"
	"time"

	"github.com/iVampireSP/go-template/pkg/foundation/lock"
)

// Mutex 分布式锁接口
type Mutex interface {
	Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, name string) error
}

// RedisMutex 基于 Redis 的分布式锁实现
type RedisMutex struct {
	locker *lock.Locker
	locks  map[string]*lock.Lock
	mu     sync.Mutex
}

// NewRedisMutex 创建 Redis 分布式锁
func NewRedisMutex(locker *lock.Locker) *RedisMutex {
	return &RedisMutex{
		locker: locker,
		locks:  make(map[string]*lock.Lock),
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

	lk, err := m.locker.Obtain(ctx, name, ttl, nil)
	if err != nil {
		if err == lock.ErrNotObtained {
			return false, nil
		}
		return false, err
	}

	m.locks[name] = lk
	return true, nil
}

// Release 释放锁
func (m *RedisMutex) Release(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lk, exists := m.locks[name]
	if !exists {
		return nil
	}

	err := lk.Release(ctx)
	delete(m.locks, name)

	if err == lock.ErrLockNotHeld {
		return nil
	}
	return err
}

// NoopMutex 空操作锁（用于测试或单实例部署）
type NoopMutex struct{}

func NewNoopMutex() *NoopMutex                                             { return &NoopMutex{} }
func (m *NoopMutex) Acquire(_ context.Context, _ string, _ time.Duration) (bool, error) { return true, nil }
func (m *NoopMutex) Release(_ context.Context, _ string) error             { return nil }
