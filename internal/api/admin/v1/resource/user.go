package resource

import (
	"time"

	"github.com/iVampireSP/go-template/ent"
)

// UserResource 用户响应
type UserResource struct {
	ID               int        `json:"id" doc:"用户ID"`
	Email            string     `json:"email" doc:"邮箱"`
	DisplayName      string     `json:"display_name" doc:"显示名称"`
	AvatarURL        *string    `json:"avatar_url,omitempty" doc:"头像URL"`
	EmailVerified    bool       `json:"email_verified" doc:"邮箱是否验证"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty" doc:"邮箱验证时间"`
	Status           string     `json:"status" doc:"状态"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty" doc:"最后登录时间"`
	LastLoginIP      *string    `json:"last_login_ip,omitempty" doc:"最后登录IP"`
	RegisterIP       *string    `json:"register_ip,omitempty" doc:"注册IP"`
	LoginCount       int        `json:"login_count" doc:"登录次数"`
	CreatedAt        time.Time  `json:"created_at" doc:"创建时间"`
	UpdatedAt        time.Time  `json:"updated_at" doc:"更新时间"`
}

// NewUserResource converts ent.User to UserResource
func NewUserResource(u *ent.User) UserResource {
	return UserResource{
		ID:              u.ID,
		Email:           u.Email,
		DisplayName:     u.DisplayName,
		AvatarURL:       u.AvatarURL,
		EmailVerified:   u.EmailVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		Status:          string(u.Status),
		LastLoginAt:     u.LastLoginAt,
		LastLoginIP:     u.LastLoginIP,
		RegisterIP:      u.RegisterIP,
		LoginCount:      u.LoginCount,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}
