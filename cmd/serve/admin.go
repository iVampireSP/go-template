package serve

import (
	"errors"

	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/i18n"
	"github.com/iVampireSP/go-template/pkg/foundation/tracing"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/iVampireSP/go-template/pkg/version"
	"github.com/spf13/cobra"
)

// Admin runs the Admin API server.
func (s *Serve) Admin(cmd *cobra.Command) error {
	if !config.Bool("http.admin_api.enabled", true) {
		return errors.New("admin server is not enabled in configuration")
	}

	tp, err := tracing.GetService("cloud-admin-api")
	if err != nil {
		return err
	}
	defer tracing.ShutdownWithTimeout(tp)

	serverCfg := httpserver.DefaultConfig("Admin API", version.Version)
	serverCfg.Description = "Leaflow Cloud Admin API"
	serverCfg.Host = config.String("http.admin_api.host", "0.0.0.0")
	serverCfg.Port = config.Int("http.admin_api.port", 8080)
	serverCfg.ServerURL = config.String("discovery.admin_api_base_url", "")

	serverCfg.CORSEnabled = config.Bool("http.cors.enabled", true)
	serverCfg.CORSAllowedOrigins = config.Strings("http.cors.allowed_origins", []string{"*"})
	serverCfg.CORSAllowedMethods = config.Strings("http.cors.allow_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	serverCfg.CORSAllowedHeaders = config.Strings("http.cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"})
	serverCfg.CORSAllowCredentials = config.Bool("http.cors.allow_credentials", true)
	serverCfg.CORSExposedHeaders = config.Strings("http.cors.expose_headers", []string{"X-Trace-Id"})
	serverCfg.CORSMaxAge = config.Int("http.cors.max_age", 86400)

	metricsCfg := httpserver.MetricsConfig{
		Enabled:         config.Bool("metrics.enabled", true),
		Host:            config.String("metrics.host", "0.0.0.0"),
		Port:            config.Int("metrics.port", 9090),
		ShutdownTimeout: serverCfg.ShutdownTimeout,
	}

	middlewares := append(httpserver.DefaultMiddlewares(), i18n.ChiMiddleware())
	server := httpserver.New("Admin API", version.Version,
		httpserver.WithHTTP(serverCfg, middlewares...),
		httpserver.WithMetrics(metricsCfg),
	)

	server.Router.Get("/.well-known/jwks.json", s.jwks.GetJWKS)
	server.Router.Get("/.well-known/discovery.json", s.discovery.GetAdminDiscovery)
	server.Router.Get("/.well-known/openid-configuration", s.discovery.GetAdminOpenIDConfiguration)

	s.adminRouter.Register(server.API, server.Router)

	return server.Run(nil)
}
