package v1

import (
	"github.com/google/wire"
	"go-template/internal/handler/grpc/documents"
)

var ProviderGrpcHandlerSet = wire.NewSet(
	documents.NewDocumentService,
)

type GrpcHandlers struct {
	DocumentService *documents.DocumentService
}

func NewGrpcHandlers(
	documentService *documents.DocumentService,
) *GrpcHandlers {
	return &GrpcHandlers{
		DocumentService: documentService,
	}
}
