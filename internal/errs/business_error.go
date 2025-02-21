package errs

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// RecoverableError 定义可以被恢复的错误的基础接口
type RecoverableError interface {
	error
	IsRecoverable() bool
}

// HTTPError HTTP 错误，可以直接映射到 HTTP 响应
type HTTPError struct {
	Code       int               `json:"code"`    // HTTP 状态码
	Message    string            `json:"message"` // 错误消息
	Details    interface{}       `json:"details"` // 详细错误信息
	Headers    map[string]string `json:"-"`       // 自定义 HTTP 头
	StatusCode int               `json:"-"`       // HTTP 状态码
}

func (e *HTTPError) Error() string {
	return e.Message
}

func (e *HTTPError) IsRecoverable() bool {
	return true
}

// NewHTTPError 创建一个新的 HTTP 错误
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Code:       statusCode,
		Message:    message,
	}
}

// WithDetails 添加详细错误信息
func (e *HTTPError) WithDetails(details interface{}) *HTTPError {
	e.Details = details
	return e
}

// WithHeaders 添加自定义 HTTP 头
func (e *HTTPError) WithHeaders(headers map[string]string) *HTTPError {
	e.Headers = headers
	return e
}

// ValidateError 数据验证错误
type ValidateError struct {
	Field   string
	Message string
}

type ValidateErrors []ValidateError

func (e ValidateErrors) Error() string {
	var messages []string
	for _, v := range e {
		messages = append(messages, fmt.Sprintf("%s: %s", v.Field, v.Message))
	}
	return strings.Join(messages, "; ")
}

func (e ValidateErrors) IsRecoverable() bool {
	return true
}

// NewBadRequestError HTTP 错误快捷方法
func NewBadRequestError(message string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, message)
}

func NewUnauthorizedError(message string) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, message)
}

func NewForbiddenError(message string) *HTTPError {
	return NewHTTPError(http.StatusForbidden, message)
}

func NewNotFoundError(message string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, message)
}

func NewInternalServerError(message string) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, message)
}

// WrapError 包装错误，将错误转换为可恢复或不可恢复的错误
func WrapError(err error) error {
	if err == nil {
		return nil
	}

	// 如果已经是可恢复错误，直接返回
	var recoverableError RecoverableError
	if errors.As(err, &recoverableError) {
		return err
	}

	// 检查是否是预定义的错误类型
	var wrappedErr error
	switch {
	case errors.Is(err, ErrInternalServerError):
		wrappedErr = NewInternalServerError("内部服务器错误")
	case errors.Is(err, ErrBadRequest):
		wrappedErr = NewBadRequestError("无效的请求")
	case errors.Is(err, ErrUnauthorized):
		wrappedErr = NewUnauthorizedError("未授权的访问")
	case errors.Is(err, ErrNotFound):
		wrappedErr = NewNotFoundError("资源未找到")
	default:
		return err
	}

	return errors.Join(wrappedErr, err)
}
