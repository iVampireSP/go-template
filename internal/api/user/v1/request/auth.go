package request

// LoginRequest 登录请求输入
type LoginRequest struct {
	Body struct {
		Email    string `json:"email" format:"email" doc:"邮箱"`
		Password string `json:"password" minLength:"1" doc:"密码"`
	}
}

// RegisterRequest 注册请求输入
type RegisterRequest struct {
	Body struct {
		Email           string `json:"email" format:"email" doc:"邮箱"`
		Password        string `json:"password" minLength:"8" doc:"密码"`
		PasswordConfirm string `json:"password_confirm" minLength:"1" doc:"确认密码"`
		DisplayName     string `json:"display_name,omitempty" doc:"显示名称"`
	}
}
