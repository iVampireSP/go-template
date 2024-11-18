package middleware

import (
	"github.com/labstack/echo/v4"
)

type JSONResponseMiddleware struct {
}

func (*JSONResponseMiddleware) Handler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Content-Type", "application/json")

			return next(c)
		}
	}
}

func NewJSONResponseMiddleware() *JSONResponseMiddleware {
	return &JSONResponseMiddleware{}
}
