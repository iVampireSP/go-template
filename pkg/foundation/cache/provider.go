package cache

import (
	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
	"github.com/redis/go-redis/v9"
)

type Provider struct {
	app *container.Application
}

func NewProvider(app *container.Application) *Provider {
	return &Provider{app: app}
}

func (p *Provider) Register() {
	p.app.Singleton(NewDefaultRedisConfig)
	p.app.Singleton(NewCache)
	p.app.OnShutdown(func() error {
		if c, err := container.Make[redis.UniversalClient](p.app.Container); err == nil && c != nil {
			return c.Close()
		}
		return nil
	})
}

func (p *Provider) Boot() {}

// NewDefaultRedisConfig returns a RedisConfig populated from the application config.
func NewDefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Host:         config.String("redis.host", "localhost"),
		Port:         config.Int("redis.port", 6379),
		Password:     config.String("redis.password"),
		DB:           config.Int("redis.db", 0),
		ClusterAddrs: config.String("redis.cluster_addrs"),
	}
}
