package dto

type TestRequest struct {
	Message string `json:"message" form:"message" validate:"string|required|min:1"`
}
