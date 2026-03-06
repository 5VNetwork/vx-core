package grpcservice

import (
	context "context"
	userlogger "github.com/5vnetwork/vx-core/app/userlogger"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	GrpcService_Communicate_FullMethodName               = "/x.grpcservice.GrpcService/Communicate"
	GrpcService_AddInbound_FullMethodName                = "/x.grpcservice.GrpcService/AddInbound"
	GrpcService_RemoveInbound_FullMethodName             = "/x.grpcservice.GrpcService/RemoveInbound"
	GrpcService_GetStatsStream_FullMethodName            = "/x.grpcservice.GrpcService/GetStatsStream"
	GrpcService_SetOutboundHandlerSpeed_FullMethodName   = "/x.grpcservice.GrpcService/SetOutboundHandlerSpeed"
	GrpcService_UserLogStream_FullMethodName             = "/x.grpcservice.GrpcService/UserLogStream"
	GrpcService_ToggleUserLog_FullMethodName             = "/x.grpcservice.GrpcService/ToggleUserLog"
	GrpcService_ToggleLogAppId_FullMethodName            = "/x.grpcservice.GrpcService/ToggleLogAppId"
	GrpcService_ChangeOutbound_FullMethodName            = "/x.grpcservice.GrpcService/ChangeOutbound"
	GrpcService_CurrentOutbound_FullMethodName           = "/x.grpcservice.GrpcService/CurrentOutbound"
	GrpcService_ChangeRoutingMode_FullMethodName         = "/x.grpcservice.GrpcService/ChangeRoutingMode"
	GrpcService_ChangeSelector_FullMethodName            = "/x.grpcservice.GrpcService/ChangeSelector"
	GrpcService_UpdateSelectorBalancer_FullMethodName    = "/x.grpcservice.GrpcService/UpdateSelectorBalancer"
	GrpcService_UpdateSelectorFilter_FullMethodName      = "/x.grpcservice.GrpcService/UpdateSelectorFilter"
	GrpcService_NotifyHandlerChange_FullMethodName       = "/x.grpcservice.GrpcService/NotifyHandlerChange"
	GrpcService_SwitchFakeDns_FullMethodName             = "/x.grpcservice.GrpcService/SwitchFakeDns"
	GrpcService_UpdateGeo_FullMethodName                 = "/x.grpcservice.GrpcService/UpdateGeo"
	GrpcService_AddGeoDomain_FullMethodName              = "/x.grpcservice.GrpcService/AddGeoDomain"
	GrpcService_RemoveGeoDomain_FullMethodName           = "/x.grpcservice.GrpcService/RemoveGeoDomain"
	GrpcService_ReplaceGeoDomains_FullMethodName         = "/x.grpcservice.GrpcService/ReplaceGeoDomains"
	GrpcService_ReplaceGeoIPs_FullMethodName             = "/x.grpcservice.GrpcService/ReplaceGeoIPs"
	GrpcService_UpdateRouter_FullMethodName              = "/x.grpcservice.GrpcService/UpdateRouter"
	GrpcService_SetSubscriptionInterval_FullMethodName   = "/x.grpcservice.GrpcService/SetSubscriptionInterval"
	GrpcService_SetAutoSubscriptionUpdate_FullMethodName = "/x.grpcservice.GrpcService/SetAutoSubscriptionUpdate"
	GrpcService_RttTest_FullMethodName                   = "/x.grpcservice.GrpcService/RttTest"
)

