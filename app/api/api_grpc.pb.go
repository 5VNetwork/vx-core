package api

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
	Api_UpdateTmStatus_FullMethodName                = "/x.api.Api/UpdateTmStatus"
	Api_Download_FullMethodName                      = "/x.api.Api/Download"
	Api_HandlerUsable_FullMethodName                 = "/x.api.Api/HandlerUsable"
	Api_SpeedTest_FullMethodName                     = "/x.api.Api/SpeedTest"
	Api_RttTest_FullMethodName                       = "/x.api.Api/RttTest"
	Api_GeoIP_FullMethodName                         = "/x.api.Api/GeoIP"
	Api_GetServerPublicKey_FullMethodName            = "/x.api.Api/GetServerPublicKey"
	Api_MonitorServer_FullMethodName                 = "/x.api.Api/MonitorServer"
	Api_ServerAction_FullMethodName                  = "/x.api.Api/ServerAction"
	Api_VproxyStatus_FullMethodName                  = "/x.api.Api/VproxyStatus"
	Api_VX_FullMethodName                            = "/x.api.Api/VX"
	Api_ServerConfig_FullMethodName                  = "/x.api.Api/ServerConfig"
	Api_UpdateServerConfig_FullMethodName            = "/x.api.Api/UpdateServerConfig"
	Api_UpdateSubscription_FullMethodName            = "/x.api.Api/UpdateSubscription"
	Api_ProcessGeoFiles_FullMethodName               = "/x.api.Api/ProcessGeoFiles"
	Api_Decode_FullMethodName                        = "/x.api.Api/Decode"
	Api_Deploy_FullMethodName                        = "/x.api.Api/Deploy"
	Api_GenerateCert_FullMethodName                  = "/x.api.Api/GenerateCert"
	Api_GetCertDomain_FullMethodName                 = "/x.api.Api/GetCertDomain"
	Api_AddInbound_FullMethodName                    = "/x.api.Api/AddInbound"
	Api_UploadLog_FullMethodName                     = "/x.api.Api/UploadLog"
	Api_DefaultNICHasGlobalV6_FullMethodName         = "/x.api.Api/DefaultNICHasGlobalV6"
	Api_ParseClashRuleFile_FullMethodName            = "/x.api.Api/ParseClashRuleFile"
	Api_ParseGeositeConfig_FullMethodName            = "/x.api.Api/ParseGeositeConfig"
	Api_ParseGeoIPConfig_FullMethodName              = "/x.api.Api/ParseGeoIPConfig"
	Api_RunRealiScanner_FullMethodName               = "/x.api.Api/RunRealiScanner"
	Api_GenerateX25519KeyPair_FullMethodName         = "/x.api.Api/GenerateX25519KeyPair"
	Api_StartMacSystemProxy_FullMethodName           = "/x.api.Api/StartMacSystemProxy"
	Api_StopMacSystemProxy_FullMethodName            = "/x.api.Api/StopMacSystemProxy"
	Api_CloseDb_FullMethodName                       = "/x.api.Api/CloseDb"
	Api_OpenDb_FullMethodName                        = "/x.api.Api/OpenDb"
	Api_InboundConfigToOutboundConfig_FullMethodName = "/x.api.Api/InboundConfigToOutboundConfig"
	Api_ToUrl_FullMethodName                         = "/x.api.Api/ToUrl"
)

