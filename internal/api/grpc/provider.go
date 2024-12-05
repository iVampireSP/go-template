package grpc

import (
	"github.com/google/wire"
	"go-template/internal/api/grpc/interceptor"
	"go-template/internal/api/grpc/v1/documents"
)

var ProviderSet = wire.NewSet(
	interceptor.NewAuth,
	interceptor.NewLogger,
	documents.NewHandler,

	NewInterceptor,
	NewHandler,
)

func NewHandler(
	documentApi *documents.Handler,
	interceptor2 *Interceptor,
) *Handlers {
	return &Handlers{
		DocumentApi: documentApi,
		Interceptor: interceptor2,
	}
}

type Handlers struct {
	DocumentApi *documents.Handler
	Interceptor *Interceptor
}

type Interceptor struct {
	Auth   *interceptor.Auth
	Logger *interceptor.Logger
}

func NewInterceptor(
	Auth *interceptor.Auth,
	Logger *interceptor.Logger,
) *Interceptor {
	return &Interceptor{
		Auth:   Auth,
		Logger: Logger,
	}
}
