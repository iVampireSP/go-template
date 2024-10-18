package grpc

import (
	"github.com/google/wire"
	"go-template/internal/handler/grpc/documents"
	"go-template/internal/handler/grpc/interceptor"
)

type Interceptor struct {
	Auth   *interceptor.Auth
	Logger *interceptor.Logger
}

func NewHandler(
	documentService *documents.DocumentService,
) *Handlers {
	return &Handlers{
		DocumentService: documentService,
	}
}

var ProviderSet = wire.NewSet(
	documents.NewDocumentService,
	interceptor.NewAuth,
	interceptor.NewLogger,
	NewHandler,
)

type Handlers struct {
	DocumentService *documents.DocumentService
	Interceptor     *Interceptor
}