// ApiClient is the client API for Api service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ApiClient interface {
	UpdateTmStatus(ctx context.Context, in *UpdateTmStatusRequest, opts ...grpc.CallOption) (*Receipt, error)
	// rpc SetTunName(SetTunNameRequest) returns (SetTunNameResponse);
	Download(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (*DownloadResponse, error)
	// rpc HandlerIp(HandlerIpRequest) returns (HandlerIpResponse);
	HandlerUsable(ctx context.Context, in *HandlerUsableRequest, opts ...grpc.CallOption) (*HandlerUsableResponse, error)
	SpeedTest(ctx context.Context, in *SpeedTestRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[SpeedTestResponse], error)
	RttTest(ctx context.Context, in *RttTestRequest, opts ...grpc.CallOption) (*RttTestResponse, error)
	GeoIP(ctx context.Context, in *GeoIPRequest, opts ...grpc.CallOption) (*GeoIPResponse, error)
	GetServerPublicKey(ctx context.Context, in *GetServerPublicKeyRequest, opts ...grpc.CallOption) (*GetServerPublicKeyResponse, error)
	MonitorServer(ctx context.Context, in *MonitorServerRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[MonitorServerResponse], error)
	ServerAction(ctx context.Context, in *ServerActionRequest, opts ...grpc.CallOption) (*ServerActionResponse, error)
	VproxyStatus(ctx context.Context, in *VproxyStatusRequest, opts ...grpc.CallOption) (*VproxyStatusResponse, error)
	VX(ctx context.Context, in *VXRequest, opts ...grpc.CallOption) (*Receipt, error)
	ServerConfig(ctx context.Context, in *ServerConfigRequest, opts ...grpc.CallOption) (*ServerConfigResponse, error)
	UpdateServerConfig(ctx context.Context, in *UpdateServerConfigRequest, opts ...grpc.CallOption) (*UpdateServerConfigResponse, error)
	// rpc XStatusChangeNotify(XStatusChangeNotifyRequest) returns (XStatusChangeNotifyResponse);
	// rpc SetSubscriptionInterval(SetSubscriptionIntervalRequest) returns (SetSubscriptionIntervalResponse);
	UpdateSubscription(ctx context.Context, in *UpdateSubscriptionRequest, opts ...grpc.CallOption) (*UpdateSubscriptionResponse, error)
	ProcessGeoFiles(ctx context.Context, in *ProcessGeoFilesRequest, opts ...grpc.CallOption) (*ProcessGeoFilesResponse, error)
	Decode(ctx context.Context, in *DecodeRequest, opts ...grpc.CallOption) (*DecodeResponse, error)
	Deploy(ctx context.Context, in *DeployRequest, opts ...grpc.CallOption) (*DeployResponse, error)
	GenerateCert(ctx context.Context, in *GenerateCertRequest, opts ...grpc.CallOption) (*GenerateCertResponse, error)
	GetCertDomain(ctx context.Context, in *GetCertDomainRequest, opts ...grpc.CallOption) (*GetCertDomainResponse, error)
	AddInbound(ctx context.Context, in *AddInboundRequest, opts ...grpc.CallOption) (*AddInboundResponse, error)
	UploadLog(ctx context.Context, in *UploadLogRequest, opts ...grpc.CallOption) (*UploadLogResponse, error)
	DefaultNICHasGlobalV6(ctx context.Context, in *DefaultNICHasGlobalV6Request, opts ...grpc.CallOption) (*DefaultNICHasGlobalV6Response, error)
	ParseClashRuleFile(ctx context.Context, in *ParseClashRuleFileRequest, opts ...grpc.CallOption) (*ParseClashRuleFileResponse, error)
	ParseGeositeConfig(ctx context.Context, in *ParseGeositeConfigRequest, opts ...grpc.CallOption) (*ParseGeositeConfigResponse, error)
	ParseGeoIPConfig(ctx context.Context, in *ParseGeoIPConfigRequest, opts ...grpc.CallOption) (*ParseGeoIPConfigResponse, error)
	RunRealiScanner(ctx context.Context, in *RunRealiScannerRequest, opts ...grpc.CallOption) (*RunRealiScannerResponse, error)
	GenerateX25519KeyPair(ctx context.Context, in *GenerateX25519KeyPairRequest, opts ...grpc.CallOption) (*GenerateX25519KeyPairResponse, error)
	StartMacSystemProxy(ctx context.Context, in *StartMacSystemProxyRequest, opts ...grpc.CallOption) (*Receipt, error)
	StopMacSystemProxy(ctx context.Context, in *StopMacSystemProxyRequest, opts ...grpc.CallOption) (*Receipt, error)
	CloseDb(ctx context.Context, in *CloseDbRequest, opts ...grpc.CallOption) (*Receipt, error)
	OpenDb(ctx context.Context, in *OpenDbRequest, opts ...grpc.CallOption) (*Receipt, error)
	InboundConfigToOutboundConfig(ctx context.Context, in *InboundConfigToOutboundConfigRequest, opts ...grpc.CallOption) (*InboundConfigToOutboundConfigResponse, error)
	ToUrl(ctx context.Context, in *ToUrlRequest, opts ...grpc.CallOption) (*ToUrlResponse, error)
}

type apiClient struct {
	cc grpc.ClientConnInterface
}

func NewApiClient(cc grpc.ClientConnInterface) ApiClient {
	return &apiClient{cc}
}

