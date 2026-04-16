package cache

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrNotObtained is returned when a lock cannot be obtained.
	ErrNotObtained = errors.New("lock: not obtained")

	// ErrLockNotHeld is returned when trying to release an inactive lock.
	ErrLockNotHeld = errors.New("lock: lock not held")
)

// Lua scripts for atomic operations
// These scripts only operate on a single key, which is compatible with Redis Cluster mode.
var (
	// luaObtain: SET key value NX PX ttl
	luaObtain = redis.NewScript(`
		return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
	`)

	// luaRelease: DEL key if value matches
	luaRelease = redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		end
		return 0
	`)

	// luaRefresh: PEXPIRE key ttl if value matches
	luaRefresh = redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		end
		return 0
	`)
)

// Locker is a distributed lock client using Redis.
type Locker struct {
	client redis.UniversalClient
}

// NewLocker creates a new Locker instance.
func NewLocker(client redis.UniversalClient) *Locker {
	return &Locker{
		client: client,
	}
}

// Lock represents an obtained distributed lock.
type Lock struct {
	locker *Locker
	key    string
	value  string
}

// ObtainOptions configures lock acquisition behavior.
type ObtainOptions struct {
	// RetryCount is the number of retries if lock is not obtained.
	// Default: 0 (no retry)
	RetryCount int

	// RetryDelay is the delay between retries.
	// Default: 100ms
	RetryDelay time.Duration
}

// Obtain tries to obtain a new lock using a key with the given TTL.
// Returns ErrNotObtained if the lock cannot be obtained.
func (l *Locker) Obtain(ctx context.Context, key string, ttl time.Duration, opts *ObtainOptions) (*Lock, error) {
	value, err := randomToken()
	if err != nil {
		return nil, err
	}

	retryCount := 0
	retryDelay := 100 * time.Millisecond
	if opts != nil {
		retryCount = opts.RetryCount
		if opts.RetryDelay > 0 {
			retryDelay = opts.RetryDelay
		}
	}

	ttlMs := int64(ttl / time.Millisecond)

	for attempt := 0; attempt <= retryCount; attempt++ {
		ok, err := l.tryObtain(ctx, key, value, ttlMs)
		if err != nil {
			return nil, err
		}
		if ok {
			return &Lock{locker: l, key: key, value: value}, nil
		}

		if attempt < retryCount {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
			}
		}
	}

	return nil, ErrNotObtained
}

func (l *Locker) tryObtain(ctx context.Context, key, value string, ttlMs int64) (bool, error) {
	result, err := luaObtain.Run(ctx, l.client, []string{key}, value, ttlMs).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result == "OK", nil
}

// Key returns the redis key used by the lock.
func (l *Lock) Key() string {
	return l.key
}

// Token returns the token value set by the lock.
func (l *Lock) Token() string {
	return l.value
}

// Release manually releases the lock.
// Returns ErrLockNotHeld if the lock is not held.
func (l *Lock) Release(ctx context.Context) error {
	if l == nil {
		return ErrLockNotHeld
	}

	result, err := luaRelease.Run(ctx, l.locker.client, []string{l.key}, l.value).Int64()
	if errors.Is(err, redis.Nil) {
		return ErrLockNotHeld
	}
	if err != nil {
		return err
	}
	if result != 1 {
		return ErrLockNotHeld
	}
	return nil
}

// Refresh extends the lock with a new TTL.
// Returns ErrNotObtained if refresh is unsuccessful.
func (l *Lock) Refresh(ctx context.Context, ttl time.Duration) error {
	if l == nil {
		return ErrLockNotHeld
	}

	ttlMs := int64(ttl / time.Millisecond)
	result, err := luaRefresh.Run(ctx, l.locker.client, []string{l.key}, l.value, ttlMs).Int64()
	if err != nil {
		return err
	}
	if result != 1 {
		return ErrNotObtained
	}
	return nil
}

// TTL returns the remaining time-to-live. Returns 0 if the lock has expired.
func (l *Lock) TTL(ctx context.Context) (time.Duration, error) {
	if l == nil {
		return 0, ErrLockNotHeld
	}

	// First check if we still own the lock
	val, err := l.locker.client.Get(ctx, l.key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if val != l.value {
		return 0, nil
	}

	// Get TTL
	ttl, err := l.locker.client.PTTL(ctx, l.key).Result()
	if err != nil {
		return 0, err
	}
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
