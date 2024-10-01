package auth

import (
	"context"
	"go-template/internal/schema"
)

type HasUserInterface interface {
	GetUserId() schema.UserId
}

func (a *Service) Compare(ctx context.Context, entity HasUserInterface) bool {
	return entity.GetUserId() == a.GetUserId(ctx)
}
