package request

import "github.com/iVampireSP/go-template/pkg/httpserver"

// ListUsersRequest 列出用户输入
type ListUsersRequest struct {
	httpserver.PaginationParams
	Search        string `query:"search" doc:"搜索邮箱或显示名"`
	Status        string `query:"status" doc:"筛选状态"`
	EmailVerified string `query:"email_verified" doc:"筛选邮箱验证状态：true/false"`
}

// GetUserRequest 获取用户输入
type GetUserRequest struct {
	ID int `path:"id" minimum:"1" doc:"用户ID"`
}
