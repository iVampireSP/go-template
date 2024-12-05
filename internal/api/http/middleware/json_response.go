package middleware

import (
	"github.com/gofiber/fiber/v2"
)

type JSONResponse struct {
}

func (*JSONResponse) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Response().Header.Set("Content-Type", "application/json")
		return c.Next()

	}
}

func NewJSONResponse() *JSONResponse {
	return &JSONResponse{}
}
