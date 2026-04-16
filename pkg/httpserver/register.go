package httpserver

import (
	"context"
	"reflect"

	"github.com/danielgtaylor/huma/v2"

	"github.com/iVampireSP/go-template/pkg/validator"
)

// GET 注册 GET 路由。
func GET[I, O any](api API, path string, op Operation, handler func(context.Context, *I) (*O, error)) {
	register(api, methodGET, path, op, handler)
}

// POST 注册 POST 路由。
func POST[I, O any](api API, path string, op Operation, handler func(context.Context, *I) (*O, error)) {
	register(api, methodPOST, path, op, handler)
}

// PUT 注册 PUT 路由。
func PUT[I, O any](api API, path string, op Operation, handler func(context.Context, *I) (*O, error)) {
	register(api, methodPUT, path, op, handler)
}

// PATCH 注册 PATCH 路由。
func PATCH[I, O any](api API, path string, op Operation, handler func(context.Context, *I) (*O, error)) {
	register(api, methodPATCH, path, op, handler)
}

// DELETE 注册 DELETE 路由。
func DELETE[I, O any](api API, path string, op Operation, handler func(context.Context, *I) (*O, error)) {
	register(api, methodDELETE, path, op, handler)
}

// OPTIONS 注册 OPTIONS 路由。
func OPTIONS[I, O any](api API, path string, op Operation, handler func(context.Context, *I) (*O, error)) {
	register(api, methodOPTIONS, path, op, handler)
}

// HEAD 注册 HEAD 路由。
func HEAD[I, O any](api API, path string, op Operation, handler func(context.Context, *I) (*O, error)) {
	register(api, methodHEAD, path, op, handler)
}

// RegisterRaw 直接使用 huma.Operation 注册路由（escape hatch）。
// 用于需要 Responses 等高级 huma.Operation 字段的特殊场景。
func RegisterRaw[I, O any](api API, op huma.Operation, handler func(context.Context, *I) (*O, error)) {
	handler = wrapValidation(handler)
	huma.Register(api.api, op, handler)
}

// register 是所有 HTTP 方法函数的内部实现。
func register[I, O any](api API, method, path string, op Operation, handler func(context.Context, *I) (*O, error)) {
	handler = wrapValidation(handler)
	huma.Register(api.api, op.toHumaOperation(method, path), handler)
}

// wrapValidation 如果 Request 类型有已注册的验证 tag，则包装 handler 在调用前运行验证。
// 无已注册 tag 时返回原 handler（零开销）。
func wrapValidation[I, O any](handler func(context.Context, *I) (*O, error)) func(context.Context, *I) (*O, error) {
	inputType := reflect.TypeOf((*I)(nil)).Elem()
	if !validator.HasRegisteredTags(inputType) {
		return handler
	}

	original := handler
	return func(ctx context.Context, input *I) (*O, error) {
		if errs := validator.ValidateFields(ctx, input); len(errs) > 0 {
			return nil, huma.Error422UnprocessableEntity("validation failed", errs...)
		}
		return original(ctx, input)
	}
}
