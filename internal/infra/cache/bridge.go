package cache

import (
	"github.com/iVampireSP/go-template/internal/infra/config"
	foundationcache "github.com/iVampireSP/go-template/pkg/foundation/cache"
	foundationlock "github.com/iVampireSP/go-template/pkg/foundation/lock"
	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Re-export foundation types.
type Locker = foundationlock.Locker
type Lock = foundationlock.Lock
type ObtainOptions = foundationlock.ObtainOptions

var (
	ErrNotObtained = foundationlock.ErrNotObtained
	ErrLockNotHeld = foundationlock.ErrLockNotHeld
	NewLocker      = foundationlock.NewLocker
)

func New() (redis.UniversalClient, *Locker) {
	_, _, sugar := logger.NewLogger(logger.Config{
		Level: config.String("log.level", "info"),
		Debug: config.Bool("app.debug", false),
	})

	client := foundationcache.New(foundationcache.RedisConfig{
		Host:         config.String("redis.host", "localhost"),
		Port:         config.Int("redis.port", 6379),
		Password:     config.String("redis.password"),
		DB:           config.Int("redis.db", 0),
		ClusterAddrs: config.String("redis.cluster_addrs"),
	}, sugar)

	locker := foundationlock.NewLocker(client)
	return client, locker
}
