package resource

import "github.com/iVampireSP/go-template/ent"

// ProfileResource 用户信息
type ProfileResource struct {
	ID            uint    `json:"id" doc:"用户ID"`
	Email         string  `json:"email" doc:"邮箱"`
	DisplayName   string  `json:"display_name" doc:"显示名称"`
	AvatarURL     *string `json:"avatar_url,omitempty" doc:"头像URL"`
	EmailVerified bool    `json:"email_verified" doc:"邮箱是否已验证"`
	Status        string  `json:"status" doc:"状态"`
	CreatedAt     string  `json:"created_at" doc:"创建时间"`
}

// NewProfileResource converts ent.User to ProfileResource
func NewProfileResource(u *ent.User) ProfileResource {
	return ProfileResource{
		ID:            uint(u.ID),
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		AvatarURL:     u.AvatarURL,
		EmailVerified: u.EmailVerified,
		Status:        string(u.Status),
		CreatedAt:     u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
