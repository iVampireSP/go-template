package errs

import "errors"

var (
	EmptyResponse          = errors.New("empty response")
	ErrInternalServerError = errors.New("there was a server error, but we have logged this request for further investigation")
	ErrValidationError     = errors.New("unprocessable Entity")
	RouteNotFound          = errors.New("route not found")
)
