// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: installation/v1alpha1/installation.proto

package installationv1alpha1

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

// InstallationsClient is the client API for Installations service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type InstallationsClient interface {
	ListInstallations(ctx context.Context, in *ListInstallationsRequest, opts ...grpc.CallOption) (*ListInstallationsResponse, error)
}

type installationsClient struct {
	cc grpc.ClientConnInterface
}

func NewInstallationsClient(cc grpc.ClientConnInterface) InstallationsClient {
	return &installationsClient{cc}
}

func (c *installationsClient) ListInstallations(ctx context.Context, in *ListInstallationsRequest, opts ...grpc.CallOption) (*ListInstallationsResponse, error) {
	out := new(ListInstallationsResponse)
	err := c.cc.Invoke(ctx, "/installation.v1alpha1.Installations/ListInstallations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InstallationsServer is the server API for Installations service.
// All implementations must embed UnimplementedInstallationsServer
// for forward compatibility
type InstallationsServer interface {
	ListInstallations(context.Context, *ListInstallationsRequest) (*ListInstallationsResponse, error)
	mustEmbedUnimplementedInstallationsServer()
}

// UnimplementedInstallationsServer must be embedded to have forward compatible implementations.
type UnimplementedInstallationsServer struct {
}

func (UnimplementedInstallationsServer) ListInstallations(context.Context, *ListInstallationsRequest) (*ListInstallationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListInstallations not implemented")
}
func (UnimplementedInstallationsServer) mustEmbedUnimplementedInstallationsServer() {}

// UnsafeInstallationsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to InstallationsServer will
// result in compilation errors.
type UnsafeInstallationsServer interface {
	mustEmbedUnimplementedInstallationsServer()
}

func RegisterInstallationsServer(s grpc.ServiceRegistrar, srv InstallationsServer) {
	s.RegisterService(&Installations_ServiceDesc, srv)
}

func _Installations_ListInstallations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListInstallationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InstallationsServer).ListInstallations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/installation.v1alpha1.Installations/ListInstallations",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InstallationsServer).ListInstallations(ctx, req.(*ListInstallationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Installations_ServiceDesc is the grpc.ServiceDesc for Installations service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Installations_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "installation.v1alpha1.Installations",
	HandlerType: (*InstallationsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListInstallations",
			Handler:    _Installations_ListInstallations_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "installation/v1alpha1/installation.proto",
}
