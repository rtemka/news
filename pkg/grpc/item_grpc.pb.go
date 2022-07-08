// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.2
// source: pkg/grpc/item.proto

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// NewsClient is the client API for News service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NewsClient interface {
	List(ctx context.Context, in *wrapperspb.Int64Value, opts ...grpc.CallOption) (*Items, error)
}

type newsClient struct {
	cc grpc.ClientConnInterface
}

func NewNewsClient(cc grpc.ClientConnInterface) NewsClient {
	return &newsClient{cc}
}

func (c *newsClient) List(ctx context.Context, in *wrapperspb.Int64Value, opts ...grpc.CallOption) (*Items, error) {
	out := new(Items)
	err := c.cc.Invoke(ctx, "/newsgrpc.News/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NewsServer is the server API for News service.
// All implementations must embed UnimplementedNewsServer
// for forward compatibility
type NewsServer interface {
	List(context.Context, *wrapperspb.Int64Value) (*Items, error)
	mustEmbedUnimplementedNewsServer()
}

// UnimplementedNewsServer must be embedded to have forward compatible implementations.
type UnimplementedNewsServer struct {
}

func (UnimplementedNewsServer) List(context.Context, *wrapperspb.Int64Value) (*Items, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedNewsServer) mustEmbedUnimplementedNewsServer() {}

// UnsafeNewsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NewsServer will
// result in compilation errors.
type UnsafeNewsServer interface {
	mustEmbedUnimplementedNewsServer()
}

func RegisterNewsServer(s grpc.ServiceRegistrar, srv NewsServer) {
	s.RegisterService(&News_ServiceDesc, srv)
}

func _News_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(wrapperspb.Int64Value)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NewsServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/newsgrpc.News/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NewsServer).List(ctx, req.(*wrapperspb.Int64Value))
	}
	return interceptor(ctx, in, info, handler)
}

// News_ServiceDesc is the grpc.ServiceDesc for News service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var News_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "newsgrpc.News",
	HandlerType: (*NewsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _News_List_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/item.proto",
}
