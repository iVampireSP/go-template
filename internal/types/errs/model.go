package errs

import "errors"

var (
	ErrPageNotFound = errors.New("page not found")
	ErrNotFound     = errors.New("not found")
)
