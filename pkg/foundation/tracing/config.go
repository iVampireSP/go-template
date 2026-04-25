package tracing

import "time"

// Config holds OpenTelemetry tracing configuration.
type Config struct {
	Enabled       bool
	Endpoint      string
	SampleRatio   float64
	QueryURL      string
	QueryUsername  string
	QueryPassword string
	QueryTimeout  time.Duration
}
