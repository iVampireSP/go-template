package user

import "github.com/iVampireSP/go-template/pkg/cerr"

var (
	ErrUserNotFound       = cerr.NotFound("user not found").WithCode("USER_NOT_FOUND")
	ErrUserInactive       = cerr.Forbidden("user account is inactive").WithCode("USER_INACTIVE")
	ErrUserLocked         = cerr.Forbidden("user account is locked, please try again later").WithCode("USER_LOCKED")
	ErrEmailExists        = cerr.Conflict("email is already registered").WithCode("EMAIL_EXISTS")
	ErrInvalidPassword    = cerr.Unauthorized("invalid password").WithCode("INVALID_PASSWORD")
	ErrEmailNotVerified   = cerr.Forbidden("email is not verified").WithCode("EMAIL_NOT_VERIFIED")
	ErrInvalidDisplayName = cerr.BadRequest("display name is required").WithCode("INVALID_DISPLAY_NAME")
	ErrInvalidToken       = cerr.Unauthorized("invalid token").WithCode("INVALID_TOKEN")
	ErrTokenExpired       = cerr.Unauthorized("token has expired").WithCode("TOKEN_EXPIRED")
)
