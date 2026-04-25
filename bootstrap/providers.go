package bootstrap

import (
	"github.com/iVampireSP/go-template/internal/providers"
	"github.com/iVampireSP/go-template/pkg/foundation/bus"
	"github.com/iVampireSP/go-template/pkg/foundation/cache"
	"github.com/iVampireSP/go-template/pkg/foundation/cron"
	"github.com/iVampireSP/go-template/pkg/foundation/email"
	"github.com/iVampireSP/go-template/pkg/foundation/jwt"
	"github.com/iVampireSP/go-template/pkg/foundation/keystore"
	"github.com/iVampireSP/go-template/pkg/foundation/lock"
	"github.com/iVampireSP/go-template/pkg/foundation/orm"
	"github.com/iVampireSP/go-template/pkg/foundation/queue"
	"github.com/iVampireSP/go-template/pkg/foundation/schedule"
	"github.com/iVampireSP/go-template/pkg/foundation/tracing"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/iVampireSP/go-template/pkg/logger"
)

// Providers returns all service provider constructors in registration order.
// Foundation providers are registered first, followed by application providers.
func Providers() []any {
	return []any{
		// Foundation providers
		logger.NewProvider,
		orm.NewProvider,
		cache.NewProvider,
		keystore.NewProvider,
		jwt.NewProvider,
		lock.NewProvider,
		cron.NewProvider,
		tracing.NewProvider,
		httpserver.NewProvider,
		email.NewProvider,
		queue.NewProvider,
		bus.NewProvider,
		schedule.NewProvider,

		// Application providers
		providers.NewAppServiceProvider,
		providers.NewConsoleServiceProvider,
		providers.NewRouteServiceProvider,
		providers.NewEventServiceProvider,
		providers.NewQueueServiceProvider,
		providers.NewScheduleServiceProvider,
	}
}
