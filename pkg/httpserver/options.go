package httpserver

// Option configures Server components.
type Option func(*Server)

// WithHTTP enables and configures the HTTP API component.
func WithHTTP(cfg *Config, middlewares ...Middleware) Option {
	return func(s *Server) {
		if cfg == nil {
			cfg = DefaultConfig(s.name, s.version)
		}
		if cfg.Name == "" {
			cfg.Name = s.name
		}
		if cfg.Version == "" {
			cfg.Version = s.version
		}
		s.config = cfg
		s.httpEnabled = true
		s.setupHTTP(middlewares...)
	}
}

// WithMetrics enables and configures the metrics component.
func WithMetrics(cfg MetricsConfig) Option {
	return func(s *Server) {
		if cfg.Name == "" {
			cfg.Name = s.name
		}
		if cfg.Host == "" {
			cfg.Host = "0.0.0.0"
		}
		if cfg.Port == 0 {
			cfg.Port = 9090
		}
		s.metricsConfig = cfg
		s.metricsEnabled = cfg.Enabled
	}
}
