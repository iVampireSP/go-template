package grpc

import (
	"github.com/google/wire"
	"leafdev.top/Leaf/leaf-library/internal/handler/grpc/documents"
	"leafdev.top/Leaf/leaf-library/internal/handler/grpc/interceptor"
)

var ProviderSet = wire.NewSet(
	interceptor.NewAuth,
	interceptor.NewLogger,
	documents.NewDocumentService,

	NewInterceptor,
	NewHandler,
)

func NewHandler(
	documentService *documents.DocumentService,
	interceptor2 *Interceptor,
) *Handlers {
	return &Handlers{
		DocumentService: documentService,
		Interceptor:     interceptor2,
	}
}

type Handlers struct {
	DocumentService *documents.DocumentService
	Interceptor     *Interceptor
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
