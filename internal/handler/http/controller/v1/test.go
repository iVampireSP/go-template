package v1

import (
	"go-template/internal/schema"
	"go-template/internal/service/auth"
	"net/http"

	"github.com/gin-gonic/gin"
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
// @Success      200  {object}  schema.ResponseBody{data=schema.CurrentUserResponse}
// @Failure      400  {object}  schema.ResponseBody
// @Router       /api/v1/ping [get]
func (u *UserController) Test(c *gin.Context) {
	user := u.authService.GetUser(c)

	var currentUserResponse = &schema.CurrentUserResponse{
		IP:        c.ClientIP(),
		Valid:     user.Valid,
		UserEmail: user.Token.Email,
		UserId:    user.Token.Sub,
		UserName:  user.Token.Name,
	}

	schema.NewResponse(c).Status(http.StatusOK).Data(currentUserResponse).Send()
}
