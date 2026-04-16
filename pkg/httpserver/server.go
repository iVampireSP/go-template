package httpserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/yaml.v3"

	"github.com/iVampireSP/go-template/pkg/cerr"
	"github.com/iVampireSP/go-template/pkg/json"
	"github.com/iVampireSP/go-template/pkg/logger"
)

// Config holds the HTTP server configuration.
type Config struct {
	Name        string
	Version     string
	Description string
	Host        string
	Port        int
	DocsPath    string
	OpenAPIPath string
	ServerURL   string // For OpenAPI documentation

	// CORS settings
	CORSEnabled          bool
	CORSAllowedOrigins   []string
	CORSAllowedMethods   []string
	CORSAllowedHeaders   []string
	CORSAllowCredentials bool
	CORSExposedHeaders   []string
	CORSMaxAge           int

	// Metrics server settings
	MetricsEnabled bool
	MetricsHost    string
	MetricsPort    int

	// Graceful shutdown settings
	ShutdownTimeout time.Duration

	// Chi middlewares to apply before routes are registered
	// This is necessary because chi requires all middlewares to be defined before routes
	ChiMiddlewares []func(http.Handler) http.Handler
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(name, version string) *Config {
	return &Config{
		Name:                 name,
		Version:              version,
		Description:          name,
		Host:                 "0.0.0.0",
		Port:                 8080,
		DocsPath:             "/docs",
		OpenAPIPath:          "/openapi",
		CORSEnabled:          true,
		CORSAllowedOrigins:   []string{"*"},
		CORSAllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		CORSAllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		CORSAllowCredentials: true,
		CORSExposedHeaders:   []string{"X-Trace-Id"},
		CORSMaxAge:           86400,
		MetricsEnabled:       false,
		MetricsHost:          "0.0.0.0",
		MetricsPort:          9090,
		ShutdownTimeout:      30 * time.Second,
	}
}

// Server wraps chi router and huma API.
type Server struct {
	Router         *chi.Mux
	API            API
	config         *Config
	server         *http.Server
	metricsServer  *MetricsServer
	metricsConfig  MetricsConfig
	metricsEnabled bool
	httpEnabled    bool
	name           string
	version        string
}

// Middleware is a chi middleware function.
type Middleware = func(http.Handler) http.Handler

// New creates a new server with optional components.
func New(name, version string, opts ...Option) *Server {
	s := &Server{
		name:    name,
		version: version,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// ServiceEndpoints contains available service endpoints.
type ServiceEndpoints struct {
	Docs    string `json:"docs" doc:"API documentation path"`
	OpenAPI string `json:"openapi" doc:"OpenAPI specification path"`
	Health  string `json:"health" doc:"Health check endpoint path"`
}

// StatusOutput is the response for the status endpoint.
type StatusOutput struct {
	Body struct {
		Name        string           `json:"name" doc:"Service name"`
		Version     string           `json:"version" doc:"Service version"`
		Description string           `json:"description" doc:"Service description"`
		Status      string           `json:"status" doc:"Service status"`
		Endpoints   ServiceEndpoints `json:"endpoints" doc:"Available endpoints"`
	}
}

// HealthOutput is the response for the health check endpoint.
type HealthOutput struct {
	Body struct {
		Status  string `json:"status" doc:"Health status"`
		Service string `json:"service" doc:"Service name"`
		Version string `json:"version" doc:"Service version"`
	}
}

// registerStatusRoutes registers status and health check routes.
func registerStatusRoutes(api API, cfg *Config) {
	// Status endpoint
	GET(api, "/", Operation{
		ID:          "status",
		Summary:     "Service status",
		Description: "Get service name, version, status and available endpoints",
		Tags:        []string{"Status"},
	}, func(ctx context.Context, input *struct{}) (*StatusOutput, error) {
		return &StatusOutput{
			Body: struct {
				Name        string           `json:"name" doc:"Service name"`
				Version     string           `json:"version" doc:"Service version"`
				Description string           `json:"description" doc:"Service description"`
				Status      string           `json:"status" doc:"Service status"`
				Endpoints   ServiceEndpoints `json:"endpoints" doc:"Available endpoints"`
			}{
				Name:        cfg.Name,
				Version:     cfg.Version,
				Description: cfg.Description,
				Status:      "running",
				Endpoints: ServiceEndpoints{
					Docs:    cfg.DocsPath,
					OpenAPI: cfg.OpenAPIPath,
					Health:  "/healthz",
				},
			},
		}, nil
	})

	// Health check endpoint
	GET(api, "/healthz", Operation{
		ID:          "health",
		Summary:     "Health check",
		Description: "Check if the service is healthy",
		Tags:        []string{"Status"},
	}, func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		return &HealthOutput{
			Body: struct {
				Status  string `json:"status" doc:"Health status"`
				Service string `json:"service" doc:"Service name"`
				Version string `json:"version" doc:"Service version"`
			}{
				Status:  "healthy",
				Service: cfg.Name,
				Version: cfg.Version,
			},
		}, nil
	})
}

// Start starts the HTTP server (non-blocking).
// Use Run() for a blocking call with graceful shutdown.
func (s *Server) Start() error {
	if s.metricsEnabled {
		if s.metricsServer == nil {
			s.metricsServer = newMetricsServer(s.metricsConfig)
		}
		s.metricsServer.Start()
	}

	if !s.httpEnabled {
		return nil
	}

	// Register status routes before starting
	registerStatusRoutes(s.API, s.config)

	addr := s.Addr()
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.Router,
	}

	logger.Info("listening", "name", s.config.Name, "on", addr)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// StartMetrics starts the metrics server if enabled (non-blocking).
func (s *Server) StartMetrics() {
	if !s.metricsEnabled {
		return
	}
	if s.metricsServer == nil {
		s.metricsServer = newMetricsServer(s.metricsConfig)
	}
	s.metricsServer.Start()
}

// Run starts the server with metrics (if enabled) and handles graceful shutdown.
// The cleanup function is called during shutdown (e.g., to stop bootstrap app).
// This is a blocking call that returns when the server is fully stopped.
func (s *Server) Run(cleanup func(context.Context) error) error {
	serverErr := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Shutting ...", "down", s.name)
	case err := <-serverErr:
		if err != nil {
			logger.Error("", "name", s.name, "error", err)
			return err
		}
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout())
	defer cancel()

	s.shutdownWithContext(ctx)

	// 3. Run cleanup function (e.g., stop bootstrap app)
	if cleanup != nil {
		if err := cleanup(ctx); err != nil {
			logger.Error("Cleanup", "error", err)
			return err
		}
	}

	logger.Info("stopped", "name", s.name)
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if !s.httpEnabled {
		return nil
	}
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// ShutdownMetricsWithTimeout gracefully shuts down the metrics server.
func (s *Server) ShutdownMetricsWithTimeout() {
	if s.metricsServer == nil {
		return
	}
	s.metricsServer.ShutdownWithTimeout()
}

// ShutdownWithTimeout gracefully shuts down enabled components.
func (s *Server) ShutdownWithTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout())
	defer cancel()

	s.shutdownWithContext(ctx)
}

