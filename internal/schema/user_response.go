package schema

type CurrentUserResponse struct {
	IP        string `json:"ip"`
	Valid     bool   `json:"valid"`
	UserEmail string `json:"userEmail"`
	UserId    UserId `json:"userId"`
	UserName  string `json:"userName"`
}
