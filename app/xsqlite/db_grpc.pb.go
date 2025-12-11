package xsqlite

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	DbService_GetHandler_FullMethodName         = "/x.db.DbService/GetHandler"
	DbService_GetAllHandlers_FullMethodName     = "/x.db.DbService/GetAllHandlers"
	DbService_GetHandlersByGroup_FullMethodName = "/x.db.DbService/GetHandlersByGroup"
	DbService_GetBatchedHandlers_FullMethodName = "/x.db.DbService/GetBatchedHandlers"
	DbService_UpdateHandler_FullMethodName      = "/x.db.DbService/UpdateHandler"
)

// DbServiceClient is the client API for DbService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DbServiceClient interface {
	GetHandler(ctx context.Context, in *GetHandlerRequest, opts ...grpc.CallOption) (*DbOutboundHandler, error)
	GetAllHandlers(ctx context.Context, in *GetAllHandlersRequest, opts ...grpc.CallOption) (*DbHandlers, error)
	GetHandlersByGroup(ctx context.Context, in *GetHandlersByGroupRequest, opts ...grpc.CallOption) (*DbHandlers, error)
	GetBatchedHandlers(ctx context.Context, in *GetBatchedHandlersRequest, opts ...grpc.CallOption) (*DbHandlers, error)
	UpdateHandler(ctx context.Context, in *UpdateHandlerRequest, opts ...grpc.CallOption) (*Receipt, error)
}

type dbServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDbServiceClient(cc grpc.ClientConnInterface) DbServiceClient {
	return &dbServiceClient{cc}
}

func (c *dbServiceClient) GetHandler(ctx context.Context, in *GetHandlerRequest, opts ...grpc.CallOption) (*DbOutboundHandler, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DbOutboundHandler)
	err := c.cc.Invoke(ctx, DbService_GetHandler_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dbServiceClient) GetAllHandlers(ctx context.Context, in *GetAllHandlersRequest, opts ...grpc.CallOption) (*DbHandlers, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DbHandlers)
	err := c.cc.Invoke(ctx, DbService_GetAllHandlers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dbServiceClient) GetHandlersByGroup(ctx context.Context, in *GetHandlersByGroupRequest, opts ...grpc.CallOption) (*DbHandlers, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DbHandlers)
	err := c.cc.Invoke(ctx, DbService_GetHandlersByGroup_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dbServiceClient) GetBatchedHandlers(ctx context.Context, in *GetBatchedHandlersRequest, opts ...grpc.CallOption) (*DbHandlers, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DbHandlers)
	err := c.cc.Invoke(ctx, DbService_GetBatchedHandlers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dbServiceClient) UpdateHandler(ctx context.Context, in *UpdateHandlerRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, DbService_UpdateHandler_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DbServiceServer is the server API for DbService service.
// All implementations must embed UnimplementedDbServiceServer
// for forward compatibility.
type DbServiceServer interface {
	GetHandler(context.Context, *GetHandlerRequest) (*DbOutboundHandler, error)
	GetAllHandlers(context.Context, *GetAllHandlersRequest) (*DbHandlers, error)
	GetHandlersByGroup(context.Context, *GetHandlersByGroupRequest) (*DbHandlers, error)
	GetBatchedHandlers(context.Context, *GetBatchedHandlersRequest) (*DbHandlers, error)
	UpdateHandler(context.Context, *UpdateHandlerRequest) (*Receipt, error)
	mustEmbedUnimplementedDbServiceServer()
}

// UnimplementedDbServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedDbServiceServer struct{}

func (UnimplementedDbServiceServer) GetHandler(context.Context, *GetHandlerRequest) (*DbOutboundHandler, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHandler not implemented")
}
func (UnimplementedDbServiceServer) GetAllHandlers(context.Context, *GetAllHandlersRequest) (*DbHandlers, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllHandlers not implemented")
}
func (UnimplementedDbServiceServer) GetHandlersByGroup(context.Context, *GetHandlersByGroupRequest) (*DbHandlers, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHandlersByGroup not implemented")
}
func (UnimplementedDbServiceServer) GetBatchedHandlers(context.Context, *GetBatchedHandlersRequest) (*DbHandlers, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBatchedHandlers not implemented")
}
func (UnimplementedDbServiceServer) UpdateHandler(context.Context, *UpdateHandlerRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateHandler not implemented")
}
func (UnimplementedDbServiceServer) mustEmbedUnimplementedDbServiceServer() {}
func (UnimplementedDbServiceServer) testEmbeddedByValue()                   {}

// UnsafeDbServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DbServiceServer will
// result in compilation errors.
type UnsafeDbServiceServer interface {
	mustEmbedUnimplementedDbServiceServer()
}

func RegisterDbServiceServer(s grpc.ServiceRegistrar, srv DbServiceServer) {
	// If the following call pancis, it indicates UnimplementedDbServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&DbService_ServiceDesc, srv)
}

func _DbService_GetHandler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetHandlerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DbServiceServer).GetHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DbService_GetHandler_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DbServiceServer).GetHandler(ctx, req.(*GetHandlerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DbService_GetAllHandlers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllHandlersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DbServiceServer).GetAllHandlers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DbService_GetAllHandlers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DbServiceServer).GetAllHandlers(ctx, req.(*GetAllHandlersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DbService_GetHandlersByGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetHandlersByGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DbServiceServer).GetHandlersByGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DbService_GetHandlersByGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DbServiceServer).GetHandlersByGroup(ctx, req.(*GetHandlersByGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DbService_GetBatchedHandlers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBatchedHandlersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DbServiceServer).GetBatchedHandlers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DbService_GetBatchedHandlers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DbServiceServer).GetBatchedHandlers(ctx, req.(*GetBatchedHandlersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DbService_UpdateHandler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateHandlerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DbServiceServer).UpdateHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DbService_UpdateHandler_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DbServiceServer).UpdateHandler(ctx, req.(*UpdateHandlerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DbService_ServiceDesc is the grpc.ServiceDesc for DbService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DbService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "x.db.DbService",
	HandlerType: (*DbServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetHandler",
			Handler:    _DbService_GetHandler_Handler,
		},
		{
			MethodName: "GetAllHandlers",
			Handler:    _DbService_GetAllHandlers_Handler,
		},
		{
			MethodName: "GetHandlersByGroup",
			Handler:    _DbService_GetHandlersByGroup_Handler,
		},
		{
			MethodName: "GetBatchedHandlers",
			Handler:    _DbService_GetBatchedHandlers_Handler,
		},
		{
			MethodName: "UpdateHandler",
			Handler:    _DbService_UpdateHandler_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/db/db.proto",
}
