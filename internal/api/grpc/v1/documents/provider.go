package documents

import (
	"go-template/internal/dao"
	v1 "go-template/proto/gen/proto/api/v1"
)

type Handler struct {
	v1.UnimplementedDocumentServiceServer
	dao *dao.Query
}

func NewHandler(dao *dao.Query) *Handler {
	return &Handler{
		v1.UnimplementedDocumentServiceServer{},
		dao,
	}
}
