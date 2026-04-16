package request

// LoginRequest 登录请求输入
type LoginRequest struct {
	Body struct {
		Email    string `json:"email" format:"email" doc:"邮箱"`
		Password string `json:"password" minLength:"1" doc:"密码"`
	}
}

type MeRequest struct{}
