package grpc

import (
	"github.com/google/wire"
	"go-template/internal/handler/grpc/documents"
	"go-template/internal/handler/grpc/interceptor"
)

var ProviderSet = wire.NewSet(
	interceptor.NewAuth,
	interceptor.NewLogger,
	documents.NewApi,

	NewInterceptor,
	NewHandler,
)

func NewHandler(
	documentApi *documents.Api,
	interceptor2 *Interceptor,
) *Handlers {
	return &Handlers{
		DocumentApi: documentApi,
		Interceptor: interceptor2,
	}
}

type Handlers struct {
	DocumentApi *documents.Api
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
