package documents

import (
	"context"
	"go-template/pkg/protos/documentService"
)

func (d *DocumentService) ListDocuments(ctx context.Context, req *documentService.ListDocumentsRequest) (*documentService.ListDocumentsResponse, error) {

	return &documentService.ListDocumentsResponse{}, nil
}
