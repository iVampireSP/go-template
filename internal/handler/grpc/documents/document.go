package documents

import (
	"context"
	v1 "go-template/proto/gen/proto/api/v1"
)

func (d *Api) ListDocuments(ctx context.Context, req *v1.ListDocumentsRequest) (*v1.ListDocumentsResponse, error) {
	return &v1.ListDocumentsResponse{}, nil
}
