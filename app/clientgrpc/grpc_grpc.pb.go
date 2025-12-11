package clientgrpc

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
	ClientService_Communicate_FullMethodName               = "/x.clientgrpc.ClientService/Communicate"
	ClientService_AddInbound_FullMethodName                = "/x.clientgrpc.ClientService/AddInbound"
	ClientService_RemoveInbound_FullMethodName             = "/x.clientgrpc.ClientService/RemoveInbound"
	ClientService_GetStatsStream_FullMethodName            = "/x.clientgrpc.ClientService/GetStatsStream"
	ClientService_SetOutboundHandlerSpeed_FullMethodName   = "/x.clientgrpc.ClientService/SetOutboundHandlerSpeed"
	ClientService_UserLogStream_FullMethodName             = "/x.clientgrpc.ClientService/UserLogStream"
	ClientService_ToggleUserLog_FullMethodName             = "/x.clientgrpc.ClientService/ToggleUserLog"
	ClientService_ToggleLogAppId_FullMethodName            = "/x.clientgrpc.ClientService/ToggleLogAppId"
	ClientService_ChangeOutbound_FullMethodName            = "/x.clientgrpc.ClientService/ChangeOutbound"
	ClientService_CurrentOutbound_FullMethodName           = "/x.clientgrpc.ClientService/CurrentOutbound"
	ClientService_ChangeRoutingMode_FullMethodName         = "/x.clientgrpc.ClientService/ChangeRoutingMode"
	ClientService_ChangeSelector_FullMethodName            = "/x.clientgrpc.ClientService/ChangeSelector"
	ClientService_UpdateSelectorBalancer_FullMethodName    = "/x.clientgrpc.ClientService/UpdateSelectorBalancer"
	ClientService_UpdateSelectorFilter_FullMethodName      = "/x.clientgrpc.ClientService/UpdateSelectorFilter"
	ClientService_NotifyHandlerChange_FullMethodName       = "/x.clientgrpc.ClientService/NotifyHandlerChange"
	ClientService_SwitchFakeDns_FullMethodName             = "/x.clientgrpc.ClientService/SwitchFakeDns"
	ClientService_UpdateGeo_FullMethodName                 = "/x.clientgrpc.ClientService/UpdateGeo"
	ClientService_AddGeoDomain_FullMethodName              = "/x.clientgrpc.ClientService/AddGeoDomain"
	ClientService_RemoveGeoDomain_FullMethodName           = "/x.clientgrpc.ClientService/RemoveGeoDomain"
	ClientService_ReplaceGeoDomains_FullMethodName         = "/x.clientgrpc.ClientService/ReplaceGeoDomains"
	ClientService_ReplaceGeoIPs_FullMethodName             = "/x.clientgrpc.ClientService/ReplaceGeoIPs"
	ClientService_UpdateRouter_FullMethodName              = "/x.clientgrpc.ClientService/UpdateRouter"
	ClientService_SetSubscriptionInterval_FullMethodName   = "/x.clientgrpc.ClientService/SetSubscriptionInterval"
	ClientService_SetAutoSubscriptionUpdate_FullMethodName = "/x.clientgrpc.ClientService/SetAutoSubscriptionUpdate"
	ClientService_RttTest_FullMethodName                   = "/x.clientgrpc.ClientService/RttTest"
)

// ClientServiceClient is the client API for ClientService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClientServiceClient interface {
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

type clientServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClientServiceClient(cc grpc.ClientConnInterface) ClientServiceClient {
	return &clientServiceClient{cc}
}

func (c *clientServiceClient) Communicate(ctx context.Context, in *CommunicateRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[CommunicateMessage], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &ClientService_ServiceDesc.Streams[0], ClientService_Communicate_FullMethodName, cOpts...)
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
type ClientService_CommunicateClient = grpc.ServerStreamingClient[CommunicateMessage]

func (c *clientServiceClient) AddInbound(ctx context.Context, in *AddInboundRequest, opts ...grpc.CallOption) (*AddInboundResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddInboundResponse)
	err := c.cc.Invoke(ctx, ClientService_AddInbound_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) RemoveInbound(ctx context.Context, in *RemoveInboundRequest, opts ...grpc.CallOption) (*RemoveInboundResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RemoveInboundResponse)
	err := c.cc.Invoke(ctx, ClientService_RemoveInbound_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) GetStatsStream(ctx context.Context, in *GetStatsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StatsResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &ClientService_ServiceDesc.Streams[1], ClientService_GetStatsStream_FullMethodName, cOpts...)
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
type ClientService_GetStatsStreamClient = grpc.ServerStreamingClient[StatsResponse]

