package request

import "github.com/iVampireSP/go-template/pkg/httpserver"

// ListAdminsRequest 列出管理员输入
type ListAdminsRequest struct {
	httpserver.PaginationParams
	Search string `query:"search" doc:"搜索邮箱或显示名"`
	Status string `query:"status" doc:"筛选状态 (active, suspended)"`
}

// GetAdminRequest 获取管理员输入
type GetAdminRequest struct {
	ID int `path:"id" minimum:"1" doc:"管理员ID"`
}

// CreateAdminRequest 创建管理员输入
type CreateAdminRequest struct {
	Body struct {
		Email       string `json:"email" format:"email" doc:"邮箱"`
		Password    string `json:"password" minLength:"8" doc:"密码"`
		DisplayName string `json:"display_name" minLength:"1" maxLength:"200" doc:"显示名称"`
	}
}

// UpdateAdminRequest 更新管理员输入
type UpdateAdminRequest struct {
	ID   int `path:"id" minimum:"1" doc:"管理员ID"`
	Body struct {
		Email       *string `json:"email,omitempty" format:"email" doc:"邮箱"`
		DisplayName *string `json:"display_name,omitempty" minLength:"1" maxLength:"200" doc:"显示名称"`
	}
}

// DeleteAdminRequest 删除管理员输入
type DeleteAdminRequest struct {
	ID int `path:"id" minimum:"1" doc:"管理员ID"`
}

// RestoreAdminRequest 恢复管理员输入
type RestoreAdminRequest struct {
	ID int `path:"id" minimum:"1" doc:"管理员ID"`
}

// UpdateAdminPasswordRequest 更新管理员密码输入
type UpdateAdminPasswordRequest struct {
	ID   int `path:"id" minimum:"1" doc:"管理员ID"`
	Body struct {
		Password string `json:"password" minLength:"8" doc:"新密码"`
	}
}

// UpdateAdminStatusRequest 更新管理员状态输入
type UpdateAdminStatusRequest struct {
	ID   int `path:"id" minimum:"1" doc:"管理员ID"`
	Body struct {
		Status string `json:"status" enum:"active,suspended" doc:"状态"`
	}
}

// ListAdminTicketDepartmentsRequest 列出管理员工单部门输入
type ListAdminTicketDepartmentsRequest struct {
	ID int `path:"id" minimum:"1" doc:"管理员ID"`
}

// UpdateAdminTicketDepartmentsRequest 更新管理员工单部门输入
type UpdateAdminTicketDepartmentsRequest struct {
	ID   int `path:"id" minimum:"1" doc:"管理员ID"`
	Body struct {
		DepartmentIDs []int `json:"department_ids" doc:"工单部门ID列表"`
	}
}
