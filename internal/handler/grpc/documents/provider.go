package documents

import (
	"go-template/internal/dao"
	"go-template/pkg/protos/documentService"
)

type DocumentService struct {
	documentService.UnimplementedDocumentServiceServer
	dao *dao.Query
}

func NewDocumentService(dao *dao.Query) *DocumentService {
	return &DocumentService{
		documentService.UnimplementedDocumentServiceServer{},
		dao,
	}
}
