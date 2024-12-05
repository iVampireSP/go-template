package middleware

import (
	"github.com/gofiber/fiber/v2"
)

type JSONResponseMiddleware struct {
}

func (*JSONResponseMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Response().Header.Set("Content-Type", "application/json")
		return c.Next()

	}
}

func NewJSONResponseMiddleware() *JSONResponseMiddleware {
	return &JSONResponseMiddleware{}
}
