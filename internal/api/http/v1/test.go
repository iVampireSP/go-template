package v1

import (
	"github.com/gofiber/fiber/v2"
	"go-template/internal/api/http/response"
	"go-template/internal/services/auth"
	"go-template/internal/types/dto"
	"net/http"
)

type UserController struct {
	authService *auth.Service
}

func NewUserController(authService *auth.Service) *UserController {
	return &UserController{authService}
}

// Test godoc
// @Summary      Greet
// @Description  测试接口，将会返回当前用户的信息
// @Tags         ping
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @deprecated   true
// @Success      200  {object}  response.Body{data=schema.CurrentUserResponse}
// @Failure      400  {object}  response.Body
// @Router       /api/v1/ping [get]
func (u *UserController) Test(c *fiber.Ctx) error {
	user := u.authService.GetUser(c)

	var currentUserResponse = &dto.CurrentUserResponse{
		IP:        c.IP(),
		Valid:     user.Valid,
		UserEmail: user.Token.Email,
		UserId:    user.Token.Sub,
		UserName:  user.Token.Name,
	}

	return response.Ctx(c).Status(http.StatusOK).Data(currentUserResponse).Send()
}
