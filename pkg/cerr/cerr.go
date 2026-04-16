// Package cerr provides a unified error handling system for the application.
// Errors defined here implement huma.StatusError and integrate with Huma's RFC 9457 error responses.
// The package name "cerr" (cloud error) avoids conflict with the standard library "errors".
package cerr

import (
	"errors"
	"fmt"
)

// Error represents a business error that can be directly converted to an HTTP response.
// It implements huma.StatusError interface.
type Error struct {
	// Status is the HTTP status code
	Status int
	// Code is the error code for i18n, e.g., "REGION_NOT_FOUND"
	// This should be the key for frontend i18n lookup
	Code string
	// Message is the default human-readable message (fallback if i18n not available)
	Message string
	// Cause is the underlying error (for error chain tracing)
	Cause error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is reports whether the error matches the target.
func (e *Error) Is(target error) bool {
	var t *Error
	if errors.As(target, &t) {
		// Match by code if both have codes
		if e.Code != "" && t.Code != "" {
			return e.Code == t.Code
		}
		// Match by status and message
		return e.Status == t.Status && e.Message == t.Message
	}
	return false
}

// clone creates a copy of the Error to avoid modifying global variables.
func (e *Error) clone() *Error {
	return &Error{
		Status:  e.Status,
		Code:    e.Code,
		Message: e.Message,
		Cause:   e.Cause,
	}
}

// WithCode sets the error code and returns a new copy.
func (e *Error) WithCode(code string) *Error {
	newErr := e.clone()
	newErr.Code = code
	return newErr
}

// WithMessage sets the error message and returns a new copy.
func (e *Error) WithMessage(message string) *Error {
	newErr := e.clone()
	newErr.Message = message
	return newErr
}

// WIthMessages formats the error message with the given arguments and returns a new copy.
// This is a convenience method that combines Format and WithMessage.
func (e *Error) WithMessagef(format string, args ...any) *Error {
	return e.WithMessage(fmt.Sprintf(format, args...))
}

// WithCause sets the underlying error and returns a new copy.
// If the error already has a cause, the new cause is joined with the existing one.
func (e *Error) WithCause(cause error) *Error {
	newErr := e.clone()
	if newErr.Cause == nil {
		newErr.Cause = cause
	} else {
		newErr.Cause = errors.Join(newErr.Cause, cause)
	}
	return newErr
}

// Format formats the error message with the given arguments and returns a new copy.
// Use this when the error message contains format specifiers like %s, %d, etc.
// Example:
//
//	var ErrMissingField = cerr.BadRequest("missing required field: %s").WithCode("MISSING_FIELD")
//	return ErrMissingField.Format("username")
func (e *Error) Format(args ...any) *Error {
	newErr := e.clone()
	newErr.Message = fmt.Sprintf(e.Message, args...)
	return newErr
}

// -----------------------------------------------------------------------------
// Constructors for common HTTP errors
// -----------------------------------------------------------------------------

// New creates a new error with the given status and message.
func New(status int, message string) *Error {
	return &Error{
		Status:  status,
		Message: message,
	}
}

// BadRequest creates a 400 Bad Request error.
func BadRequest(message string) *Error {
	return New(400, message)
}

// Unauthorized creates a 401 Unauthorized error.
func Unauthorized(message string) *Error {
	return New(401, message)
}

// Forbidden creates a 403 Forbidden error.
func Forbidden(message string) *Error {
	return New(403, message)
}

// NotFound creates a 404 Not Found error.
func NotFound(message string) *Error {
	return New(404, message)
}

// Conflict creates a 409 Conflict error.
func Conflict(message string) *Error {
	return New(409, message)
}

// Gone creates a 410 Gone error.
func Gone(message string) *Error {
	return New(410, message)
}

// PreconditionFailed creates a 412 Precondition Failed error.
func PreconditionFailed(message string) *Error {
	return New(412, message)
}

// UnprocessableEntity creates a 422 Unprocessable Entity error.
func UnprocessableEntity(message string) *Error {
	return New(422, message)
}

// TooManyRequests creates a 429 Too Many Requests error.
func TooManyRequests(message string) *Error {
	return New(429, message)
}

// Internal creates a 500 Internal Server Error.
func Internal(message string) *Error {
	return New(500, message)
}

// ServiceUnavailable creates a 503 Service Unavailable error.
func ServiceUnavailable(message string) *Error {
	return New(503, message)
}

// Validation creates a 422 Unprocessable Entity error with validation details.
// Use huma.ErrorDetail for field errors to ensure RFC 9457 compliance.
func Validation(message string) *Error {
	return New(422, message).WithCode("VALIDATION_FAILED")
}

// -----------------------------------------------------------------------------
// Helper functions
// -----------------------------------------------------------------------------

// As is a convenience wrapper around errors.As for cerr.Error.
func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// IsNotFound checks if the error is a 404 Not Found error.
func IsNotFound(err error) bool {
	if e, ok := As(err); ok {
		return e.Status == 404
	}
	return false
}

// IsUnauthorized checks if the error is a 401 Unauthorized error.
func IsUnauthorized(err error) bool {
	if e, ok := As(err); ok {
		return e.Status == 401
	}
	return false
}

// IsForbidden checks if the error is a 403 Forbidden error.
func IsForbidden(err error) bool {
	if e, ok := As(err); ok {
		return e.Status == 403
	}
	return false
}

// IsClientError checks if the error is a 4xx client error.
func IsClientError(err error) bool {
	if e, ok := As(err); ok {
		return e.Status >= 400 && e.Status < 500
	}
	return false
}

// IsServerError checks if the error is a 5xx server error.
func IsServerError(err error) bool {
	if e, ok := As(err); ok {
		return e.Status >= 500
	}
	return false
}

// GetStatus returns the HTTP status code of the error, or 500 if not a cerr.Error.
func GetStatus(err error) int {
	if e, ok := As(err); ok {
		return e.Status
	}
	return 500
}

// GetCode returns the error code, or empty string if not a cerr.Error.
func GetCode(err error) string {
	if e, ok := As(err); ok {
		return e.Code
	}
	return ""
}
