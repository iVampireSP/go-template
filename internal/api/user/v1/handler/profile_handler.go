package handler

import (
	"context"
	"errors"

	"github.com/iVampireSP/go-template/internal/api/user/v1/request"
	"github.com/iVampireSP/go-template/internal/api/user/v1/resource"
	"github.com/iVampireSP/go-template/internal/api/user/v1/response"
	"github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/iVampireSP/go-template/pkg/cerr"
)

// ProfileHandler 用户个人资料handler
type ProfileHandler struct {
	userService *user.User
}

func NewProfileHandler(userService *user.User) *ProfileHandler {
	return &ProfileHandler{userService: userService}
}

// Me 获取当前用户信息
func (c *ProfileHandler) Me(ctx context.Context, _ *struct{}) (*response.ProfileResponse, error) {
	u := user.MustContext(ctx)
	full, err := c.userService.GetByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	result := resource.NewProfileResource(full)
	return response.NewProfileResponse(result), nil
}

// UpdateMe 更新当前用户资料
func (c *ProfileHandler) UpdateMe(ctx context.Context, input *request.UpdateMeRequest) (*response.ProfileResponse, error) {
	u := user.MustContext(ctx)

	updated, err := c.userService.UpdateProfile(ctx, u.ID, user.UpdateProfileInput{
		DisplayName: input.Body.DisplayName,
		AvatarURL:   input.Body.AvatarURL,
	})
	if err != nil {
		return nil, err
	}

	result := resource.NewProfileResource(updated)
	return response.NewProfileResponse(result), nil
}

// UpdatePassword 修改当前用户密码
func (c *ProfileHandler) UpdatePassword(ctx context.Context, input *request.UpdatePasswordRequest) (*struct{}, error) {
	u := user.MustContext(ctx)

	if err := c.userService.VerifyPassword(ctx, u.ID, input.Body.CurrentPassword); err != nil {
		if errors.Is(err, user.ErrInvalidPassword) {
			return nil, cerr.BadRequest("current password is invalid").WithCode("INVALID_CURRENT_PASSWORD")
		}
		return nil, err
	}

	if err := c.userService.UpdatePassword(ctx, u.ID, input.Body.NewPassword); err != nil {
		return nil, err
	}

	return nil, nil
}
