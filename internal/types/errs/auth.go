package errs

import "errors"

var (
	NotValidToken         = errors.New("JWT not valid")
	JWTFormatError        = errors.New("JWT format error")
	NotBearerType         = errors.New("not bearer token")
	TokenError            = errors.New("token type error")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrAudienceNotAllowed = errors.New("audience not allowed")

	ErrNotYourResource  = errors.New("this resource not yours")
	ErrPermissionDenied = errors.New("permission denied")
)