func (c *apiClient) UpdateTmStatus(ctx context.Context, in *UpdateTmStatusRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, Api_UpdateTmStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) Download(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (*DownloadResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DownloadResponse)
	err := c.cc.Invoke(ctx, Api_Download_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) HandlerUsable(ctx context.Context, in *HandlerUsableRequest, opts ...grpc.CallOption) (*HandlerUsableResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HandlerUsableResponse)
	err := c.cc.Invoke(ctx, Api_HandlerUsable_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) SpeedTest(ctx context.Context, in *SpeedTestRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[SpeedTestResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Api_ServiceDesc.Streams[0], Api_SpeedTest_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[SpeedTestRequest, SpeedTestResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Api_SpeedTestClient = grpc.ServerStreamingClient[SpeedTestResponse]

func (c *apiClient) RttTest(ctx context.Context, in *RttTestRequest, opts ...grpc.CallOption) (*RttTestResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RttTestResponse)
	err := c.cc.Invoke(ctx, Api_RttTest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) GeoIP(ctx context.Context, in *GeoIPRequest, opts ...grpc.CallOption) (*GeoIPResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GeoIPResponse)
	err := c.cc.Invoke(ctx, Api_GeoIP_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) GetServerPublicKey(ctx context.Context, in *GetServerPublicKeyRequest, opts ...grpc.CallOption) (*GetServerPublicKeyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetServerPublicKeyResponse)
	err := c.cc.Invoke(ctx, Api_GetServerPublicKey_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) MonitorServer(ctx context.Context, in *MonitorServerRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[MonitorServerResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Api_ServiceDesc.Streams[1], Api_MonitorServer_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[MonitorServerRequest, MonitorServerResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Api_MonitorServerClient = grpc.ServerStreamingClient[MonitorServerResponse]

func (c *apiClient) ServerAction(ctx context.Context, in *ServerActionRequest, opts ...grpc.CallOption) (*ServerActionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ServerActionResponse)
	err := c.cc.Invoke(ctx, Api_ServerAction_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) VproxyStatus(ctx context.Context, in *VproxyStatusRequest, opts ...grpc.CallOption) (*VproxyStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VproxyStatusResponse)
	err := c.cc.Invoke(ctx, Api_VproxyStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) VX(ctx context.Context, in *VXRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, Api_VX_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) ServerConfig(ctx context.Context, in *ServerConfigRequest, opts ...grpc.CallOption) (*ServerConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ServerConfigResponse)
	err := c.cc.Invoke(ctx, Api_ServerConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) UpdateServerConfig(ctx context.Context, in *UpdateServerConfigRequest, opts ...grpc.CallOption) (*UpdateServerConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateServerConfigResponse)
	err := c.cc.Invoke(ctx, Api_UpdateServerConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) UpdateSubscription(ctx context.Context, in *UpdateSubscriptionRequest, opts ...grpc.CallOption) (*UpdateSubscriptionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateSubscriptionResponse)
	err := c.cc.Invoke(ctx, Api_UpdateSubscription_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) ProcessGeoFiles(ctx context.Context, in *ProcessGeoFilesRequest, opts ...grpc.CallOption) (*ProcessGeoFilesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ProcessGeoFilesResponse)
	err := c.cc.Invoke(ctx, Api_ProcessGeoFiles_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) Decode(ctx context.Context, in *DecodeRequest, opts ...grpc.CallOption) (*DecodeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DecodeResponse)
	err := c.cc.Invoke(ctx, Api_Decode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) Deploy(ctx context.Context, in *DeployRequest, opts ...grpc.CallOption) (*DeployResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeployResponse)
	err := c.cc.Invoke(ctx, Api_Deploy_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) GenerateCert(ctx context.Context, in *GenerateCertRequest, opts ...grpc.CallOption) (*GenerateCertResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GenerateCertResponse)
	err := c.cc.Invoke(ctx, Api_GenerateCert_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) GetCertDomain(ctx context.Context, in *GetCertDomainRequest, opts ...grpc.CallOption) (*GetCertDomainResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetCertDomainResponse)
	err := c.cc.Invoke(ctx, Api_GetCertDomain_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) AddInbound(ctx context.Context, in *AddInboundRequest, opts ...grpc.CallOption) (*AddInboundResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddInboundResponse)
	err := c.cc.Invoke(ctx, Api_AddInbound_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) UploadLog(ctx context.Context, in *UploadLogRequest, opts ...grpc.CallOption) (*UploadLogResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UploadLogResponse)
	err := c.cc.Invoke(ctx, Api_UploadLog_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) DefaultNICHasGlobalV6(ctx context.Context, in *DefaultNICHasGlobalV6Request, opts ...grpc.CallOption) (*DefaultNICHasGlobalV6Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DefaultNICHasGlobalV6Response)
	err := c.cc.Invoke(ctx, Api_DefaultNICHasGlobalV6_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) ParseClashRuleFile(ctx context.Context, in *ParseClashRuleFileRequest, opts ...grpc.CallOption) (*ParseClashRuleFileResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ParseClashRuleFileResponse)
	err := c.cc.Invoke(ctx, Api_ParseClashRuleFile_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) ParseGeositeConfig(ctx context.Context, in *ParseGeositeConfigRequest, opts ...grpc.CallOption) (*ParseGeositeConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ParseGeositeConfigResponse)
	err := c.cc.Invoke(ctx, Api_ParseGeositeConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) ParseGeoIPConfig(ctx context.Context, in *ParseGeoIPConfigRequest, opts ...grpc.CallOption) (*ParseGeoIPConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ParseGeoIPConfigResponse)
	err := c.cc.Invoke(ctx, Api_ParseGeoIPConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) RunRealiScanner(ctx context.Context, in *RunRealiScannerRequest, opts ...grpc.CallOption) (*RunRealiScannerResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RunRealiScannerResponse)
	err := c.cc.Invoke(ctx, Api_RunRealiScanner_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) GenerateX25519KeyPair(ctx context.Context, in *GenerateX25519KeyPairRequest, opts ...grpc.CallOption) (*GenerateX25519KeyPairResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GenerateX25519KeyPairResponse)
	err := c.cc.Invoke(ctx, Api_GenerateX25519KeyPair_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) StartMacSystemProxy(ctx context.Context, in *StartMacSystemProxyRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, Api_StartMacSystemProxy_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) StopMacSystemProxy(ctx context.Context, in *StopMacSystemProxyRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, Api_StopMacSystemProxy_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) CloseDb(ctx context.Context, in *CloseDbRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, Api_CloseDb_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) OpenDb(ctx context.Context, in *OpenDbRequest, opts ...grpc.CallOption) (*Receipt, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Receipt)
	err := c.cc.Invoke(ctx, Api_OpenDb_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) InboundConfigToOutboundConfig(ctx context.Context, in *InboundConfigToOutboundConfigRequest, opts ...grpc.CallOption) (*InboundConfigToOutboundConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InboundConfigToOutboundConfigResponse)
	err := c.cc.Invoke(ctx, Api_InboundConfigToOutboundConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) ToUrl(ctx context.Context, in *ToUrlRequest, opts ...grpc.CallOption) (*ToUrlResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ToUrlResponse)
	err := c.cc.Invoke(ctx, Api_ToUrl_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ApiServer is the server API for Api service.
// All implementations must embed UnimplementedApiServer
// for forward compatibility.
type ApiServer interface {
	UpdateTmStatus(context.Context, *UpdateTmStatusRequest) (*Receipt, error)
	// rpc SetTunName(SetTunNameRequest) returns (SetTunNameResponse);
	Download(context.Context, *DownloadRequest) (*DownloadResponse, error)
	// rpc HandlerIp(HandlerIpRequest) returns (HandlerIpResponse);
	HandlerUsable(context.Context, *HandlerUsableRequest) (*HandlerUsableResponse, error)
	SpeedTest(*SpeedTestRequest, grpc.ServerStreamingServer[SpeedTestResponse]) error
	RttTest(context.Context, *RttTestRequest) (*RttTestResponse, error)
	GeoIP(context.Context, *GeoIPRequest) (*GeoIPResponse, error)
	GetServerPublicKey(context.Context, *GetServerPublicKeyRequest) (*GetServerPublicKeyResponse, error)
	MonitorServer(*MonitorServerRequest, grpc.ServerStreamingServer[MonitorServerResponse]) error
	ServerAction(context.Context, *ServerActionRequest) (*ServerActionResponse, error)
	VproxyStatus(context.Context, *VproxyStatusRequest) (*VproxyStatusResponse, error)
	VX(context.Context, *VXRequest) (*Receipt, error)
	ServerConfig(context.Context, *ServerConfigRequest) (*ServerConfigResponse, error)
	UpdateServerConfig(context.Context, *UpdateServerConfigRequest) (*UpdateServerConfigResponse, error)
	// rpc XStatusChangeNotify(XStatusChangeNotifyRequest) returns (XStatusChangeNotifyResponse);
	// rpc SetSubscriptionInterval(SetSubscriptionIntervalRequest) returns (SetSubscriptionIntervalResponse);
	UpdateSubscription(context.Context, *UpdateSubscriptionRequest) (*UpdateSubscriptionResponse, error)
	ProcessGeoFiles(context.Context, *ProcessGeoFilesRequest) (*ProcessGeoFilesResponse, error)
	Decode(context.Context, *DecodeRequest) (*DecodeResponse, error)
	Deploy(context.Context, *DeployRequest) (*DeployResponse, error)
	GenerateCert(context.Context, *GenerateCertRequest) (*GenerateCertResponse, error)
	GetCertDomain(context.Context, *GetCertDomainRequest) (*GetCertDomainResponse, error)
	AddInbound(context.Context, *AddInboundRequest) (*AddInboundResponse, error)
	UploadLog(context.Context, *UploadLogRequest) (*UploadLogResponse, error)
	DefaultNICHasGlobalV6(context.Context, *DefaultNICHasGlobalV6Request) (*DefaultNICHasGlobalV6Response, error)
	ParseClashRuleFile(context.Context, *ParseClashRuleFileRequest) (*ParseClashRuleFileResponse, error)
	ParseGeositeConfig(context.Context, *ParseGeositeConfigRequest) (*ParseGeositeConfigResponse, error)
	ParseGeoIPConfig(context.Context, *ParseGeoIPConfigRequest) (*ParseGeoIPConfigResponse, error)
	RunRealiScanner(context.Context, *RunRealiScannerRequest) (*RunRealiScannerResponse, error)
	GenerateX25519KeyPair(context.Context, *GenerateX25519KeyPairRequest) (*GenerateX25519KeyPairResponse, error)
	StartMacSystemProxy(context.Context, *StartMacSystemProxyRequest) (*Receipt, error)
	StopMacSystemProxy(context.Context, *StopMacSystemProxyRequest) (*Receipt, error)
	CloseDb(context.Context, *CloseDbRequest) (*Receipt, error)
	OpenDb(context.Context, *OpenDbRequest) (*Receipt, error)
	InboundConfigToOutboundConfig(context.Context, *InboundConfigToOutboundConfigRequest) (*InboundConfigToOutboundConfigResponse, error)
	ToUrl(context.Context, *ToUrlRequest) (*ToUrlResponse, error)
	mustEmbedUnimplementedApiServer()
}

// UnimplementedApiServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedApiServer struct{}

func (UnimplementedApiServer) UpdateTmStatus(context.Context, *UpdateTmStatusRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTmStatus not implemented")
}
func (UnimplementedApiServer) Download(context.Context, *DownloadRequest) (*DownloadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Download not implemented")
}
func (UnimplementedApiServer) HandlerUsable(context.Context, *HandlerUsableRequest) (*HandlerUsableResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandlerUsable not implemented")
}
func (UnimplementedApiServer) SpeedTest(*SpeedTestRequest, grpc.ServerStreamingServer[SpeedTestResponse]) error {
	return status.Errorf(codes.Unimplemented, "method SpeedTest not implemented")
}
func (UnimplementedApiServer) RttTest(context.Context, *RttTestRequest) (*RttTestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RttTest not implemented")
}
func (UnimplementedApiServer) GeoIP(context.Context, *GeoIPRequest) (*GeoIPResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GeoIP not implemented")
}
func (UnimplementedApiServer) GetServerPublicKey(context.Context, *GetServerPublicKeyRequest) (*GetServerPublicKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetServerPublicKey not implemented")
}
func (UnimplementedApiServer) MonitorServer(*MonitorServerRequest, grpc.ServerStreamingServer[MonitorServerResponse]) error {
	return status.Errorf(codes.Unimplemented, "method MonitorServer not implemented")
}
func (UnimplementedApiServer) ServerAction(context.Context, *ServerActionRequest) (*ServerActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ServerAction not implemented")
}
func (UnimplementedApiServer) VproxyStatus(context.Context, *VproxyStatusRequest) (*VproxyStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VproxyStatus not implemented")
}
func (UnimplementedApiServer) VX(context.Context, *VXRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VX not implemented")
}
func (UnimplementedApiServer) ServerConfig(context.Context, *ServerConfigRequest) (*ServerConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ServerConfig not implemented")
}
func (UnimplementedApiServer) UpdateServerConfig(context.Context, *UpdateServerConfigRequest) (*UpdateServerConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateServerConfig not implemented")
}
func (UnimplementedApiServer) UpdateSubscription(context.Context, *UpdateSubscriptionRequest) (*UpdateSubscriptionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSubscription not implemented")
}
func (UnimplementedApiServer) ProcessGeoFiles(context.Context, *ProcessGeoFilesRequest) (*ProcessGeoFilesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessGeoFiles not implemented")
}
func (UnimplementedApiServer) Decode(context.Context, *DecodeRequest) (*DecodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Decode not implemented")
}
func (UnimplementedApiServer) Deploy(context.Context, *DeployRequest) (*DeployResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Deploy not implemented")
}
func (UnimplementedApiServer) GenerateCert(context.Context, *GenerateCertRequest) (*GenerateCertResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateCert not implemented")
}
func (UnimplementedApiServer) GetCertDomain(context.Context, *GetCertDomainRequest) (*GetCertDomainResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCertDomain not implemented")
}
func (UnimplementedApiServer) AddInbound(context.Context, *AddInboundRequest) (*AddInboundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddInbound not implemented")
}
func (UnimplementedApiServer) UploadLog(context.Context, *UploadLogRequest) (*UploadLogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UploadLog not implemented")
}
func (UnimplementedApiServer) DefaultNICHasGlobalV6(context.Context, *DefaultNICHasGlobalV6Request) (*DefaultNICHasGlobalV6Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DefaultNICHasGlobalV6 not implemented")
}
func (UnimplementedApiServer) ParseClashRuleFile(context.Context, *ParseClashRuleFileRequest) (*ParseClashRuleFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ParseClashRuleFile not implemented")
}
func (UnimplementedApiServer) ParseGeositeConfig(context.Context, *ParseGeositeConfigRequest) (*ParseGeositeConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ParseGeositeConfig not implemented")
}
func (UnimplementedApiServer) ParseGeoIPConfig(context.Context, *ParseGeoIPConfigRequest) (*ParseGeoIPConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ParseGeoIPConfig not implemented")
}
func (UnimplementedApiServer) RunRealiScanner(context.Context, *RunRealiScannerRequest) (*RunRealiScannerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RunRealiScanner not implemented")
}
func (UnimplementedApiServer) GenerateX25519KeyPair(context.Context, *GenerateX25519KeyPairRequest) (*GenerateX25519KeyPairResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateX25519KeyPair not implemented")
}
func (UnimplementedApiServer) StartMacSystemProxy(context.Context, *StartMacSystemProxyRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartMacSystemProxy not implemented")
}
func (UnimplementedApiServer) StopMacSystemProxy(context.Context, *StopMacSystemProxyRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopMacSystemProxy not implemented")
}
func (UnimplementedApiServer) CloseDb(context.Context, *CloseDbRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseDb not implemented")
}
func (UnimplementedApiServer) OpenDb(context.Context, *OpenDbRequest) (*Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OpenDb not implemented")
}
func (UnimplementedApiServer) InboundConfigToOutboundConfig(context.Context, *InboundConfigToOutboundConfigRequest) (*InboundConfigToOutboundConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InboundConfigToOutboundConfig not implemented")
}
func (UnimplementedApiServer) ToUrl(context.Context, *ToUrlRequest) (*ToUrlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToUrl not implemented")
}
func (UnimplementedApiServer) mustEmbedUnimplementedApiServer() {}
func (UnimplementedApiServer) testEmbeddedByValue()             {}

// UnsafeApiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ApiServer will
// result in compilation errors.
type UnsafeApiServer interface {
	mustEmbedUnimplementedApiServer()
}

func RegisterApiServer(s grpc.ServiceRegistrar, srv ApiServer) {
	// If the following call pancis, it indicates UnimplementedApiServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Api_ServiceDesc, srv)
}

func _Api_UpdateTmStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateTmStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).UpdateTmStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_UpdateTmStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).UpdateTmStatus(ctx, req.(*UpdateTmStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_Download_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).Download(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_Download_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).Download(ctx, req.(*DownloadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_HandlerUsable_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandlerUsableRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).HandlerUsable(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_HandlerUsable_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).HandlerUsable(ctx, req.(*HandlerUsableRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_SpeedTest_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SpeedTestRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ApiServer).SpeedTest(m, &grpc.GenericServerStream[SpeedTestRequest, SpeedTestResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Api_SpeedTestServer = grpc.ServerStreamingServer[SpeedTestResponse]

func _Api_RttTest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RttTestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).RttTest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_RttTest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).RttTest(ctx, req.(*RttTestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_GeoIP_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GeoIPRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).GeoIP(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_GeoIP_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).GeoIP(ctx, req.(*GeoIPRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_GetServerPublicKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetServerPublicKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).GetServerPublicKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_GetServerPublicKey_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).GetServerPublicKey(ctx, req.(*GetServerPublicKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_MonitorServer_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(MonitorServerRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ApiServer).MonitorServer(m, &grpc.GenericServerStream[MonitorServerRequest, MonitorServerResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Api_MonitorServerServer = grpc.ServerStreamingServer[MonitorServerResponse]

func _Api_ServerAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServerActionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).ServerAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_ServerAction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).ServerAction(ctx, req.(*ServerActionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_VproxyStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VproxyStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).VproxyStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_VproxyStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).VproxyStatus(ctx, req.(*VproxyStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_VX_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VXRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).VX(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_VX_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).VX(ctx, req.(*VXRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_ServerConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServerConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).ServerConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_ServerConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).ServerConfig(ctx, req.(*ServerConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_UpdateServerConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateServerConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).UpdateServerConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_UpdateServerConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).UpdateServerConfig(ctx, req.(*UpdateServerConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_UpdateSubscription_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSubscriptionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).UpdateSubscription(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_UpdateSubscription_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).UpdateSubscription(ctx, req.(*UpdateSubscriptionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_ProcessGeoFiles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProcessGeoFilesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).ProcessGeoFiles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_ProcessGeoFiles_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).ProcessGeoFiles(ctx, req.(*ProcessGeoFilesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_Decode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DecodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).Decode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_Decode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).Decode(ctx, req.(*DecodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_Deploy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeployRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).Deploy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_Deploy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).Deploy(ctx, req.(*DeployRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_GenerateCert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateCertRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).GenerateCert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_GenerateCert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).GenerateCert(ctx, req.(*GenerateCertRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_GetCertDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCertDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).GetCertDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_GetCertDomain_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).GetCertDomain(ctx, req.(*GetCertDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_AddInbound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddInboundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).AddInbound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_AddInbound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).AddInbound(ctx, req.(*AddInboundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_UploadLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).UploadLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_UploadLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).UploadLog(ctx, req.(*UploadLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_DefaultNICHasGlobalV6_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DefaultNICHasGlobalV6Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).DefaultNICHasGlobalV6(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_DefaultNICHasGlobalV6_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).DefaultNICHasGlobalV6(ctx, req.(*DefaultNICHasGlobalV6Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_ParseClashRuleFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ParseClashRuleFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).ParseClashRuleFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_ParseClashRuleFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).ParseClashRuleFile(ctx, req.(*ParseClashRuleFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_ParseGeositeConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ParseGeositeConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).ParseGeositeConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_ParseGeositeConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).ParseGeositeConfig(ctx, req.(*ParseGeositeConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_ParseGeoIPConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ParseGeoIPConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).ParseGeoIPConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_ParseGeoIPConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).ParseGeoIPConfig(ctx, req.(*ParseGeoIPConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_RunRealiScanner_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunRealiScannerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).RunRealiScanner(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_RunRealiScanner_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).RunRealiScanner(ctx, req.(*RunRealiScannerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_GenerateX25519KeyPair_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateX25519KeyPairRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).GenerateX25519KeyPair(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_GenerateX25519KeyPair_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).GenerateX25519KeyPair(ctx, req.(*GenerateX25519KeyPairRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_StartMacSystemProxy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartMacSystemProxyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).StartMacSystemProxy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_StartMacSystemProxy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).StartMacSystemProxy(ctx, req.(*StartMacSystemProxyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_StopMacSystemProxy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopMacSystemProxyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).StopMacSystemProxy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_StopMacSystemProxy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).StopMacSystemProxy(ctx, req.(*StopMacSystemProxyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_CloseDb_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseDbRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).CloseDb(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_CloseDb_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).CloseDb(ctx, req.(*CloseDbRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_OpenDb_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OpenDbRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).OpenDb(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_OpenDb_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).OpenDb(ctx, req.(*OpenDbRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_InboundConfigToOutboundConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InboundConfigToOutboundConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).InboundConfigToOutboundConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_InboundConfigToOutboundConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).InboundConfigToOutboundConfig(ctx, req.(*InboundConfigToOutboundConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Api_ToUrl_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ToUrlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServer).ToUrl(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Api_ToUrl_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServer).ToUrl(ctx, req.(*ToUrlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Api_ServiceDesc is the grpc.ServiceDesc for Api service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Api_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "x.api.Api",
	HandlerType: (*ApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpdateTmStatus",
			Handler:    _Api_UpdateTmStatus_Handler,
		},
		{
			MethodName: "Download",
			Handler:    _Api_Download_Handler,
		},
		{
			MethodName: "HandlerUsable",
			Handler:    _Api_HandlerUsable_Handler,
		},
		{
			MethodName: "RttTest",
			Handler:    _Api_RttTest_Handler,
		},
		{
			MethodName: "GeoIP",
			Handler:    _Api_GeoIP_Handler,
		},
		{
			MethodName: "GetServerPublicKey",
			Handler:    _Api_GetServerPublicKey_Handler,
		},
		{
			MethodName: "ServerAction",
			Handler:    _Api_ServerAction_Handler,
		},
		{
			MethodName: "VproxyStatus",
			Handler:    _Api_VproxyStatus_Handler,
		},
		{
			MethodName: "VX",
			Handler:    _Api_VX_Handler,
		},
		{
			MethodName: "ServerConfig",
			Handler:    _Api_ServerConfig_Handler,
		},
		{
			MethodName: "UpdateServerConfig",
			Handler:    _Api_UpdateServerConfig_Handler,
		},
		{
			MethodName: "UpdateSubscription",
			Handler:    _Api_UpdateSubscription_Handler,
		},
		{
			MethodName: "ProcessGeoFiles",
			Handler:    _Api_ProcessGeoFiles_Handler,
		},
		{
			MethodName: "Decode",
			Handler:    _Api_Decode_Handler,
		},
		{
			MethodName: "Deploy",
			Handler:    _Api_Deploy_Handler,
		},
		{
			MethodName: "GenerateCert",
			Handler:    _Api_GenerateCert_Handler,
		},
		{
			MethodName: "GetCertDomain",
			Handler:    _Api_GetCertDomain_Handler,
		},
		{
			MethodName: "AddInbound",
			Handler:    _Api_AddInbound_Handler,
		},
		{
			MethodName: "UploadLog",
			Handler:    _Api_UploadLog_Handler,
		},
		{
			MethodName: "DefaultNICHasGlobalV6",
			Handler:    _Api_DefaultNICHasGlobalV6_Handler,
		},
		{
			MethodName: "ParseClashRuleFile",
			Handler:    _Api_ParseClashRuleFile_Handler,
		},
		{
			MethodName: "ParseGeositeConfig",
			Handler:    _Api_ParseGeositeConfig_Handler,
		},
		{
			MethodName: "ParseGeoIPConfig",
			Handler:    _Api_ParseGeoIPConfig_Handler,
		},
		{
			MethodName: "RunRealiScanner",
			Handler:    _Api_RunRealiScanner_Handler,
		},
		{
			MethodName: "GenerateX25519KeyPair",
			Handler:    _Api_GenerateX25519KeyPair_Handler,
		},
		{
			MethodName: "StartMacSystemProxy",
			Handler:    _Api_StartMacSystemProxy_Handler,
		},
		{
			MethodName: "StopMacSystemProxy",
			Handler:    _Api_StopMacSystemProxy_Handler,
		},
		{
			MethodName: "CloseDb",
			Handler:    _Api_CloseDb_Handler,
		},
		{
			MethodName: "OpenDb",
			Handler:    _Api_OpenDb_Handler,
		},
		{
			MethodName: "InboundConfigToOutboundConfig",
			Handler:    _Api_InboundConfigToOutboundConfig_Handler,
		},
		{
			MethodName: "ToUrl",
			Handler:    _Api_ToUrl_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SpeedTest",
			Handler:       _Api_SpeedTest_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "MonitorServer",
			Handler:       _Api_MonitorServer_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "app/api/api.proto",
}
