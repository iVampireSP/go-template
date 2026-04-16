package response

import "github.com/iVampireSP/go-template/internal/api/admin/v1/resource"

// ProfileResponse is the Huma output for admin profile.
type ProfileResponse struct {
	Body resource.ProfileResource
}

func NewProfileResponse(profileResource resource.ProfileResource) *ProfileResponse {
	return &ProfileResponse{
		Body: profileResource,
	}
}

// LoginResponse is the Huma output for login.
type LoginResponse struct {
	Body resource.LoginResource
}

func NewLoginResponse(loginResource resource.LoginResource) *LoginResponse {
	return &LoginResponse{
		Body: loginResource,
	}
}
