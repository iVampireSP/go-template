package constants

import "errors"

var (
	ErrInternalServerError = errors.New("there was a server error, but we have logged this request for further investigation")
)
