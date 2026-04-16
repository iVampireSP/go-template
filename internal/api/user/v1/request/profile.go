package request

// UpdateMeRequest 更新当前用户资料请求
type UpdateMeRequest struct {
	Body struct {
		DisplayName *string `json:"display_name,omitempty" minLength:"1" maxLength:"200" doc:"显示名称"`
		AvatarURL   *string `json:"avatar_url,omitempty" maxLength:"500" doc:"头像 URL；传空字符串清空"`
	}
}

// UpdatePasswordRequest 修改密码请求
type UpdatePasswordRequest struct {
	Body struct {
		CurrentPassword string `json:"current_password" minLength:"1" doc:"当前密码"`
		NewPassword     string `json:"new_password" minLength:"8" doc:"新密码"`
	}
}
