package documents

import (
	"go-template/internal/dao"
	v1 "go-template/proto/gen/proto/api/v1"
)

type Api struct {
	v1.UnimplementedDocumentServiceServer
	dao *dao.Query
}

func NewApi(dao *dao.Query) *Api {
	return &Api{
		v1.UnimplementedDocumentServiceServer{},
		dao,
	}
}
