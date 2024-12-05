package dto

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

type IError interface {
	Error() string
}

type ValidateError struct {
	Message string `json:"message"`
}

// Body 为 HTTP 响应
type Body struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	//ValidationErrors *[]ValidateError `json:"validation_error,omitempty"`
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
	Wrap    bool `json:"-"`
}

type HttpResponse struct {
	body       *Body
	httpStatus int
	ctx        *fiber.Ctx
}

func Ctx(c *fiber.Ctx) *HttpResponse {
	return &HttpResponse{
		body: &Body{
			Wrap: true,
		},
		ctx:        c,
		httpStatus: 0,
	}
}

func (r *HttpResponse) Message(message string) *HttpResponse {
	r.body.Message = message

	if r.httpStatus == 0 {
		r.httpStatus = http.StatusOK
	}

	return r
}

// WithoutWrap 将不在 body 中包裹 data
func (r *HttpResponse) WithoutWrap() *HttpResponse {
	r.body.Wrap = false

	return r
}

func (r *HttpResponse) Wrap() *HttpResponse {
	r.body.Wrap = true
	return r
}

func (r *HttpResponse) Data(data any) *HttpResponse {
	r.body.Data = data
	return r

}

func (r *HttpResponse) Error(err IError) *HttpResponse {
	if err != nil {
		r.body.Error = err.Error()

		if r.httpStatus == 0 {
			r.httpStatus = http.StatusBadRequest
		}

		if r.body.Message == "" {
			r.Message("Something went wrong")
		}

		r.body.Success = false
	}

	return r

}

func (r *HttpResponse) Status(status int) *HttpResponse {
	r.httpStatus = status
	return r

}

func (r *HttpResponse) Send() error {
	if r.httpStatus == 0 {
		r.httpStatus = http.StatusOK
	}

	// if 20x or 20x, set success
	r.body.Success = r.httpStatus >= http.StatusOK && r.httpStatus < http.StatusMultipleChoices

	// if 403 or 401 but not have message
	if r.httpStatus == http.StatusForbidden || r.httpStatus == http.StatusUnauthorized {
		if r.body.Message == "" {
			r.Message("Unauthorized")
		}
	}

	if r.body.Wrap {
		return r.ctx.Status(r.httpStatus).JSON(r.body)
	}

	return r.ctx.Status(r.httpStatus).JSON(r.body.Data)
}

//func (r *HttpResponse) ValidationError(validationErrors *[]ValidateError) *HttpResponse {
//	if validationErrors == nil || len(*validationErrors) == 0 {
//	}
//
//	r.body.ValidationErrors = validationErrors
//
//	r.Error(errs.ErrBadRequest)
//
//	return r
//}

//
//func ResponseMessage(c *gin.Context, code int, message string, data interface{}) {
//	c.JSON(code, &Body{
//		Message: message,
//		Data:    data,
//	})
//	c.Abort()
//}
//
//func ResponseError(c *gin.Context, code int, err error) {
//	c.JSON(code, &Body{
//		Error: err.Error(),
//	})
//	c.Abort()
//}
//
//func Response(c *gin.Context, code int, data interface{}) {
//	c.JSON(code, &Body{
//		Data: data,
//	})
//	c.Abort()
//}
