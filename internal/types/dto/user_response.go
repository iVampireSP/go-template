package dto

import (
	"go-template/internal/types/user"
)

type CurrentUserResponse struct {
	IP        string  `json:"ip"`
	Valid     bool    `json:"valid"`
	UserEmail string  `json:"userEmail"`
	UserId    user.Id `json:"userId"`
	UserName  string  `json:"userName"`
}
