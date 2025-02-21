package errs

var (
	ErrNotFound            = NewNotFoundError("not found")
	ErrUnauthorized        = NewUnauthorizedError("unauthorized")
	ErrBadRequest          = NewBadRequestError("bad request")
	ErrInternalServerError = NewInternalServerError("there was a server error, but we have logged this request for further investigation")
)
