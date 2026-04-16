package handler

import (
	"context"

	"github.com/iVampireSP/go-template/internal/api/admin/v1/request"
	"github.com/iVampireSP/go-template/internal/api/admin/v1/resource"
	"github.com/iVampireSP/go-template/internal/api/admin/v1/response"
	"github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/iVampireSP/go-template/pkg/paginator"
)

// UserHandler 用户管理handler
type UserHandler struct {
	userService *user.User
}

// NewUserHandler 创建用户管理handler
func NewUserHandler(userService *user.User) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// List 列出用户
func (c *UserHandler) List(ctx context.Context, input *request.ListUsersRequest) (*response.UserListResponse, error) {
	users, total, err := c.userService.List(ctx, user.UserListInput{
		Search:        input.Search,
		Status:        input.Status,
		EmailVerified: input.EmailVerified,
		Page:          input.Page,
		PerPage:       input.PerPage,
	})
	if err != nil {
		return nil, err
	}

	userResources := make([]resource.UserResource, 0, len(users))
	for _, u := range users {
		userResources = append(userResources, resource.NewUserResource(u))
	}

	page := paginator.New(userResources, total, input.PerPage, input.Page)
	return response.NewUserListResponse(page), nil
}

// Get 获取用户详情
func (c *UserHandler) Get(ctx context.Context, input *request.GetUserRequest) (*response.UserResponse, error) {
	u, err := c.userService.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return response.NewUserResponse(resource.NewUserResource(u)), nil
}