func (s *Server) shutdownWithContext(ctx context.Context) {
	// 1. Shutdown API server
	if err := s.Shutdown(ctx); err != nil {
		logger.Error("shutdown", "name", s.name, "error", err)
	}

	// 2. Shutdown metrics server
	if s.metricsServer != nil {
		if err := s.metricsServer.Shutdown(ctx); err != nil {
			logger.Error("metrics server shutdown", "name", s.name, "error", err)
		}
	}
}

func (s *Server) shutdownTimeout() time.Duration {
	shutdownTimeout := 30 * time.Second
	if s.config != nil && s.config.ShutdownTimeout != 0 {
		shutdownTimeout = s.config.ShutdownTimeout
	}
	if s.metricsConfig.ShutdownTimeout != 0 {
		shutdownTimeout = s.metricsConfig.ShutdownTimeout
	}
	return shutdownTimeout
}

// Addr returns the server address.
func (s *Server) Addr() string {
	if s.config == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
}

// MetricsAddr returns the metrics server address.
func (s *Server) MetricsAddr() string {
	if !s.metricsEnabled {
		return ""
	}
	return fmt.Sprintf("%s:%d", s.metricsConfig.Host, s.metricsConfig.Port)
}

// ErrorModel is a custom error model that ensures empty arrays are serialized as []
// instead of being omitted (unlike huma.ErrorModel which uses omitempty).
// It also includes a Code field for i18n support.
type ErrorModel struct {
	Type   string              `json:"type"`
	Title  string              `json:"title"`
	Status int                 `json:"status"`
	Detail string              `json:"detail"`
	Code   string              `json:"code"`   // i18n error code, e.g., "REFRESH_TOKEN_NOT_FOUND"
	Errors []*huma.ErrorDetail `json:"errors"` // No omitempty - always include
}