// GrpcServiceClient is the client API for GrpcService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GrpcServiceClient interface {
	// server to client
	Communicate(ctx context.Context, in *CommunicateRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[CommunicateMessage], error)
	// inbound
	AddInbound(ctx context.Context, in *AddInboundRequest, opts ...grpc.CallOption) (*AddInboundResponse, error)
	RemoveInbound(ctx context.Context, in *RemoveInboundRequest, opts ...grpc.CallOption) (*RemoveInboundResponse, error)
	// stats
	GetStatsStream(ctx context.Context, in *GetStatsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StatsResponse], error)
	SetOutboundHandlerSpeed(ctx context.Context, in *SetOutboundHandlerSpeedRequest, opts ...grpc.CallOption) (*SetOutboundHandlerSpeedResponse, error)
	// log
	// rpc ChangeLogLevel(ChangeLogLevelRequest) returns (ChangeLogLevelResponse);
	// rpc LogStream(LogStreamRequest) returns (stream LogMessage);
	UserLogStream(ctx context.Context, in *UserLogStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[userlogger.UserLogMessage], error)
	ToggleUserLog(ctx context.Context, in *ToggleUserLogRequest, opts ...grpc.CallOption) (*ToggleUserLogResponse, error)
	ToggleLogAppId(ctx context.Context, in *ToggleLogAppIdRequest, opts ...grpc.CallOption) (*ToggleLogAppIdResponse, error)
	// outbound
	ChangeOutbound(ctx context.Context, in *ChangeOutboundRequest, opts ...grpc.CallOption) (*ChangeOutboundResponse, error)
	CurrentOutbound(ctx context.Context, in *CurrentOutboundRequest, opts ...grpc.CallOption) (*CurrentOutboundResponse, error)
	// routing
	ChangeRoutingMode(ctx context.Context, in *ChangeRoutingModeRequest, opts ...grpc.CallOption) (*ChangeRoutingModeResponse, error)
	ChangeSelector(ctx context.Context, in *ChangeSelectorRequest, opts ...grpc.CallOption) (*ChangeSelectorResponse, error)
	UpdateSelectorBalancer(ctx context.Context, in *UpdateSelectorBalancerRequest, opts ...grpc.CallOption) (*Receipt, error)
	UpdateSelectorFilter(ctx context.Context, in *UpdateSelectorFilterRequest, opts ...grpc.CallOption) (*Receipt, error)
	NotifyHandlerChange(ctx context.Context, in *HandlerChangeNotify, opts ...grpc.CallOption) (*HandlerChangeNotifyResponse, error)
	// fake dns
	SwitchFakeDns(ctx context.Context, in *SwitchFakeDnsRequest, opts ...grpc.CallOption) (*SwitchFakeDnsResponse, error)
	// geo
	UpdateGeo(ctx context.Context, in *UpdateGeoRequest, opts ...grpc.CallOption) (*UpdateGeoResponse, error)
	AddGeoDomain(ctx context.Context, in *AddGeoDomainRequest, opts ...grpc.CallOption) (*Receipt, error)
	RemoveGeoDomain(ctx context.Context, in *RemoveGeoDomainRequest, opts ...grpc.CallOption) (*Receipt, error)
	ReplaceGeoDomains(ctx context.Context, in *ReplaceDomainSetRequest, opts ...grpc.CallOption) (*Receipt, error)
	ReplaceGeoIPs(ctx context.Context, in *ReplaceIPSetRequest, opts ...grpc.CallOption) (*Receipt, error)
	// app id
	UpdateRouter(ctx context.Context, in *UpdateRouterRequest, opts ...grpc.CallOption) (*UpdateRouterResponse, error)
	// subscription
	SetSubscriptionInterval(ctx context.Context, in *SetSubscriptionIntervalRequest, opts ...grpc.CallOption) (*SetSubscriptionIntervalResponse, error)
	SetAutoSubscriptionUpdate(ctx context.Context, in *SetAutoSubscriptionUpdateRequest, opts ...grpc.CallOption) (*Receipt, error)
	RttTest(ctx context.Context, in *RttTestRequest, opts ...grpc.CallOption) (*RttTestResponse, error)
}

type grpcServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGrpcServiceClient(cc grpc.ClientConnInterface) GrpcServiceClient {
	return &grpcServiceClient{cc}
}

