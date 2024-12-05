package dto

import (
	"go-template/internal/types/auth"
)

type CurrentUserResponse struct {
	IP        string  `json:"ip"`
	Valid     bool    `json:"valid"`
	UserEmail string  `json:"userEmail"`
	UserId    auth.Id `json:"userId"`
	UserName  string  `json:"userName"`
}
