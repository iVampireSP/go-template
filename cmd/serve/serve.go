package serve

import (
	adminroutes "github.com/iVampireSP/go-template/internal/api/admin/v1/route"
	userroutes "github.com/iVampireSP/go-template/internal/api/user/v1/route"
	wellknownhandler "github.com/iVampireSP/go-template/internal/api/wellknown/handler"
	"github.com/iVampireSP/go-template/pkg/foundation/jwt"
	"github.com/spf13/cobra"
)

// Serve holds dependencies for all API servers.
type Serve struct {
	jwt       *jwt.JWT
	jwks      *wellknownhandler.JWKSHandler
	discovery *wellknownhandler.DiscoveryHandler

	// admin API
	adminRouter *adminroutes.Router

	// user API
	userRouter *userroutes.Router
}

// NewServe declares serve command dependencies.
func NewServe(
	jwt *jwt.JWT,
	jwks *wellknownhandler.JWKSHandler,
	discovery *wellknownhandler.DiscoveryHandler,
	adminRouter *adminroutes.Router,
	userRouter *userroutes.Router,
) *Serve {
	return &Serve{
		jwt:         jwt,
		jwks:        jwks,
		discovery:   discovery,
		adminRouter: adminRouter,
		userRouter:  userRouter,
	}
}

// Command constructs the serve cobra command tree.
func (s *Serve) Command() *cobra.Command {
	cmd := &cobra.Command{Use: "serve", Short: "Start API servers"}

	adminCmd := &cobra.Command{Use: "admin", Short: "Start Admin API server"}
	apiCmd := &cobra.Command{Use: "api", Short: "Start User API server"}

	cmd.AddCommand(adminCmd, apiCmd)
	return cmd
}
