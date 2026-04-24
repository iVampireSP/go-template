package serve

import (
	"errors"
	"io"

	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/i18n"
	"github.com/iVampireSP/go-template/pkg/foundation/tracing"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/iVampireSP/go-template/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// Api runs the User API server.
func (s *Serve) Api(cmd *cobra.Command) error {
	klog.SetOutput(io.Discard)
	klog.LogToStderr(false)

	if !config.Bool("http.api.enabled", true) {
		return errors.New("API server is not enabled in configuration")
	}

	tp, err := tracing.GetService("cloud-user-api")
	if err != nil {
		return err
	}
	defer tracing.ShutdownWithTimeout(tp)

	serverCfg := httpserver.DefaultConfig("User API", version.Version)
	serverCfg.Description = "Leaflow Cloud User API"
	serverCfg.Host = config.String("http.api.host", "0.0.0.0")
	serverCfg.Port = config.Int("http.api.port", 8080)
	serverCfg.ServerURL = config.String("discovery.user_api_base_url", "")

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
	server := httpserver.New("User API", version.Version,
		httpserver.WithHTTP(serverCfg, middlewares...),
		httpserver.WithMetrics(metricsCfg),
	)

	server.Router.Get("/.well-known/jwks.json", s.jwks.GetJWKS)
	server.Router.Get("/.well-known/discovery.json", s.discovery.GetUserDiscovery)
	server.Router.Get("/.well-known/openid-configuration", s.discovery.GetUserOpenIDConfiguration)

	s.userRouter.Register(server.API, server.Router)

	return server.Run(nil)
}
