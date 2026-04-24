package tracing

import "errors"

var (
	ErrTracingDisabled       = errors.New("tracing is disabled")
	ErrTracingNotInitialized = errors.New("tracing is not initialized")
	ErrInvalidTraceID        = errors.New("invalid trace id")
)