func (c *grpcServiceClient) Communicate(ctx context.Context, in *CommunicateRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[CommunicateMessage], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &GrpcService_ServiceDesc.Streams[0], GrpcService_Communicate_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[CommunicateRequest, CommunicateMessage]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GrpcService_CommunicateClient = grpc.ServerStreamingClient[CommunicateMessage]

func (c *grpcServiceClient) AddInbound(ctx context.Context, in *AddInboundRequest, opts ...grpc.CallOption) (*AddInboundResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddInboundResponse)
	err := c.cc.Invoke(ctx, GrpcService_AddInbound_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) RemoveInbound(ctx context.Context, in *RemoveInboundRequest, opts ...grpc.CallOption) (*RemoveInboundResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RemoveInboundResponse)
	err := c.cc.Invoke(ctx, GrpcService_RemoveInbound_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) GetStatsStream(ctx context.Context, in *GetStatsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StatsResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &GrpcService_ServiceDesc.Streams[1], GrpcService_GetStatsStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[GetStatsRequest, StatsResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GrpcService_GetStatsStreamClient = grpc.ServerStreamingClient[StatsResponse]

func (c *grpcServiceClient) SetOutboundHandlerSpeed(ctx context.Context, in *SetOutboundHandlerSpeedRequest, opts ...grpc.CallOption) (*SetOutboundHandlerSpeedResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetOutboundHandlerSpeedResponse)
	err := c.cc.Invoke(ctx, GrpcService_SetOutboundHandlerSpeed_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) UserLogStream(ctx context.Context, in *UserLogStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[userlogger.UserLogMessage], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &GrpcService_ServiceDesc.Streams[2], GrpcService_UserLogStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[UserLogStreamRequest, userlogger.UserLogMessage]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GrpcService_UserLogStreamClient = grpc.ServerStreamingClient[userlogger.UserLogMessage]

func (c *grpcServiceClient) ToggleUserLog(ctx context.Context, in *ToggleUserLogRequest, opts ...grpc.CallOption) (*ToggleUserLogResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ToggleUserLogResponse)
	err := c.cc.Invoke(ctx, GrpcService_ToggleUserLog_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) ToggleLogAppId(ctx context.Context, in *ToggleLogAppIdRequest, opts ...grpc.CallOption) (*ToggleLogAppIdResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ToggleLogAppIdResponse)
	err := c.cc.Invoke(ctx, GrpcService_ToggleLogAppId_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) ChangeOutbound(ctx context.Context, in *ChangeOutboundRequest, opts ...grpc.CallOption) (*ChangeOutboundResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChangeOutboundResponse)
	err := c.cc.Invoke(ctx, GrpcService_ChangeOutbound_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) CurrentOutbound(ctx context.Context, in *CurrentOutboundRequest, opts ...grpc.CallOption) (*CurrentOutboundResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CurrentOutboundResponse)
	err := c.cc.Invoke(ctx, GrpcService_CurrentOutbound_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) ChangeRoutingMode(ctx context.Context, in *ChangeRoutingModeRequest, opts ...grpc.CallOption) (*ChangeRoutingModeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChangeRoutingModeResponse)
	err := c.cc.Invoke(ctx, GrpcService_ChangeRoutingMode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) ChangeSelector(ctx context.Context, in *ChangeSelectorRequest, opts ...grpc.CallOption) (*ChangeSelectorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChangeSelectorResponse)
	err := c.cc.Invoke(ctx, GrpcService_ChangeSelector_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) UpdateSelectorBalancer(ctx context.Context, in *UpdateSelectorBalancerRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, GrpcService_UpdateSelectorBalancer_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) UpdateSelectorFilter(ctx context.Context, in *UpdateSelectorFilterRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, GrpcService_UpdateSelectorFilter_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) NotifyHandlerChange(ctx context.Context, in *HandlerChangeNotify, opts ...grpc.CallOption) (*HandlerChangeNotifyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HandlerChangeNotifyResponse)
	err := c.cc.Invoke(ctx, GrpcService_NotifyHandlerChange_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) SwitchFakeDns(ctx context.Context, in *SwitchFakeDnsRequest, opts ...grpc.CallOption) (*SwitchFakeDnsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SwitchFakeDnsResponse)
	err := c.cc.Invoke(ctx, GrpcService_SwitchFakeDns_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) UpdateGeo(ctx context.Context, in *UpdateGeoRequest, opts ...grpc.CallOption) (*UpdateGeoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateGeoResponse)
	err := c.cc.Invoke(ctx, GrpcService_UpdateGeo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) AddGeoDomain(ctx context.Context, in *AddGeoDomainRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, GrpcService_AddGeoDomain_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) RemoveGeoDomain(ctx context.Context, in *RemoveGeoDomainRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, GrpcService_RemoveGeoDomain_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) ReplaceGeoDomains(ctx context.Context, in *ReplaceDomainSetRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, GrpcService_ReplaceGeoDomains_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) ReplaceGeoIPs(ctx context.Context, in *ReplaceIPSetRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, GrpcService_ReplaceGeoIPs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) UpdateRouter(ctx context.Context, in *UpdateRouterRequest, opts ...grpc.CallOption) (*UpdateRouterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateRouterResponse)
	err := c.cc.Invoke(ctx, GrpcService_UpdateRouter_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) SetSubscriptionInterval(ctx context.Context, in *SetSubscriptionIntervalRequest, opts ...grpc.CallOption) (*SetSubscriptionIntervalResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetSubscriptionIntervalResponse)
	err := c.cc.Invoke(ctx, GrpcService_SetSubscriptionInterval_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) SetAutoSubscriptionUpdate(ctx context.Context, in *SetAutoSubscriptionUpdateRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, GrpcService_SetAutoSubscriptionUpdate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcServiceClient) RttTest(ctx context.Context, in *RttTestRequest, opts ...grpc.CallOption) (*RttTestResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RttTestResponse)
	err := c.cc.Invoke(ctx, GrpcService_RttTest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GrpcServiceServer is the server API for GrpcService service.
// All implementations must embed UnimplementedGrpcServiceServer
// for forward compatibility.
type GrpcServiceServer interface {
	// server to client
	Communicate(*CommunicateRequest, grpc.ServerStreamingServer[CommunicateMessage]) error
	// inbound
	AddInbound(context.Context, *AddInboundRequest) (*AddInboundResponse, error)
	RemoveInbound(context.Context, *RemoveInboundRequest) (*RemoveInboundResponse, error)
	// stats
	GetStatsStream(*GetStatsRequest, grpc.ServerStreamingServer[StatsResponse]) error
	SetOutboundHandlerSpeed(context.Context, *SetOutboundHandlerSpeedRequest) (*SetOutboundHandlerSpeedResponse, error)
	// log
	// rpc ChangeLogLevel(ChangeLogLevelRequest) returns (ChangeLogLevelResponse);
	// rpc LogStream(LogStreamRequest) returns (stream LogMessage);
	UserLogStream(*UserLogStreamRequest, grpc.ServerStreamingServer[userlogger.UserLogMessage]) error
	ToggleUserLog(context.Context, *ToggleUserLogRequest) (*ToggleUserLogResponse, error)
	ToggleLogAppId(context.Context, *ToggleLogAppIdRequest) (*ToggleLogAppIdResponse, error)
	// outbound
	ChangeOutbound(context.Context, *ChangeOutboundRequest) (*ChangeOutboundResponse, error)
	CurrentOutbound(context.Context, *CurrentOutboundRequest) (*CurrentOutboundResponse, error)
	// routing
	ChangeRoutingMode(context.Context, *ChangeRoutingModeRequest) (*ChangeRoutingModeResponse, error)
	ChangeSelector(context.Context, *ChangeSelectorRequest) (*ChangeSelectorResponse, error)
	UpdateSelectorBalancer(context.Context, *UpdateSelectorBalancerRequest) (*Receipt, error)
	UpdateSelectorFilter(context.Context, *UpdateSelectorFilterRequest) (*Receipt, error)
	NotifyHandlerChange(context.Context, *HandlerChangeNotify) (*HandlerChangeNotifyResponse, error)
	// fake dns
	SwitchFakeDns(context.Context, *SwitchFakeDnsRequest) (*SwitchFakeDnsResponse, error)
	// geo
	UpdateGeo(context.Context, *UpdateGeoRequest) (*UpdateGeoResponse, error)
	AddGeoDomain(context.Context, *AddGeoDomainRequest) (*Receipt, error)
	RemoveGeoDomain(context.Context, *RemoveGeoDomainRequest) (*Receipt, error)
	ReplaceGeoDomains(context.Context, *ReplaceDomainSetRequest) (*Receipt, error)
	ReplaceGeoIPs(context.Context, *ReplaceIPSetRequest) (*Receipt, error)
	// app id
	UpdateRouter(context.Context, *UpdateRouterRequest) (*UpdateRouterResponse, error)
	// subscription
	SetSubscriptionInterval(context.Context, *SetSubscriptionIntervalRequest) (*SetSubscriptionIntervalResponse, error)
	SetAutoSubscriptionUpdate(context.Context, *SetAutoSubscriptionUpdateRequest) (*Receipt, error)
	RttTest(context.Context, *RttTestRequest) (*RttTestResponse, error)
	mustEmbedUnimplementedGrpcServiceServer()
}

// UnimplementedGrpcServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedGrpcServiceServer struct{}

func (UnimplementedGrpcServiceServer) Communicate(*CommunicateRequest, grpc.ServerStreamingServer[CommunicateMessage]) error {
	return status.Errorf(codes.Unimplemented, "method Communicate not implemented")
}
func (UnimplementedGrpcServiceServer) AddInbound(context.Context, *AddInboundRequest) (*AddInboundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddInbound not implemented")
}
func (UnimplementedGrpcServiceServer) RemoveInbound(context.Context, *RemoveInboundRequest) (*RemoveInboundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveInbound not implemented")
}
func (UnimplementedGrpcServiceServer) GetStatsStream(*GetStatsRequest, grpc.ServerStreamingServer[StatsResponse]) error {
	return status.Errorf(codes.Unimplemented, "method GetStatsStream not implemented")
}
func (UnimplementedGrpcServiceServer) SetOutboundHandlerSpeed(context.Context, *SetOutboundHandlerSpeedRequest) (*SetOutboundHandlerSpeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetOutboundHandlerSpeed not implemented")
}
func (UnimplementedGrpcServiceServer) UserLogStream(*UserLogStreamRequest, grpc.ServerStreamingServer[userlogger.UserLogMessage]) error {
	return status.Errorf(codes.Unimplemented, "method UserLogStream not implemented")
}
func (UnimplementedGrpcServiceServer) ToggleUserLog(context.Context, *ToggleUserLogRequest) (*ToggleUserLogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleUserLog not implemented")
}
func (UnimplementedGrpcServiceServer) ToggleLogAppId(context.Context, *ToggleLogAppIdRequest) (*ToggleLogAppIdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleLogAppId not implemented")
}
func (UnimplementedGrpcServiceServer) ChangeOutbound(context.Context, *ChangeOutboundRequest) (*ChangeOutboundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeOutbound not implemented")
}
func (UnimplementedGrpcServiceServer) CurrentOutbound(context.Context, *CurrentOutboundRequest) (*CurrentOutboundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CurrentOutbound not implemented")
}
func (UnimplementedGrpcServiceServer) ChangeRoutingMode(context.Context, *ChangeRoutingModeRequest) (*ChangeRoutingModeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeRoutingMode not implemented")
}
func (UnimplementedGrpcServiceServer) ChangeSelector(context.Context, *ChangeSelectorRequest) (*ChangeSelectorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeSelector not implemented")
}
func (UnimplementedGrpcServiceServer) UpdateSelectorBalancer(context.Context, *UpdateSelectorBalancerRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSelectorBalancer not implemented")
}
func (UnimplementedGrpcServiceServer) UpdateSelectorFilter(context.Context, *UpdateSelectorFilterRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSelectorFilter not implemented")
}
func (UnimplementedGrpcServiceServer) NotifyHandlerChange(context.Context, *HandlerChangeNotify) (*HandlerChangeNotifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyHandlerChange not implemented")
}
func (UnimplementedGrpcServiceServer) SwitchFakeDns(context.Context, *SwitchFakeDnsRequest) (*SwitchFakeDnsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SwitchFakeDns not implemented")
}
func (UnimplementedGrpcServiceServer) UpdateGeo(context.Context, *UpdateGeoRequest) (*UpdateGeoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGeo not implemented")
}
func (UnimplementedGrpcServiceServer) AddGeoDomain(context.Context, *AddGeoDomainRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddGeoDomain not implemented")
}
func (UnimplementedGrpcServiceServer) RemoveGeoDomain(context.Context, *RemoveGeoDomainRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveGeoDomain not implemented")
}
func (UnimplementedGrpcServiceServer) ReplaceGeoDomains(context.Context, *ReplaceDomainSetRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplaceGeoDomains not implemented")
}
func (UnimplementedGrpcServiceServer) ReplaceGeoIPs(context.Context, *ReplaceIPSetRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplaceGeoIPs not implemented")
}
func (UnimplementedGrpcServiceServer) UpdateRouter(context.Context, *UpdateRouterRequest) (*UpdateRouterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRouter not implemented")
}
func (UnimplementedGrpcServiceServer) SetSubscriptionInterval(context.Context, *SetSubscriptionIntervalRequest) (*SetSubscriptionIntervalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetSubscriptionInterval not implemented")
}
func (UnimplementedGrpcServiceServer) SetAutoSubscriptionUpdate(context.Context, *SetAutoSubscriptionUpdateRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetAutoSubscriptionUpdate not implemented")
}
func (UnimplementedGrpcServiceServer) RttTest(context.Context, *RttTestRequest) (*RttTestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RttTest not implemented")
}
func (UnimplementedGrpcServiceServer) mustEmbedUnimplementedGrpcServiceServer() {}
func (UnimplementedGrpcServiceServer) testEmbeddedByValue()                     {}

// UnsafeGrpcServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GrpcServiceServer will
// result in compilation errors.
type UnsafeGrpcServiceServer interface {
	mustEmbedUnimplementedGrpcServiceServer()
}

func RegisterGrpcServiceServer(s grpc.ServiceRegistrar, srv GrpcServiceServer) {
	// If the following call pancis, it indicates UnimplementedGrpcServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&GrpcService_ServiceDesc, srv)
}

func _GrpcService_Communicate_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CommunicateRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GrpcServiceServer).Communicate(m, &grpc.GenericServerStream[CommunicateRequest, CommunicateMessage]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GrpcService_CommunicateServer = grpc.ServerStreamingServer[CommunicateMessage]

func _GrpcService_AddInbound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddInboundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).AddInbound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_AddInbound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).AddInbound(ctx, req.(*AddInboundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_RemoveInbound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveInboundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).RemoveInbound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_RemoveInbound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).RemoveInbound(ctx, req.(*RemoveInboundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_GetStatsStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetStatsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GrpcServiceServer).GetStatsStream(m, &grpc.GenericServerStream[GetStatsRequest, StatsResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GrpcService_GetStatsStreamServer = grpc.ServerStreamingServer[StatsResponse]

func _GrpcService_SetOutboundHandlerSpeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetOutboundHandlerSpeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).SetOutboundHandlerSpeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_SetOutboundHandlerSpeed_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).SetOutboundHandlerSpeed(ctx, req.(*SetOutboundHandlerSpeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_UserLogStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(UserLogStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GrpcServiceServer).UserLogStream(m, &grpc.GenericServerStream[UserLogStreamRequest, userlogger.UserLogMessage]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GrpcService_UserLogStreamServer = grpc.ServerStreamingServer[userlogger.UserLogMessage]

func _GrpcService_ToggleUserLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ToggleUserLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).ToggleUserLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_ToggleUserLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).ToggleUserLog(ctx, req.(*ToggleUserLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_ToggleLogAppId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ToggleLogAppIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).ToggleLogAppId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_ToggleLogAppId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).ToggleLogAppId(ctx, req.(*ToggleLogAppIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_ChangeOutbound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangeOutboundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).ChangeOutbound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_ChangeOutbound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).ChangeOutbound(ctx, req.(*ChangeOutboundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_CurrentOutbound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CurrentOutboundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).CurrentOutbound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_CurrentOutbound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).CurrentOutbound(ctx, req.(*CurrentOutboundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_ChangeRoutingMode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangeRoutingModeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).ChangeRoutingMode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_ChangeRoutingMode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).ChangeRoutingMode(ctx, req.(*ChangeRoutingModeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_ChangeSelector_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangeSelectorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).ChangeSelector(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_ChangeSelector_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).ChangeSelector(ctx, req.(*ChangeSelectorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_UpdateSelectorBalancer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSelectorBalancerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).UpdateSelectorBalancer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_UpdateSelectorBalancer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).UpdateSelectorBalancer(ctx, req.(*UpdateSelectorBalancerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_UpdateSelectorFilter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSelectorFilterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).UpdateSelectorFilter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_UpdateSelectorFilter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).UpdateSelectorFilter(ctx, req.(*UpdateSelectorFilterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_NotifyHandlerChange_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandlerChangeNotify)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).NotifyHandlerChange(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_NotifyHandlerChange_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).NotifyHandlerChange(ctx, req.(*HandlerChangeNotify))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_SwitchFakeDns_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SwitchFakeDnsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).SwitchFakeDns(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_SwitchFakeDns_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).SwitchFakeDns(ctx, req.(*SwitchFakeDnsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_UpdateGeo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateGeoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).UpdateGeo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_UpdateGeo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).UpdateGeo(ctx, req.(*UpdateGeoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_AddGeoDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddGeoDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).AddGeoDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_AddGeoDomain_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).AddGeoDomain(ctx, req.(*AddGeoDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_RemoveGeoDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveGeoDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).RemoveGeoDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_RemoveGeoDomain_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).RemoveGeoDomain(ctx, req.(*RemoveGeoDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_ReplaceGeoDomains_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplaceDomainSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).ReplaceGeoDomains(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_ReplaceGeoDomains_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).ReplaceGeoDomains(ctx, req.(*ReplaceDomainSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_ReplaceGeoIPs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplaceIPSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).ReplaceGeoIPs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_ReplaceGeoIPs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).ReplaceGeoIPs(ctx, req.(*ReplaceIPSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_UpdateRouter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRouterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).UpdateRouter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_UpdateRouter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).UpdateRouter(ctx, req.(*UpdateRouterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_SetSubscriptionInterval_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetSubscriptionIntervalRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).SetSubscriptionInterval(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_SetSubscriptionInterval_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).SetSubscriptionInterval(ctx, req.(*SetSubscriptionIntervalRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_SetAutoSubscriptionUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetAutoSubscriptionUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).SetAutoSubscriptionUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_SetAutoSubscriptionUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).SetAutoSubscriptionUpdate(ctx, req.(*SetAutoSubscriptionUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcService_RttTest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RttTestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServiceServer).RttTest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GrpcService_RttTest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServiceServer).RttTest(ctx, req.(*RttTestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GrpcService_ServiceDesc is the grpc.ServiceDesc for GrpcService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GrpcService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "x.grpcservice.GrpcService",
	HandlerType: (*GrpcServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddInbound",
			Handler:    _GrpcService_AddInbound_Handler,
		},
		{
			MethodName: "RemoveInbound",
			Handler:    _GrpcService_RemoveInbound_Handler,
		},
		{
			MethodName: "SetOutboundHandlerSpeed",
			Handler:    _GrpcService_SetOutboundHandlerSpeed_Handler,
		},
		{
			MethodName: "ToggleUserLog",
			Handler:    _GrpcService_ToggleUserLog_Handler,
		},
		{
			MethodName: "ToggleLogAppId",
			Handler:    _GrpcService_ToggleLogAppId_Handler,
		},
		{
			MethodName: "ChangeOutbound",
			Handler:    _GrpcService_ChangeOutbound_Handler,
		},
		{
			MethodName: "CurrentOutbound",
			Handler:    _GrpcService_CurrentOutbound_Handler,
		},
		{
			MethodName: "ChangeRoutingMode",
			Handler:    _GrpcService_ChangeRoutingMode_Handler,
		},
		{
			MethodName: "ChangeSelector",
			Handler:    _GrpcService_ChangeSelector_Handler,
		},
		{
			MethodName: "UpdateSelectorBalancer",
			Handler:    _GrpcService_UpdateSelectorBalancer_Handler,
		},
		{
			MethodName: "UpdateSelectorFilter",
			Handler:    _GrpcService_UpdateSelectorFilter_Handler,
		},
		{
			MethodName: "NotifyHandlerChange",
			Handler:    _GrpcService_NotifyHandlerChange_Handler,
		},
		{
			MethodName: "SwitchFakeDns",
			Handler:    _GrpcService_SwitchFakeDns_Handler,
		},
		{
			MethodName: "UpdateGeo",
			Handler:    _GrpcService_UpdateGeo_Handler,
		},
		{
			MethodName: "AddGeoDomain",
			Handler:    _GrpcService_AddGeoDomain_Handler,
		},
		{
			MethodName: "RemoveGeoDomain",
			Handler:    _GrpcService_RemoveGeoDomain_Handler,
		},
		{
			MethodName: "ReplaceGeoDomains",
			Handler:    _GrpcService_ReplaceGeoDomains_Handler,
		},
		{
			MethodName: "ReplaceGeoIPs",
			Handler:    _GrpcService_ReplaceGeoIPs_Handler,
		},
		{
			MethodName: "UpdateRouter",
			Handler:    _GrpcService_UpdateRouter_Handler,
		},
		{
			MethodName: "SetSubscriptionInterval",
			Handler:    _GrpcService_SetSubscriptionInterval_Handler,
		},
		{
			MethodName: "SetAutoSubscriptionUpdate",
			Handler:    _GrpcService_SetAutoSubscriptionUpdate_Handler,
		},
		{
			MethodName: "RttTest",
			Handler:    _GrpcService_RttTest_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Communicate",
			Handler:       _GrpcService_Communicate_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetStatsStream",
			Handler:       _GrpcService_GetStatsStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "UserLogStream",
			Handler:       _GrpcService_UserLogStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "app/grpcservice/grpc.proto",
}
