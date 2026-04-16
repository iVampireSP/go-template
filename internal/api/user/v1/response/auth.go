package response

import "github.com/iVampireSP/go-template/internal/api/user/v1/resource"

// TokenResponse is the Huma output for token issuance.
type TokenResponse struct {
	Body struct {
		User        *resource.ProfileResource `json:"user,omitempty" doc:"用户信息"`
		AccessToken string                    `json:"access_token,omitempty" doc:"访问令牌"`
		TokenType   string                    `json:"token_type,omitempty" doc:"令牌类型"`
		ExpiresIn   int64                     `json:"expires_in,omitempty" doc:"过期时间（秒）"`
	}
}

// ProfileResponse is the Huma output for a single ProfileResource.
type ProfileResponse struct {
	Body resource.ProfileResource
}

func NewProfileResponse(profileResource resource.ProfileResource) *ProfileResponse {
	return &ProfileResponse{
		Body: profileResource,
	}
}
