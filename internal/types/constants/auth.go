package constants

import (
	"errors"
	"go-template/internal/types/auth"
)

const (
	AuthHeader = "Authorization"
	AuthPrefix = "Bearer"

	// AnonymousUser 调试模式下的用户
	AnonymousUser auth.Id = "anonymous"

	AuthMiddlewareKey               = "auth.user"
	AuthAssistantShareMiddlewareKey = "auth.assistant.share"
)

var (
	ErrNotValidToken      = errors.New("JWT not valid")
	ErrJWTFormatError     = errors.New("JWT format error")
	ErrNotBearerType      = errors.New("not bearer token")
	ErrEmptyResponse      = errors.New("empty response")
	ErrTokenError         = errors.New("token type error")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrAudienceNotAllowed = errors.New("audience not allowed")

	ErrNotYourResource  = errors.New("this resource not yours")
	ErrPermissionDenied = errors.New("permission denied")
)