func (c *clientServiceClient) SetOutboundHandlerSpeed(ctx context.Context, in *SetOutboundHandlerSpeedRequest, opts ...grpc.CallOption) (*SetOutboundHandlerSpeedResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetOutboundHandlerSpeedResponse)
	err := c.cc.Invoke(ctx, ClientService_SetOutboundHandlerSpeed_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) UserLogStream(ctx context.Context, in *UserLogStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[userlogger.UserLogMessage], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &ClientService_ServiceDesc.Streams[2], ClientService_UserLogStream_FullMethodName, cOpts...)
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
type ClientService_UserLogStreamClient = grpc.ServerStreamingClient[userlogger.UserLogMessage]

func (c *clientServiceClient) ToggleUserLog(ctx context.Context, in *ToggleUserLogRequest, opts ...grpc.CallOption) (*ToggleUserLogResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ToggleUserLogResponse)
	err := c.cc.Invoke(ctx, ClientService_ToggleUserLog_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) ToggleLogAppId(ctx context.Context, in *ToggleLogAppIdRequest, opts ...grpc.CallOption) (*ToggleLogAppIdResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ToggleLogAppIdResponse)
	err := c.cc.Invoke(ctx, ClientService_ToggleLogAppId_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) ChangeOutbound(ctx context.Context, in *ChangeOutboundRequest, opts ...grpc.CallOption) (*ChangeOutboundResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChangeOutboundResponse)
	err := c.cc.Invoke(ctx, ClientService_ChangeOutbound_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) CurrentOutbound(ctx context.Context, in *CurrentOutboundRequest, opts ...grpc.CallOption) (*CurrentOutboundResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CurrentOutboundResponse)
	err := c.cc.Invoke(ctx, ClientService_CurrentOutbound_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) ChangeRoutingMode(ctx context.Context, in *ChangeRoutingModeRequest, opts ...grpc.CallOption) (*ChangeRoutingModeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChangeRoutingModeResponse)
	err := c.cc.Invoke(ctx, ClientService_ChangeRoutingMode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) ChangeSelector(ctx context.Context, in *ChangeSelectorRequest, opts ...grpc.CallOption) (*ChangeSelectorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChangeSelectorResponse)
	err := c.cc.Invoke(ctx, ClientService_ChangeSelector_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) UpdateSelectorBalancer(ctx context.Context, in *UpdateSelectorBalancerRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, ClientService_UpdateSelectorBalancer_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) UpdateSelectorFilter(ctx context.Context, in *UpdateSelectorFilterRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, ClientService_UpdateSelectorFilter_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) NotifyHandlerChange(ctx context.Context, in *HandlerChangeNotify, opts ...grpc.CallOption) (*HandlerChangeNotifyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HandlerChangeNotifyResponse)
	err := c.cc.Invoke(ctx, ClientService_NotifyHandlerChange_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) SwitchFakeDns(ctx context.Context, in *SwitchFakeDnsRequest, opts ...grpc.CallOption) (*SwitchFakeDnsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SwitchFakeDnsResponse)
	err := c.cc.Invoke(ctx, ClientService_SwitchFakeDns_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) UpdateGeo(ctx context.Context, in *UpdateGeoRequest, opts ...grpc.CallOption) (*UpdateGeoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateGeoResponse)
	err := c.cc.Invoke(ctx, ClientService_UpdateGeo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) AddGeoDomain(ctx context.Context, in *AddGeoDomainRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, ClientService_AddGeoDomain_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) RemoveGeoDomain(ctx context.Context, in *RemoveGeoDomainRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, ClientService_RemoveGeoDomain_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) ReplaceGeoDomains(ctx context.Context, in *ReplaceDomainSetRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, ClientService_ReplaceGeoDomains_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) ReplaceGeoIPs(ctx context.Context, in *ReplaceIPSetRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, ClientService_ReplaceGeoIPs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) UpdateRouter(ctx context.Context, in *UpdateRouterRequest, opts ...grpc.CallOption) (*UpdateRouterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateRouterResponse)
	err := c.cc.Invoke(ctx, ClientService_UpdateRouter_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) SetSubscriptionInterval(ctx context.Context, in *SetSubscriptionIntervalRequest, opts ...grpc.CallOption) (*SetSubscriptionIntervalResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetSubscriptionIntervalResponse)
	err := c.cc.Invoke(ctx, ClientService_SetSubscriptionInterval_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) SetAutoSubscriptionUpdate(ctx context.Context, in *SetAutoSubscriptionUpdateRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, ClientService_SetAutoSubscriptionUpdate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) RttTest(ctx context.Context, in *RttTestRequest, opts ...grpc.CallOption) (*RttTestResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RttTestResponse)
	err := c.cc.Invoke(ctx, ClientService_RttTest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClientServiceServer is the server API for ClientService service.
// All implementations must embed UnimplementedClientServiceServer
// for forward compatibility.
type ClientServiceServer interface {
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
	mustEmbedUnimplementedClientServiceServer()
}

// UnimplementedClientServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedClientServiceServer struct{}

func (UnimplementedClientServiceServer) Communicate(*CommunicateRequest, grpc.ServerStreamingServer[CommunicateMessage]) error {
	return status.Errorf(codes.Unimplemented, "method Communicate not implemented")
}
func (UnimplementedClientServiceServer) AddInbound(context.Context, *AddInboundRequest) (*AddInboundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddInbound not implemented")
}
func (UnimplementedClientServiceServer) RemoveInbound(context.Context, *RemoveInboundRequest) (*RemoveInboundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveInbound not implemented")
}
func (UnimplementedClientServiceServer) GetStatsStream(*GetStatsRequest, grpc.ServerStreamingServer[StatsResponse]) error {
	return status.Errorf(codes.Unimplemented, "method GetStatsStream not implemented")
}
func (UnimplementedClientServiceServer) SetOutboundHandlerSpeed(context.Context, *SetOutboundHandlerSpeedRequest) (*SetOutboundHandlerSpeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetOutboundHandlerSpeed not implemented")
}
func (UnimplementedClientServiceServer) UserLogStream(*UserLogStreamRequest, grpc.ServerStreamingServer[userlogger.UserLogMessage]) error {
	return status.Errorf(codes.Unimplemented, "method UserLogStream not implemented")
}
func (UnimplementedClientServiceServer) ToggleUserLog(context.Context, *ToggleUserLogRequest) (*ToggleUserLogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleUserLog not implemented")
}
func (UnimplementedClientServiceServer) ToggleLogAppId(context.Context, *ToggleLogAppIdRequest) (*ToggleLogAppIdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleLogAppId not implemented")
}
func (UnimplementedClientServiceServer) ChangeOutbound(context.Context, *ChangeOutboundRequest) (*ChangeOutboundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeOutbound not implemented")
}
func (UnimplementedClientServiceServer) CurrentOutbound(context.Context, *CurrentOutboundRequest) (*CurrentOutboundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CurrentOutbound not implemented")
}
func (UnimplementedClientServiceServer) ChangeRoutingMode(context.Context, *ChangeRoutingModeRequest) (*ChangeRoutingModeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeRoutingMode not implemented")
}
func (UnimplementedClientServiceServer) ChangeSelector(context.Context, *ChangeSelectorRequest) (*ChangeSelectorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeSelector not implemented")
}
func (UnimplementedClientServiceServer) UpdateSelectorBalancer(context.Context, *UpdateSelectorBalancerRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSelectorBalancer not implemented")
}
func (UnimplementedClientServiceServer) UpdateSelectorFilter(context.Context, *UpdateSelectorFilterRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSelectorFilter not implemented")
}
func (UnimplementedClientServiceServer) NotifyHandlerChange(context.Context, *HandlerChangeNotify) (*HandlerChangeNotifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyHandlerChange not implemented")
}
func (UnimplementedClientServiceServer) SwitchFakeDns(context.Context, *SwitchFakeDnsRequest) (*SwitchFakeDnsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SwitchFakeDns not implemented")
}
func (UnimplementedClientServiceServer) UpdateGeo(context.Context, *UpdateGeoRequest) (*UpdateGeoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGeo not implemented")
}
func (UnimplementedClientServiceServer) AddGeoDomain(context.Context, *AddGeoDomainRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddGeoDomain not implemented")
}
func (UnimplementedClientServiceServer) RemoveGeoDomain(context.Context, *RemoveGeoDomainRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveGeoDomain not implemented")
}
func (UnimplementedClientServiceServer) ReplaceGeoDomains(context.Context, *ReplaceDomainSetRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplaceGeoDomains not implemented")
}
func (UnimplementedClientServiceServer) ReplaceGeoIPs(context.Context, *ReplaceIPSetRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplaceGeoIPs not implemented")
}
func (UnimplementedClientServiceServer) UpdateRouter(context.Context, *UpdateRouterRequest) (*UpdateRouterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRouter not implemented")
}
func (UnimplementedClientServiceServer) SetSubscriptionInterval(context.Context, *SetSubscriptionIntervalRequest) (*SetSubscriptionIntervalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetSubscriptionInterval not implemented")
}
func (UnimplementedClientServiceServer) SetAutoSubscriptionUpdate(context.Context, *SetAutoSubscriptionUpdateRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetAutoSubscriptionUpdate not implemented")
}
func (UnimplementedClientServiceServer) RttTest(context.Context, *RttTestRequest) (*RttTestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RttTest not implemented")
}
func (UnimplementedClientServiceServer) mustEmbedUnimplementedClientServiceServer() {}
func (UnimplementedClientServiceServer) testEmbeddedByValue()                       {}

// UnsafeClientServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientServiceServer will
// result in compilation errors.
type UnsafeClientServiceServer interface {
	mustEmbedUnimplementedClientServiceServer()
}

func RegisterClientServiceServer(s grpc.ServiceRegistrar, srv ClientServiceServer) {
	// If the following call pancis, it indicates UnimplementedClientServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ClientService_ServiceDesc, srv)
}

func _ClientService_Communicate_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CommunicateRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ClientServiceServer).Communicate(m, &grpc.GenericServerStream[CommunicateRequest, CommunicateMessage]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ClientService_CommunicateServer = grpc.ServerStreamingServer[CommunicateMessage]

func _ClientService_AddInbound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddInboundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).AddInbound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_AddInbound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).AddInbound(ctx, req.(*AddInboundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_RemoveInbound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveInboundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).RemoveInbound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_RemoveInbound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).RemoveInbound(ctx, req.(*RemoveInboundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_GetStatsStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetStatsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ClientServiceServer).GetStatsStream(m, &grpc.GenericServerStream[GetStatsRequest, StatsResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ClientService_GetStatsStreamServer = grpc.ServerStreamingServer[StatsResponse]

func _ClientService_SetOutboundHandlerSpeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetOutboundHandlerSpeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).SetOutboundHandlerSpeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_SetOutboundHandlerSpeed_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).SetOutboundHandlerSpeed(ctx, req.(*SetOutboundHandlerSpeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_UserLogStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(UserLogStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ClientServiceServer).UserLogStream(m, &grpc.GenericServerStream[UserLogStreamRequest, userlogger.UserLogMessage]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ClientService_UserLogStreamServer = grpc.ServerStreamingServer[userlogger.UserLogMessage]

func _ClientService_ToggleUserLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ToggleUserLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).ToggleUserLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_ToggleUserLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).ToggleUserLog(ctx, req.(*ToggleUserLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_ToggleLogAppId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ToggleLogAppIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).ToggleLogAppId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_ToggleLogAppId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).ToggleLogAppId(ctx, req.(*ToggleLogAppIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_ChangeOutbound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangeOutboundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).ChangeOutbound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_ChangeOutbound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).ChangeOutbound(ctx, req.(*ChangeOutboundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_CurrentOutbound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CurrentOutboundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).CurrentOutbound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_CurrentOutbound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).CurrentOutbound(ctx, req.(*CurrentOutboundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_ChangeRoutingMode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangeRoutingModeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).ChangeRoutingMode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_ChangeRoutingMode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).ChangeRoutingMode(ctx, req.(*ChangeRoutingModeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_ChangeSelector_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangeSelectorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).ChangeSelector(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_ChangeSelector_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).ChangeSelector(ctx, req.(*ChangeSelectorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_UpdateSelectorBalancer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSelectorBalancerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).UpdateSelectorBalancer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_UpdateSelectorBalancer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).UpdateSelectorBalancer(ctx, req.(*UpdateSelectorBalancerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_UpdateSelectorFilter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSelectorFilterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).UpdateSelectorFilter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_UpdateSelectorFilter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).UpdateSelectorFilter(ctx, req.(*UpdateSelectorFilterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_NotifyHandlerChange_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandlerChangeNotify)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).NotifyHandlerChange(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_NotifyHandlerChange_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).NotifyHandlerChange(ctx, req.(*HandlerChangeNotify))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_SwitchFakeDns_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SwitchFakeDnsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).SwitchFakeDns(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_SwitchFakeDns_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).SwitchFakeDns(ctx, req.(*SwitchFakeDnsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_UpdateGeo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateGeoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).UpdateGeo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_UpdateGeo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).UpdateGeo(ctx, req.(*UpdateGeoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_AddGeoDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddGeoDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).AddGeoDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_AddGeoDomain_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).AddGeoDomain(ctx, req.(*AddGeoDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_RemoveGeoDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveGeoDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).RemoveGeoDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_RemoveGeoDomain_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).RemoveGeoDomain(ctx, req.(*RemoveGeoDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_ReplaceGeoDomains_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplaceDomainSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).ReplaceGeoDomains(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_ReplaceGeoDomains_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).ReplaceGeoDomains(ctx, req.(*ReplaceDomainSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_ReplaceGeoIPs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplaceIPSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).ReplaceGeoIPs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_ReplaceGeoIPs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).ReplaceGeoIPs(ctx, req.(*ReplaceIPSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_UpdateRouter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRouterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).UpdateRouter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_UpdateRouter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).UpdateRouter(ctx, req.(*UpdateRouterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_SetSubscriptionInterval_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetSubscriptionIntervalRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).SetSubscriptionInterval(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_SetSubscriptionInterval_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).SetSubscriptionInterval(ctx, req.(*SetSubscriptionIntervalRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_SetAutoSubscriptionUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetAutoSubscriptionUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).SetAutoSubscriptionUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_SetAutoSubscriptionUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).SetAutoSubscriptionUpdate(ctx, req.(*SetAutoSubscriptionUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_RttTest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RttTestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).RttTest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientService_RttTest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).RttTest(ctx, req.(*RttTestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ClientService_ServiceDesc is the grpc.ServiceDesc for ClientService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClientService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "x.clientgrpc.ClientService",
	HandlerType: (*ClientServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddInbound",
			Handler:    _ClientService_AddInbound_Handler,
		},
		{
			MethodName: "RemoveInbound",
			Handler:    _ClientService_RemoveInbound_Handler,
		},
		{
			MethodName: "SetOutboundHandlerSpeed",
			Handler:    _ClientService_SetOutboundHandlerSpeed_Handler,
		},
		{
			MethodName: "ToggleUserLog",
			Handler:    _ClientService_ToggleUserLog_Handler,
		},
		{
			MethodName: "ToggleLogAppId",
			Handler:    _ClientService_ToggleLogAppId_Handler,
		},
		{
			MethodName: "ChangeOutbound",
			Handler:    _ClientService_ChangeOutbound_Handler,
		},
		{
			MethodName: "CurrentOutbound",
			Handler:    _ClientService_CurrentOutbound_Handler,
		},
		{
			MethodName: "ChangeRoutingMode",
			Handler:    _ClientService_ChangeRoutingMode_Handler,
		},
		{
			MethodName: "ChangeSelector",
			Handler:    _ClientService_ChangeSelector_Handler,
		},
		{
			MethodName: "UpdateSelectorBalancer",
			Handler:    _ClientService_UpdateSelectorBalancer_Handler,
		},
		{
			MethodName: "UpdateSelectorFilter",
			Handler:    _ClientService_UpdateSelectorFilter_Handler,
		},
		{
			MethodName: "NotifyHandlerChange",
			Handler:    _ClientService_NotifyHandlerChange_Handler,
		},
		{
			MethodName: "SwitchFakeDns",
			Handler:    _ClientService_SwitchFakeDns_Handler,
		},
		{
			MethodName: "UpdateGeo",
			Handler:    _ClientService_UpdateGeo_Handler,
		},
		{
			MethodName: "AddGeoDomain",
			Handler:    _ClientService_AddGeoDomain_Handler,
		},
		{
			MethodName: "RemoveGeoDomain",
			Handler:    _ClientService_RemoveGeoDomain_Handler,
		},
		{
			MethodName: "ReplaceGeoDomains",
			Handler:    _ClientService_ReplaceGeoDomains_Handler,
		},
		{
			MethodName: "ReplaceGeoIPs",
			Handler:    _ClientService_ReplaceGeoIPs_Handler,
		},
		{
			MethodName: "UpdateRouter",
			Handler:    _ClientService_UpdateRouter_Handler,
		},
		{
			MethodName: "SetSubscriptionInterval",
			Handler:    _ClientService_SetSubscriptionInterval_Handler,
		},
		{
			MethodName: "SetAutoSubscriptionUpdate",
			Handler:    _ClientService_SetAutoSubscriptionUpdate_Handler,
		},
		{
			MethodName: "RttTest",
			Handler:    _ClientService_RttTest_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Communicate",
			Handler:       _ClientService_Communicate_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetStatsStream",
			Handler:       _ClientService_GetStatsStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "UserLogStream",
			Handler:       _ClientService_UserLogStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "app/clientgrpc/grpc.proto",
}
