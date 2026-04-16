package resource

// ProfileResource 管理员信息
type ProfileResource struct {
	ID          int    `json:"id" doc:"Admin ID"`
	Email       string `json:"email" doc:"Admin email"`
	DisplayName string `json:"display_name" doc:"Display name"`
	Status      string `json:"status" doc:"Account status"`
	CreatedAt   string `json:"created_at" doc:"Creation time"`
}

type LoginResource struct {
	Admin       *ProfileResource `json:"admin,omitempty" doc:"Admin information"`
	AccessToken string           `json:"access_token,omitempty" doc:"JWT access token"`
	TokenType   string           `json:"token_type,omitempty" doc:"Token type (Bearer)"`
	ExpiresIn   int64            `json:"expires_in,omitempty" doc:"Access token expiration in seconds"`
}
