package orm

import (
	"time"

	"github.com/iVampireSP/go-template/ent"
	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
)

type Provider struct {
	app *container.Application
}

func NewProvider(app *container.Application) *Provider {
	return &Provider{app: app}
}

func (p *Provider) Register() {
	p.app.Singleton(NewDefaultConfig)
	p.app.Singleton(NewORM)
	p.app.OnShutdown(func() error {
		if c, err := container.Make[*ent.Client](p.app.Container); err == nil && c != nil {
			return c.Close()
		}
		return nil
	})
}

func (p *Provider) Boot() {}

// NewDefaultConfig returns an ORM Config populated from the application config.
func NewDefaultConfig() Config {
	return Config{
		Host:            config.String("database.app.host", "localhost"),
		Port:            config.Int("database.app.port", 4000),
		User:            config.String("database.app.user", "root"),
		Password:        config.String("database.app.password"),
		Name:            config.String("database.app.name", "cloud"),
		MaxOpenConns:    config.Int("database.app.max_open_conns", 25),
		MaxIdleConns:    config.Int("database.app.max_idle_conns", 5),
		ConnMaxLifetime: time.Duration(config.Int("database.app.conn_max_lifetime_seconds", 300)) * time.Second,
		ConnMaxIdleTime: time.Duration(config.Int("database.app.conn_max_idle_time_seconds", 60)) * time.Second,
	}
}
