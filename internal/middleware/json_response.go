package middleware

import (
	"github.com/gin-gonic/gin"
)

type JSONResponseMiddleware struct {
}

func (*JSONResponseMiddleware) ContentTypeJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Next()
}

func NewJSONResponseMiddleware() *JSONResponseMiddleware {
	return &JSONResponseMiddleware{}
}
