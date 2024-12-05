package api

import (
	"github.com/google/wire"
	"go-template/internal/api/grpc"
	"go-template/internal/api/http"
)

var Provide = wire.NewSet(
	grpc.ProviderSet,
	http.ProviderSet,
	NewApi,
)

type Api struct {
	GRPC *grpc.Handlers
	HTTP *http.Handlers
}

func NewApi(
	grpcHandlers *grpc.Handlers,
	httpHandlers *http.Handlers,
) *Api {
	return &Api{
		GRPC: grpcHandlers,
		HTTP: httpHandlers,
	}
}
