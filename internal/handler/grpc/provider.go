package grpc

import (
	"github.com/google/wire"
	"go-template/internal/handler/grpc/documents"
)

var ProviderGrpcHandlerSet = wire.NewSet(
	documents.NewDocumentService,
)

type Handlers struct {
	DocumentService *documents.DocumentService
}

func NewGrpcHandlers(
	documentService *documents.DocumentService,
) *Handlers {
	return &Handlers{
		DocumentService: documentService,
	}
}
