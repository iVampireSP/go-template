package documents

import (
	"go-template/ent"
	v1 "go-template/proto/gen/proto/api/v1"
)

type Handler struct {
	v1.UnimplementedDocumentServiceServer
	ent *ent.Client
}

func NewHandler(ent *ent.Client) *Handler {
	return &Handler{
		v1.UnimplementedDocumentServiceServer{},
		ent,
	}
}
