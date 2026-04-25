package serve

import (
	adminroutes "github.com/iVampireSP/go-template/internal/api/admin/v1/route"
	userroutes "github.com/iVampireSP/go-template/internal/api/user/v1/route"
	wellknownhandler "github.com/iVampireSP/go-template/internal/api/wellknown/handler"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
	"github.com/iVampireSP/go-template/pkg/foundation/jwt"
	"github.com/spf13/cobra"
)

// Serve holds dependencies for all API servers.
type Serve struct {
	app       *container.Application
	jwt       *jwt.JWT
	jwks      *wellknownhandler.JWKSHandler
	discovery *wellknownhandler.DiscoveryHandler

	// admin API
	adminRouter *adminroutes.Router

	// user API
	userRouter *userroutes.Router
}

// NewServe declares serve command dependencies.
func NewServe(app *container.Application) *Serve {
	return &Serve{app: app}
}

// Command constructs the serve cobra command tree.
func (s *Serve) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "serve", Short: "Start API servers",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return s.app.Invoke(func(
				j *jwt.JWT,
				jwks *wellknownhandler.JWKSHandler,
				discovery *wellknownhandler.DiscoveryHandler,
				adminRouter *adminroutes.Router,
				userRouter *userroutes.Router,
			) {
				s.jwt = j
				s.jwks = jwks
				s.discovery = discovery
				s.adminRouter = adminRouter
				s.userRouter = userRouter
			})
		},
	}

	adminCmd := &cobra.Command{Use: "admin", Short: "Start Admin API server",
		RunE: func(c *cobra.Command, _ []string) error { return s.Admin(c) }}
	apiCmd := &cobra.Command{Use: "api", Short: "Start User API server",
		RunE: func(c *cobra.Command, _ []string) error { return s.Api(c) }}

	cmd.AddCommand(adminCmd, apiCmd)
	return cmd
}
