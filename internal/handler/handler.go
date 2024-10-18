package handler

import (
	"github.com/google/wire"
	"go-template/internal/handler/grpc"
	"go-template/internal/handler/http"
)

var ProviderSet = wire.NewSet(
	grpc.ProviderSet,
	http.ProviderSet,
	NewHandler,
)

type Handler struct {
	GRPC *grpc.Handlers
	HTTP *http.Handlers
}

func NewHandler(
	grpcHandlers *grpc.Handlers,
	httpHandlers *http.Handlers,
) *Handler {
	return &Handler{
		GRPC: grpcHandlers,
		HTTP: httpHandlers,
	}
}