// Error implements the error interface.
func (e *ErrorModel) Error() string {
	return e.Detail
}

// GetStatus implements huma.StatusError interface.
func (e *ErrorModel) GetStatus() int {
	return e.Status
}

// jsonErrorHandler returns an http.HandlerFunc that writes a JSON error response.
// Used for NotFound and MethodNotAllowed handlers.
func jsonErrorHandler(status int, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		model := &ErrorModel{
			Type:   "about:blank",
			Title:  http.StatusText(status),
			Status: status,
			Detail: message,
			Errors: make([]*huma.ErrorDetail, 0),
		}

		json.NewEncoder(w).Encode(model)
	}
}

// setupErrorHandler configures Huma to properly handle cerr.Error.
// Since cerr.Error no longer implements huma.StatusError, Huma will always
// call this NewError function, allowing us to convert cerr.Error to huma.ErrorModel.
//
// For 4xx: returns error detail and code for i18n
// For 5xx: logs full error, returns generic message to client
func setupErrorHandler() {
	huma.NewError = func(status int, message string, errs ...error) huma.StatusError {
		// Skip logging for OpenAPI schema generation (empty message and no errors)
		if message == "" && len(errs) == 0 {
			return &ErrorModel{
				Type:   "about:blank",
				Title:  http.StatusText(status),
				Status: status,
				Detail: message,
				Errors: make([]*huma.ErrorDetail, 0),
			}
		}

		// Default to 500 if status is not set (unknown errors should be treated as server errors)
		if status == 0 {
			status = http.StatusInternalServerError
		}

		model := &ErrorModel{
			Type:   "about:blank",
			Title:  http.StatusText(status),
			Status: status,
			Detail: message,
			Errors: make([]*huma.ErrorDetail, 0), // Empty array instead of nil
		}

		for _, err := range errs {
			if err == nil {
				continue
			}

			// Check if it's a cerr.Error (anywhere in the error chain)
			if ce, ok := errors.AsType[*cerr.Error](err); ok {
				if ce.Status != 0 {
					model.Status = ce.Status
					model.Title = http.StatusText(ce.Status)
				}
				if ce.Message != "" {
					model.Detail = ce.Message
				}
				if ce.Code != "" {
					model.Code = ce.Code
				}

				// For 5xx: log full error chain and return generic message
				if model.Status >= 500 {
					if errors.Is(err, context.Canceled) {
						logger.Debug("[ ] ( )", "code", ce.Code, "message", ce.Message, "canceled", unwrapAll(err))
					} else {
						logger.Error("[ ] ( )", "code", ce.Code, "message", ce.Message, "cause", unwrapAll(err))
					}
					model.Detail = "An internal server error occurred"
					model.Code = "INTERNAL_ERROR"
				}

				// Debug-level logging for all errors (suppressed in non-debug mode)
				logger.Debug("[ ] ( )", "code", ce.Code, "status", model.Status, "message", ce.Message, "cause", unwrapAll(err))
				continue
			}

			// Check if it's a huma.ErrorDetail (validation error)
			if detail, ok := errors.AsType[*huma.ErrorDetail](err); ok {
				model.Errors = append(model.Errors, detail)
				continue
			}

			// Unknown error
			if model.Status >= 500 {
				if errors.Is(err, context.Canceled) {
					logger.Debug("request canceled by", "client", err)
				} else {
					logger.Error("Unhandled server ( )", "error", err, "chain", unwrapAll(err))
				}
				model.Detail = "An internal server error occurred"
			}
			logger.Debug("", "error", model.Status, "err", err)
		}

		return model
	}
}

