package response

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Body struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Wrap    bool   `json:"-"`
}

type HttpResponse struct {
	body       *Body
	httpStatus int
	ctx        echo.Context
}

func Ctx(c echo.Context) *HttpResponse {
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

func (r *HttpResponse) Error(err error) *HttpResponse {
	if err != nil {
		var errMsg = err.Error()

		if errMsg == "EOF" {
			errMsg = "Request body is empty or missing some fields, make sure you have provided all the required fields"
		}

		r.body.Error = errMsg

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

	if r.body.Wrap {
		return r.ctx.JSON(r.httpStatus, r.body)
	}

	return r.ctx.JSON(r.httpStatus, r.body.Data)
}

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
