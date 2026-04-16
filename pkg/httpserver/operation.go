package httpserver

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// Security 是 OpenAPI 安全要求。
type Security = []map[string][]string

// API 封装 huma.API，对路由层隐藏底层框架。
type API struct {
	api huma.API
}

// WrapAPI 从 huma.API 创建 API。
func WrapAPI(api huma.API) API {
	return API{api: api}
}

// Group 创建新的 API 分组（对应 huma.NewGroup）。
func (a API) Group() API {
	return API{api: huma.NewGroup(a.api)}
}

// UseMiddleware 添加 Huma 中间件到此分组。
func (a API) UseMiddleware(middlewares ...func(ctx huma.Context, next func(huma.Context))) {
	a.api.(interface {
		UseMiddleware(...func(ctx huma.Context, next func(huma.Context)))
	}).UseMiddleware(middlewares...)
}

// Huma 返回底层 huma.API，用于 middleware 等需要直接访问 huma 的场景。
func (a API) Huma() huma.API {
	return a.api
}

// Operation 路由操作描述。Method 和 Path 由 GET/POST/... 函数提供，不在此结构中。
type Operation struct {
	ID            string   // OperationID
	Summary       string   // 短描述
	Description   string   // 长描述（可选）
	Tags          []string // OpenAPI 分组标签
	Security      Security // 安全要求（可选）
	DefaultStatus int      // 默认成功状态码（0 = 由 Huma 自动决定）
}

// toHumaOperation 将 Operation 转为 huma.Operation。
func (o Operation) toHumaOperation(method, path string) huma.Operation {
	return huma.Operation{
		OperationID:   o.ID,
		Method:        method,
		Path:          path,
		Summary:       o.Summary,
		Description:   o.Description,
		Tags:          o.Tags,
		Security:      o.Security,
		DefaultStatus: o.DefaultStatus,
	}
}

// methodGET 等常量避免每次调用 net/http。
const (
	methodGET     = http.MethodGet
	methodPOST    = http.MethodPost
	methodPUT     = http.MethodPut
	methodPATCH   = http.MethodPatch
	methodDELETE  = http.MethodDelete
	methodOPTIONS = http.MethodOptions
	methodHEAD    = http.MethodHead
)
