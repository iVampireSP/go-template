package api

import (
	"github.com/google/wire"
	"go-template/internal/api/grpc"
)

var Provide = wire.NewSet(
	grpc.ProviderSet,
	NewApi,
)

type Api struct {
	GRPC *grpc.Handlers
}

func NewApi(
	grpcHandlers *grpc.Handlers,
) *Api {
	return &Api{
		GRPC: grpcHandlers,
	}
}
