package v1

import (
	"github.com/gofiber/fiber/v2"
	"go-template/internal/api/http/response"
	"go-template/internal/schema"
	"go-template/internal/service/auth"
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
	user, ok := u.authService.GetUser(c)
	if !ok {
		return response.Ctx(c).Status(http.StatusUnauthorized).Send()
	}

	var currentUserResponse = &schema.CurrentUserResponse{
		IP:        c.IP(),
		Valid:     user.Valid,
		UserEmail: user.Token.Email,
		UserId:    user.Token.Sub,
		UserName:  user.Token.Name,
	}

	return response.Ctx(c).Data(currentUserResponse).Send()
}