// debugRequestLogger is a chi middleware that logs all requests and response status codes at debug level.
func debugRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			status := ww.Status()
			logger.Debug("", "method", r.Method, "request_uri", r.RequestURI, "status", status, "since", time.Since(start), "remote_addr", r.RemoteAddr)
		}()

		next.ServeHTTP(ww, r)
	})
}

func (s *Server) setupHTTP(middlewares ...Middleware) {
	if s.config == nil {
		s.config = DefaultConfig(s.name, s.version)
	}

	// Create chi router
	r := chi.NewRouter()

	// OpenTelemetry HTTP tracing middleware
	r.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, s.config.Name,
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
		)
	})

	// Return trace ID in response header so clients can reference it for debugging.
	// otelhttp (outer) already created the span, so it's in the context here.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if spanCtx := trace.SpanFromContext(r.Context()).SpanContext(); spanCtx.HasTraceID() {
				w.Header().Set("X-Trace-Id", spanCtx.TraceID().String())
			}
			next.ServeHTTP(w, r)
		})
	})

	// CORS middleware
	if s.config.CORSEnabled {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   s.config.CORSAllowedOrigins,
			AllowedMethods:   s.config.CORSAllowedMethods,
			AllowedHeaders:   s.config.CORSAllowedHeaders,
			AllowCredentials: s.config.CORSAllowCredentials,
			ExposedHeaders:   s.config.CORSExposedHeaders,
			MaxAge:           s.config.CORSMaxAge,
		}))
	}

	// Apply config-provided middlewares (must be registered before any routes)
	for _, mw := range s.config.ChiMiddlewares {
		r.Use(mw)
	}

	// User-provided middlewares (must be registered before any routes)
	for _, m := range middlewares {
		r.Use(m)
	}

	// Debug-level request logger
	r.Use(debugRequestLogger)

	// Custom error handlers for unmatched routes
	r.NotFound(jsonErrorHandler(http.StatusNotFound, "The requested resource was not found"))
	r.MethodNotAllowed(jsonErrorHandler(http.StatusMethodNotAllowed, "The requested method is not allowed for this resource"))

	// Setup cerr error handler
	setupErrorHandler()

	// Create Huma API config
	humaConfig := huma.DefaultConfig(s.config.Name, s.config.Version)
	humaConfig.Info.Description = s.config.Description
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT Bearer token authentication",
		},
	}
	humaConfig.DocsPath = s.config.DocsPath
	humaConfig.OpenAPIPath = s.config.OpenAPIPath

	if s.config.ServerURL != "" {
		humaConfig.Servers = []*huma.Server{
			{URL: s.config.ServerURL, Description: s.config.Name},
		}
	}

	// Override JSON format for better performance
	sonicFormat := huma.Format{
		Marshal: func(w io.Writer, v any) error {
			return json.NewEncoder(w).Encode(v)
		},
		Unmarshal: json.Unmarshal,
	}
	humaConfig.Formats["application/json"] = sonicFormat
	humaConfig.Formats["json"] = sonicFormat

	// Register YAML format for content negotiation
	humaConfig.Formats["application/yaml"] = huma.Format{
		Marshal: func(w io.Writer, v any) error {
			enc := yaml.NewEncoder(w)
			enc.SetIndent(2)
			return enc.Encode(v)
		},
		Unmarshal: yaml.Unmarshal,
	}
	humaConfig.Formats["yaml"] = humaConfig.Formats["application/yaml"]

	// Create Huma API
	api := humachi.New(r, humaConfig)

	s.Router = r
	s.API = WrapAPI(api)
}

// unwrapAll returns a string representation of the full error chain.
func unwrapAll(err error) string {
	if err == nil {
		return "<nil>"
	}

	var chain []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		chain = append(chain, fmt.Sprintf("%T: %v", e, e))
	}

	if len(chain) == 1 {
		return chain[0]
	}
	return fmt.Sprintf("[%s]", strings.Join(chain, " -> "))
}
