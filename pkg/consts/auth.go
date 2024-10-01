package consts

import (
	"errors"
	"go-template/internal/schema"
)

const (
	AuthHeader = "Authorization"
	AuthPrefix = "Bearer"

	//AnonymousUser schema.UserId = 1
	AnonymousUser schema.UserId = "anonymous"

	AuthMiddlewareKey               = "auth.user"
	AuthAssistantShareMiddlewareKey = "auth.assistant.share"
)

var (
	ErrNotValidToken  = errors.New("无效的 JWT 令牌")
	ErrJWTFormatError = errors.New("JWT 格式错误")
	ErrNotBearerType  = errors.New("不是 Bearer 类型")
	ErrEmptyResponse  = errors.New("我们的服务器返回了空请求，可能某些环节出了问题")
	ErrTokenError     = errors.New("token 类型错误")
	ErrBearerToken    = errors.New("无效的 Bearer 令牌")

	ErrNotYourResource  = errors.New("你不能修改这个资源，因为它不是你创建的。")
	ErrPermissionDenied = errors.New("没有权限访问此资源")
)
