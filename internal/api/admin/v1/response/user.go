package response

import (
	"github.com/iVampireSP/go-template/internal/api/admin/v1/resource"
	"github.com/iVampireSP/go-template/pkg/paginator"
)

// UserResponse is the Huma output for a single UserResource.
type UserResponse struct {
	Body resource.UserResource
}

func NewUserResponse(userResource resource.UserResource) *UserResponse {
	return &UserResponse{
		Body: userResource,
	}
}

// UserListResponse is the Huma output for a list of UserResource.
type UserListResponse struct {
	Body paginator.LengthAwarePaginator[resource.UserResource]
}

func NewUserListResponse(p *paginator.LengthAwarePaginator[resource.UserResource]) *UserListResponse {
	return &UserListResponse{
		Body: *p,
	}
}
