package handler

import "go-template/internal/handler/grpc"

type Handler struct {
	GRPC *grpc.Handlers
}

func NewHandler(
	grpcHandlers *grpc.Handlers,
) *Handler {
	return &Handler{
		GRPC: grpcHandlers,
	}
}
