// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: pkg/protos/documentService/service.proto

package documentService

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DocumentServiceClient is the client API for DocumentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DocumentServiceClient interface {
	ListDocuments(ctx context.Context, in *ListDocumentsRequest, opts ...grpc.CallOption) (*ListDocumentsResponse, error)
}

type documentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDocumentServiceClient(cc grpc.ClientConnInterface) DocumentServiceClient {
	return &documentServiceClient{cc}
}

func (c *documentServiceClient) ListDocuments(ctx context.Context, in *ListDocumentsRequest, opts ...grpc.CallOption) (*ListDocumentsResponse, error) {
	out := new(ListDocumentsResponse)
	err := c.cc.Invoke(ctx, "/DocumentService/ListDocuments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DocumentServiceServer is the server API for DocumentService service.
// All implementations must embed UnimplementedDocumentServiceServer
// for forward compatibility
type DocumentServiceServer interface {
	ListDocuments(context.Context, *ListDocumentsRequest) (*ListDocumentsResponse, error)
	mustEmbedUnimplementedDocumentServiceServer()
}

// UnimplementedDocumentServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDocumentServiceServer struct {
}

func (UnimplementedDocumentServiceServer) ListDocuments(context.Context, *ListDocumentsRequest) (*ListDocumentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDocuments not implemented")
}
func (UnimplementedDocumentServiceServer) mustEmbedUnimplementedDocumentServiceServer() {}

// UnsafeDocumentServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DocumentServiceServer will
// result in compilation errors.
type UnsafeDocumentServiceServer interface {
	mustEmbedUnimplementedDocumentServiceServer()
}

func RegisterDocumentServiceServer(s grpc.ServiceRegistrar, srv DocumentServiceServer) {
	s.RegisterService(&DocumentService_ServiceDesc, srv)
}

func _DocumentService_ListDocuments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListDocumentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DocumentServiceServer).ListDocuments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DocumentService/ListDocuments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DocumentServiceServer).ListDocuments(ctx, req.(*ListDocumentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DocumentService_ServiceDesc is the grpc.ServiceDesc for DocumentService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DocumentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DocumentService",
	HandlerType: (*DocumentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListDocuments",
			Handler:    _DocumentService_ListDocuments_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/protos/documentService/service.proto",
}
