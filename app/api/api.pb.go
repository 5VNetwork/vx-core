package api

import (
	configs "github.com/5vnetwork/vx-core/app/configs"
	server "github.com/5vnetwork/vx-core/app/configs/server"
	geo "github.com/5vnetwork/vx-core/common/geo"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type XStatusChangeNotifyRequest_Status int32

const (
	XStatusChangeNotifyRequest_STATUS_START    XStatusChangeNotifyRequest_Status = 0
	XStatusChangeNotifyRequest_STATUS_STOP     XStatusChangeNotifyRequest_Status = 1
	XStatusChangeNotifyRequest_STATUS_STARTING XStatusChangeNotifyRequest_Status = 2
	XStatusChangeNotifyRequest_STATUS_STOPPING XStatusChangeNotifyRequest_Status = 3
)

// Enum value maps for XStatusChangeNotifyRequest_Status.
var (
	XStatusChangeNotifyRequest_Status_name = map[int32]string{
		0: "STATUS_START",
		1: "STATUS_STOP",
		2: "STATUS_STARTING",
		3: "STATUS_STOPPING",
	}
	XStatusChangeNotifyRequest_Status_value = map[string]int32{
		"STATUS_START":    0,
		"STATUS_STOP":     1,
		"STATUS_STARTING": 2,
		"STATUS_STOPPING": 3,
	}
)

func (x XStatusChangeNotifyRequest_Status) Enum() *XStatusChangeNotifyRequest_Status {
	p := new(XStatusChangeNotifyRequest_Status)
	*p = x
	return p
}

func (x XStatusChangeNotifyRequest_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (XStatusChangeNotifyRequest_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_app_api_api_proto_enumTypes[0].Descriptor()
}

func (XStatusChangeNotifyRequest_Status) Type() protoreflect.EnumType {
	return &file_app_api_api_proto_enumTypes[0]
}

func (x XStatusChangeNotifyRequest_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use XStatusChangeNotifyRequest_Status.Descriptor instead.
func (XStatusChangeNotifyRequest_Status) EnumDescriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{1, 0}
}

type ServerActionRequest_Action int32

const (
	ServerActionRequest_ACTION_SHUTDOWN ServerActionRequest_Action = 0
	ServerActionRequest_ACTION_RESTART  ServerActionRequest_Action = 1 // ACTION_SUSPEND = 2;
)

// Enum value maps for ServerActionRequest_Action.
var (
	ServerActionRequest_Action_name = map[int32]string{
		0: "ACTION_SHUTDOWN",
		1: "ACTION_RESTART",
	}
	ServerActionRequest_Action_value = map[string]int32{
		"ACTION_SHUTDOWN": 0,
		"ACTION_RESTART":  1,
	}
)

func (x ServerActionRequest_Action) Enum() *ServerActionRequest_Action {
	p := new(ServerActionRequest_Action)
	*p = x
	return p
}

func (x ServerActionRequest_Action) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ServerActionRequest_Action) Descriptor() protoreflect.EnumDescriptor {
	return file_app_api_api_proto_enumTypes[1].Descriptor()
}

func (ServerActionRequest_Action) Type() protoreflect.EnumType {
	return &file_app_api_api_proto_enumTypes[1]
}

func (x ServerActionRequest_Action) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ServerActionRequest_Action.Descriptor instead.
func (ServerActionRequest_Action) EnumDescriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{26, 0}
}

type ApiServerConfig struct {
	state      protoimpl.MessageState `protogen:"open.v1"`
	ListenAddr string                 `protobuf:"bytes,1,opt,name=listen_addr,json=listenAddr,proto3" json:"listen_addr,omitempty"`
	GeoipPath  string                 `protobuf:"bytes,2,opt,name=geoip_path,json=geoipPath,proto3" json:"geoip_path,omitempty"`
	TunName    string                 `protobuf:"bytes,3,opt,name=tun_name,json=tunName,proto3" json:"tun_name,omitempty"`
	LogLevel   uint32                 `protobuf:"varint,4,opt,name=log_level,json=logLevel,proto3" json:"log_level,omitempty"`
	DbPath     string                 `protobuf:"bytes,5,opt,name=db_path,json=dbPath,proto3" json:"db_path,omitempty"`
	// milliseconds since epoch
	LastUpdateTime uint32 `protobuf:"varint,6,opt,name=last_update_time,json=lastUpdateTime,proto3" json:"last_update_time,omitempty"`
	// minutes
	Interval         uint32 `protobuf:"varint,7,opt,name=interval,proto3" json:"interval,omitempty"`
	BindToDefaultNic bool   `protobuf:"varint,8,opt,name=bind_to_default_nic,json=bindToDefaultNic,proto3" json:"bind_to_default_nic,omitempty"`
	ClientCert       []byte `protobuf:"bytes,9,opt,name=client_cert,json=clientCert,proto3" json:"client_cert,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *ApiServerConfig) Reset() {
	*x = ApiServerConfig{}
	mi := &file_app_api_api_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ApiServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ApiServerConfig) ProtoMessage() {}

func (x *ApiServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ApiServerConfig.ProtoReflect.Descriptor instead.
func (*ApiServerConfig) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{0}
}

func (x *ApiServerConfig) GetListenAddr() string {
	if x != nil {
		return x.ListenAddr
	}
	return ""
}

func (x *ApiServerConfig) GetGeoipPath() string {
	if x != nil {
		return x.GeoipPath
	}
	return ""
}

func (x *ApiServerConfig) GetTunName() string {
	if x != nil {
		return x.TunName
	}
	return ""
}

func (x *ApiServerConfig) GetLogLevel() uint32 {
	if x != nil {
		return x.LogLevel
	}
	return 0
}

func (x *ApiServerConfig) GetDbPath() string {
	if x != nil {
		return x.DbPath
	}
	return ""
}

func (x *ApiServerConfig) GetLastUpdateTime() uint32 {
	if x != nil {
		return x.LastUpdateTime
	}
	return 0
}

func (x *ApiServerConfig) GetInterval() uint32 {
	if x != nil {
		return x.Interval
	}
	return 0
}

func (x *ApiServerConfig) GetBindToDefaultNic() bool {
	if x != nil {
		return x.BindToDefaultNic
	}
	return false
}

func (x *ApiServerConfig) GetClientCert() []byte {
	if x != nil {
		return x.ClientCert
	}
	return nil
}

type XStatusChangeNotifyRequest struct {
	state         protoimpl.MessageState            `protogen:"open.v1"`
	Status        XStatusChangeNotifyRequest_Status `protobuf:"varint,1,opt,name=status,proto3,enum=x.api.XStatusChangeNotifyRequest_Status" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *XStatusChangeNotifyRequest) Reset() {
	*x = XStatusChangeNotifyRequest{}
	mi := &file_app_api_api_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *XStatusChangeNotifyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*XStatusChangeNotifyRequest) ProtoMessage() {}

func (x *XStatusChangeNotifyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use XStatusChangeNotifyRequest.ProtoReflect.Descriptor instead.
func (*XStatusChangeNotifyRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{1}
}

func (x *XStatusChangeNotifyRequest) GetStatus() XStatusChangeNotifyRequest_Status {
	if x != nil {
		return x.Status
	}
	return XStatusChangeNotifyRequest_STATUS_START
}

type XStatusChangeNotifyResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *XStatusChangeNotifyResponse) Reset() {
	*x = XStatusChangeNotifyResponse{}
	mi := &file_app_api_api_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *XStatusChangeNotifyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*XStatusChangeNotifyResponse) ProtoMessage() {}

func (x *XStatusChangeNotifyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use XStatusChangeNotifyResponse.ProtoReflect.Descriptor instead.
func (*XStatusChangeNotifyResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{2}
}

type SetSubscriptionIntervalRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// minus means no auto update
	// minutes
	Interval      int32 `protobuf:"varint,1,opt,name=interval,proto3" json:"interval,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SetSubscriptionIntervalRequest) Reset() {
	*x = SetSubscriptionIntervalRequest{}
	mi := &file_app_api_api_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetSubscriptionIntervalRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetSubscriptionIntervalRequest) ProtoMessage() {}

func (x *SetSubscriptionIntervalRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetSubscriptionIntervalRequest.ProtoReflect.Descriptor instead.
func (*SetSubscriptionIntervalRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{3}
}

func (x *SetSubscriptionIntervalRequest) GetInterval() int32 {
	if x != nil {
		return x.Interval
	}
	return 0
}

type SetSubscriptionIntervalResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SetSubscriptionIntervalResponse) Reset() {
	*x = SetSubscriptionIntervalResponse{}
	mi := &file_app_api_api_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetSubscriptionIntervalResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetSubscriptionIntervalResponse) ProtoMessage() {}

func (x *SetSubscriptionIntervalResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetSubscriptionIntervalResponse.ProtoReflect.Descriptor instead.
func (*SetSubscriptionIntervalResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{4}
}

type UpdateSubscriptionRequest struct {
	state         protoimpl.MessageState   `protogen:"open.v1"`
	Id            int64                    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	All           bool                     `protobuf:"varint,2,opt,name=all,proto3" json:"all,omitempty"`
	Handlers      []*configs.HandlerConfig `protobuf:"bytes,3,rep,name=handlers,proto3" json:"handlers,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateSubscriptionRequest) Reset() {
	*x = UpdateSubscriptionRequest{}
	mi := &file_app_api_api_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateSubscriptionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateSubscriptionRequest) ProtoMessage() {}

func (x *UpdateSubscriptionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateSubscriptionRequest.ProtoReflect.Descriptor instead.
func (*UpdateSubscriptionRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{5}
}

func (x *UpdateSubscriptionRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *UpdateSubscriptionRequest) GetAll() bool {
	if x != nil {
		return x.All
	}
	return false
}

func (x *UpdateSubscriptionRequest) GetHandlers() []*configs.HandlerConfig {
	if x != nil {
		return x.Handlers
	}
	return nil
}

type UpdateSubscriptionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       int32                  `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Fail          int32                  `protobuf:"varint,2,opt,name=fail,proto3" json:"fail,omitempty"`
	SuccessNodes  int32                  `protobuf:"varint,3,opt,name=success_nodes,json=successNodes,proto3" json:"success_nodes,omitempty"`
	ErrorReasons  map[string]string      `protobuf:"bytes,5,rep,name=error_reasons,json=errorReasons,proto3" json:"error_reasons,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	FailedNodes   []string               `protobuf:"bytes,6,rep,name=failed_nodes,json=failedNodes,proto3" json:"failed_nodes,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateSubscriptionResponse) Reset() {
	*x = UpdateSubscriptionResponse{}
	mi := &file_app_api_api_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateSubscriptionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateSubscriptionResponse) ProtoMessage() {}

func (x *UpdateSubscriptionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateSubscriptionResponse.ProtoReflect.Descriptor instead.
func (*UpdateSubscriptionResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{6}
}

func (x *UpdateSubscriptionResponse) GetSuccess() int32 {
	if x != nil {
		return x.Success
	}
	return 0
}

func (x *UpdateSubscriptionResponse) GetFail() int32 {
	if x != nil {
		return x.Fail
	}
	return 0
}

func (x *UpdateSubscriptionResponse) GetSuccessNodes() int32 {
	if x != nil {
		return x.SuccessNodes
	}
	return 0
}

func (x *UpdateSubscriptionResponse) GetErrorReasons() map[string]string {
	if x != nil {
		return x.ErrorReasons
	}
	return nil
}

func (x *UpdateSubscriptionResponse) GetFailedNodes() []string {
	if x != nil {
		return x.FailedNodes
	}
	return nil
}

type SetTunNameRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TunName       string                 `protobuf:"bytes,1,opt,name=tun_name,json=tunName,proto3" json:"tun_name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SetTunNameRequest) Reset() {
	*x = SetTunNameRequest{}
	mi := &file_app_api_api_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetTunNameRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetTunNameRequest) ProtoMessage() {}

func (x *SetTunNameRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetTunNameRequest.ProtoReflect.Descriptor instead.
func (*SetTunNameRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{7}
}

func (x *SetTunNameRequest) GetTunName() string {
	if x != nil {
		return x.TunName
	}
	return ""
}

type SetTunNameResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SetTunNameResponse) Reset() {
	*x = SetTunNameResponse{}
	mi := &file_app_api_api_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetTunNameResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetTunNameResponse) ProtoMessage() {}

func (x *SetTunNameResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetTunNameResponse.ProtoReflect.Descriptor instead.
func (*SetTunNameResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{8}
}

type DownloadRequest struct {
	state    protoimpl.MessageState   `protogen:"open.v1"`
	Url      string                   `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Handlers []*configs.HandlerConfig `protobuf:"bytes,2,rep,name=handlers,proto3" json:"handlers,omitempty"`
	// if nil, download to memory
	Dest          string `protobuf:"bytes,3,opt,name=dest,proto3" json:"dest,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DownloadRequest) Reset() {
	*x = DownloadRequest{}
	mi := &file_app_api_api_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DownloadRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadRequest) ProtoMessage() {}

func (x *DownloadRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadRequest.ProtoReflect.Descriptor instead.
func (*DownloadRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{9}
}

func (x *DownloadRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *DownloadRequest) GetHandlers() []*configs.HandlerConfig {
	if x != nil {
		return x.Handlers
	}
	return nil
}

func (x *DownloadRequest) GetDest() string {
	if x != nil {
		return x.Dest
	}
	return ""
}

type DownloadResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// key is handler id, value is usage in bytes
	Usage map[string]uint32 `protobuf:"bytes,1,rep,name=usage,proto3" json:"usage,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
	// if request to download to memory, this is the data
	Data          []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DownloadResponse) Reset() {
	*x = DownloadResponse{}
	mi := &file_app_api_api_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DownloadResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadResponse) ProtoMessage() {}

func (x *DownloadResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadResponse.ProtoReflect.Descriptor instead.
func (*DownloadResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{10}
}

func (x *DownloadResponse) GetUsage() map[string]uint32 {
	if x != nil {
		return x.Usage
	}
	return nil
}

func (x *DownloadResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type HandlerIpRequest struct {
	state         protoimpl.MessageState         `protogen:"open.v1"`
	Handler       *configs.OutboundHandlerConfig `protobuf:"bytes,1,opt,name=handler,proto3" json:"handler,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerIpRequest) Reset() {
	*x = HandlerIpRequest{}
	mi := &file_app_api_api_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerIpRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerIpRequest) ProtoMessage() {}

func (x *HandlerIpRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerIpRequest.ProtoReflect.Descriptor instead.
func (*HandlerIpRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{11}
}

func (x *HandlerIpRequest) GetHandler() *configs.OutboundHandlerConfig {
	if x != nil {
		return x.Handler
	}
	return nil
}

type RttTestRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Addr          string                 `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	Port          uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RttTestRequest) Reset() {
	*x = RttTestRequest{}
	mi := &file_app_api_api_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RttTestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RttTestRequest) ProtoMessage() {}

func (x *RttTestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RttTestRequest.ProtoReflect.Descriptor instead.
func (*RttTestRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{12}
}

func (x *RttTestRequest) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *RttTestRequest) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

type RttTestResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Ping          uint32                 `protobuf:"varint,1,opt,name=ping,proto3" json:"ping,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RttTestResponse) Reset() {
	*x = RttTestResponse{}
	mi := &file_app_api_api_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RttTestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RttTestResponse) ProtoMessage() {}

func (x *RttTestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RttTestResponse.ProtoReflect.Descriptor instead.
func (*RttTestResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{13}
}

func (x *RttTestResponse) GetPing() uint32 {
	if x != nil {
		return x.Ping
	}
	return 0
}

type HandlerIpResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// ipv6 address of the server
	Ip6 string `protobuf:"bytes,1,opt,name=ip6,proto3" json:"ip6,omitempty"`
	// ipv4 address of the server
	Ip4           string `protobuf:"bytes,2,opt,name=ip4,proto3" json:"ip4,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerIpResponse) Reset() {
	*x = HandlerIpResponse{}
	mi := &file_app_api_api_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerIpResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerIpResponse) ProtoMessage() {}

func (x *HandlerIpResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerIpResponse.ProtoReflect.Descriptor instead.
func (*HandlerIpResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{14}
}

func (x *HandlerIpResponse) GetIp6() string {
	if x != nil {
		return x.Ip6
	}
	return ""
}

func (x *HandlerIpResponse) GetIp4() string {
	if x != nil {
		return x.Ip4
	}
	return ""
}

type HandlerUsableRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Handler       *configs.HandlerConfig `protobuf:"bytes,1,opt,name=handler,proto3" json:"handler,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerUsableRequest) Reset() {
	*x = HandlerUsableRequest{}
	mi := &file_app_api_api_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerUsableRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerUsableRequest) ProtoMessage() {}

func (x *HandlerUsableRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerUsableRequest.ProtoReflect.Descriptor instead.
func (*HandlerUsableRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{15}
}

func (x *HandlerUsableRequest) GetHandler() *configs.HandlerConfig {
	if x != nil {
		return x.Handler
	}
	return nil
}

type HandlerUsableResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Ping          int32                  `protobuf:"varint,1,opt,name=ping,proto3" json:"ping,omitempty"`
	Ip            string                 `protobuf:"bytes,2,opt,name=ip,proto3" json:"ip,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerUsableResponse) Reset() {
	*x = HandlerUsableResponse{}
	mi := &file_app_api_api_proto_msgTypes[16]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerUsableResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerUsableResponse) ProtoMessage() {}

func (x *HandlerUsableResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[16]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerUsableResponse.ProtoReflect.Descriptor instead.
func (*HandlerUsableResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{16}
}

func (x *HandlerUsableResponse) GetPing() int32 {
	if x != nil {
		return x.Ping
	}
	return 0
}

func (x *HandlerUsableResponse) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

// speed test
type SpeedTestRequest struct {
	state    protoimpl.MessageState   `protogen:"open.v1"`
	Handlers []*configs.HandlerConfig `protobuf:"bytes,1,rep,name=handlers,proto3" json:"handlers,omitempty"`
	// 1 or 10
	Size          uint32 `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SpeedTestRequest) Reset() {
	*x = SpeedTestRequest{}
	mi := &file_app_api_api_proto_msgTypes[17]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SpeedTestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SpeedTestRequest) ProtoMessage() {}

func (x *SpeedTestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[17]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SpeedTestRequest.ProtoReflect.Descriptor instead.
func (*SpeedTestRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{17}
}

func (x *SpeedTestRequest) GetHandlers() []*configs.HandlerConfig {
	if x != nil {
		return x.Handlers
	}
	return nil
}

func (x *SpeedTestRequest) GetSize() uint32 {
	if x != nil {
		return x.Size
	}
	return 0
}

type SpeedTestResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// rate
	// uint64 up = 2;
	// rate. If 0, it means unusable
	Down int32  `protobuf:"varint,3,opt,name=down,proto3" json:"down,omitempty"`
	Tag  string `protobuf:"bytes,4,opt,name=tag,proto3" json:"tag,omitempty"`
	// uint32 usage_up = 6;
	UsageDown     uint32 `protobuf:"varint,7,opt,name=usage_down,json=usageDown,proto3" json:"usage_down,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SpeedTestResponse) Reset() {
	*x = SpeedTestResponse{}
	mi := &file_app_api_api_proto_msgTypes[18]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SpeedTestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SpeedTestResponse) ProtoMessage() {}

func (x *SpeedTestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[18]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SpeedTestResponse.ProtoReflect.Descriptor instead.
func (*SpeedTestResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{18}
}

func (x *SpeedTestResponse) GetDown() int32 {
	if x != nil {
		return x.Down
	}
	return 0
}

func (x *SpeedTestResponse) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *SpeedTestResponse) GetUsageDown() uint32 {
	if x != nil {
		return x.UsageDown
	}
	return 0
}

type GeoIPRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Ips           []string               `protobuf:"bytes,1,rep,name=ips,proto3" json:"ips,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GeoIPRequest) Reset() {
	*x = GeoIPRequest{}
	mi := &file_app_api_api_proto_msgTypes[19]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GeoIPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoIPRequest) ProtoMessage() {}

func (x *GeoIPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[19]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoIPRequest.ProtoReflect.Descriptor instead.
func (*GeoIPRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{19}
}

func (x *GeoIPRequest) GetIps() []string {
	if x != nil {
		return x.Ips
	}
	return nil
}

type GeoIPResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// country code. length is same as ips in GeoIPRequest
	// country code is Alpha-2(ISO 3166)
	Countries     []string `protobuf:"bytes,1,rep,name=countries,proto3" json:"countries,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GeoIPResponse) Reset() {
	*x = GeoIPResponse{}
	mi := &file_app_api_api_proto_msgTypes[20]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GeoIPResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoIPResponse) ProtoMessage() {}

func (x *GeoIPResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[20]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoIPResponse.ProtoReflect.Descriptor instead.
func (*GeoIPResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{20}
}

func (x *GeoIPResponse) GetCountries() []string {
	if x != nil {
		return x.Countries
	}
	return nil
}

// server status
type ServerSshConfig struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Address          string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port             uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	Username         string                 `protobuf:"bytes,3,opt,name=username,proto3" json:"username,omitempty"`
	SudoPassword     string                 `protobuf:"bytes,4,opt,name=sudo_password,json=sudoPassword,proto3" json:"sudo_password,omitempty"`
	SshKey           []byte                 `protobuf:"bytes,6,opt,name=ssh_key,json=sshKey,proto3" json:"ssh_key,omitempty"`
	SshKeyPath       string                 `protobuf:"bytes,7,opt,name=ssh_key_path,json=sshKeyPath,proto3" json:"ssh_key_path,omitempty"`
	SshKeyPassphrase string                 `protobuf:"bytes,8,opt,name=ssh_key_passphrase,json=sshKeyPassphrase,proto3" json:"ssh_key_passphrase,omitempty"`
	ServerPubKey     []byte                 `protobuf:"bytes,9,opt,name=server_pub_key,json=serverPubKey,proto3" json:"server_pub_key,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *ServerSshConfig) Reset() {
	*x = ServerSshConfig{}
	mi := &file_app_api_api_proto_msgTypes[21]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerSshConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerSshConfig) ProtoMessage() {}

func (x *ServerSshConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[21]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerSshConfig.ProtoReflect.Descriptor instead.
func (*ServerSshConfig) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{21}
}

func (x *ServerSshConfig) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ServerSshConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ServerSshConfig) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *ServerSshConfig) GetSudoPassword() string {
	if x != nil {
		return x.SudoPassword
	}
	return ""
}

func (x *ServerSshConfig) GetSshKey() []byte {
	if x != nil {
		return x.SshKey
	}
	return nil
}

func (x *ServerSshConfig) GetSshKeyPath() string {
	if x != nil {
		return x.SshKeyPath
	}
	return ""
}

func (x *ServerSshConfig) GetSshKeyPassphrase() string {
	if x != nil {
		return x.SshKeyPassphrase
	}
	return ""
}

func (x *ServerSshConfig) GetServerPubKey() []byte {
	if x != nil {
		return x.ServerPubKey
	}
	return nil
}

type MonitorServerRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SshConfig     *ServerSshConfig       `protobuf:"bytes,1,opt,name=ssh_config,json=sshConfig,proto3" json:"ssh_config,omitempty"`
	Interval      uint32                 `protobuf:"varint,2,opt,name=interval,proto3" json:"interval,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MonitorServerRequest) Reset() {
	*x = MonitorServerRequest{}
	mi := &file_app_api_api_proto_msgTypes[22]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MonitorServerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MonitorServerRequest) ProtoMessage() {}

func (x *MonitorServerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[22]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MonitorServerRequest.ProtoReflect.Descriptor instead.
func (*MonitorServerRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{22}
}

func (x *MonitorServerRequest) GetSshConfig() *ServerSshConfig {
	if x != nil {
		return x.SshConfig
	}
	return nil
}

func (x *MonitorServerRequest) GetInterval() uint32 {
	if x != nil {
		return x.Interval
	}
	return 0
}

type MonitorServerResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Cpu           uint32                 `protobuf:"varint,1,opt,name=cpu,proto3" json:"cpu,omitempty"`
	UsedMemory    uint64                 `protobuf:"varint,2,opt,name=used_memory,json=usedMemory,proto3" json:"used_memory,omitempty"`
	TotalMemory   uint64                 `protobuf:"varint,3,opt,name=total_memory,json=totalMemory,proto3" json:"total_memory,omitempty"`
	UsedDisk      uint32                 `protobuf:"varint,4,opt,name=used_disk,json=usedDisk,proto3" json:"used_disk,omitempty"`
	TotalDisk     uint32                 `protobuf:"varint,5,opt,name=total_disk,json=totalDisk,proto3" json:"total_disk,omitempty"`
	NetInSpeed    uint32                 `protobuf:"varint,6,opt,name=net_in_speed,json=netInSpeed,proto3" json:"net_in_speed,omitempty"`
	NetOutSpeed   uint32                 `protobuf:"varint,7,opt,name=net_out_speed,json=netOutSpeed,proto3" json:"net_out_speed,omitempty"`
	NetInUsage    uint64                 `protobuf:"varint,8,opt,name=net_in_usage,json=netInUsage,proto3" json:"net_in_usage,omitempty"`
	NetOutUsage   uint64                 `protobuf:"varint,9,opt,name=net_out_usage,json=netOutUsage,proto3" json:"net_out_usage,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MonitorServerResponse) Reset() {
	*x = MonitorServerResponse{}
	mi := &file_app_api_api_proto_msgTypes[23]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MonitorServerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MonitorServerResponse) ProtoMessage() {}

func (x *MonitorServerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[23]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MonitorServerResponse.ProtoReflect.Descriptor instead.
func (*MonitorServerResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{23}
}

func (x *MonitorServerResponse) GetCpu() uint32 {
	if x != nil {
		return x.Cpu
	}
	return 0
}

func (x *MonitorServerResponse) GetUsedMemory() uint64 {
	if x != nil {
		return x.UsedMemory
	}
	return 0
}

func (x *MonitorServerResponse) GetTotalMemory() uint64 {
	if x != nil {
		return x.TotalMemory
	}
	return 0
}

func (x *MonitorServerResponse) GetUsedDisk() uint32 {
	if x != nil {
		return x.UsedDisk
	}
	return 0
}

func (x *MonitorServerResponse) GetTotalDisk() uint32 {
	if x != nil {
		return x.TotalDisk
	}
	return 0
}

func (x *MonitorServerResponse) GetNetInSpeed() uint32 {
	if x != nil {
		return x.NetInSpeed
	}
	return 0
}

func (x *MonitorServerResponse) GetNetOutSpeed() uint32 {
	if x != nil {
		return x.NetOutSpeed
	}
	return 0
}

func (x *MonitorServerResponse) GetNetInUsage() uint64 {
	if x != nil {
		return x.NetInUsage
	}
	return 0
}

func (x *MonitorServerResponse) GetNetOutUsage() uint64 {
	if x != nil {
		return x.NetOutUsage
	}
	return 0
}

type DeployRequest struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	SshConfig      *ServerSshConfig       `protobuf:"bytes,1,opt,name=ssh_config,json=sshConfig,proto3" json:"ssh_config,omitempty"`
	HysteriaConfig []byte                 `protobuf:"bytes,2,opt,name=hysteria_config,json=hysteriaConfig,proto3" json:"hysteria_config,omitempty"`
	XrayConfig     []byte                 `protobuf:"bytes,3,opt,name=xray_config,json=xrayConfig,proto3" json:"xray_config,omitempty"`
	Files          map[string][]byte      `protobuf:"bytes,4,rep,name=files,proto3" json:"files,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	VxConfig       *server.ServerConfig   `protobuf:"bytes,5,opt,name=vx_config,json=vxConfig,proto3" json:"vx_config,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *DeployRequest) Reset() {
	*x = DeployRequest{}
	mi := &file_app_api_api_proto_msgTypes[24]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeployRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeployRequest) ProtoMessage() {}

func (x *DeployRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[24]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeployRequest.ProtoReflect.Descriptor instead.
func (*DeployRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{24}
}

func (x *DeployRequest) GetSshConfig() *ServerSshConfig {
	if x != nil {
		return x.SshConfig
	}
	return nil
}

func (x *DeployRequest) GetHysteriaConfig() []byte {
	if x != nil {
		return x.HysteriaConfig
	}
	return nil
}

func (x *DeployRequest) GetXrayConfig() []byte {
	if x != nil {
		return x.XrayConfig
	}
	return nil
}

func (x *DeployRequest) GetFiles() map[string][]byte {
	if x != nil {
		return x.Files
	}
	return nil
}

func (x *DeployRequest) GetVxConfig() *server.ServerConfig {
	if x != nil {
		return x.VxConfig
	}
	return nil
}

type DeployResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeployResponse) Reset() {
	*x = DeployResponse{}
	mi := &file_app_api_api_proto_msgTypes[25]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeployResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeployResponse) ProtoMessage() {}

func (x *DeployResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[25]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeployResponse.ProtoReflect.Descriptor instead.
func (*DeployResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{25}
}

type ServerActionRequest struct {
	state         protoimpl.MessageState     `protogen:"open.v1"`
	Action        ServerActionRequest_Action `protobuf:"varint,1,opt,name=action,proto3,enum=x.api.ServerActionRequest_Action" json:"action,omitempty"`
	SshConfig     *ServerSshConfig           `protobuf:"bytes,2,opt,name=ssh_config,json=sshConfig,proto3" json:"ssh_config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerActionRequest) Reset() {
	*x = ServerActionRequest{}
	mi := &file_app_api_api_proto_msgTypes[26]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerActionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerActionRequest) ProtoMessage() {}

func (x *ServerActionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[26]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerActionRequest.ProtoReflect.Descriptor instead.
func (*ServerActionRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{26}
}

func (x *ServerActionRequest) GetAction() ServerActionRequest_Action {
	if x != nil {
		return x.Action
	}
	return ServerActionRequest_ACTION_SHUTDOWN
}

func (x *ServerActionRequest) GetSshConfig() *ServerSshConfig {
	if x != nil {
		return x.SshConfig
	}
	return nil
}

type ServerActionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerActionResponse) Reset() {
	*x = ServerActionResponse{}
	mi := &file_app_api_api_proto_msgTypes[27]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerActionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerActionResponse) ProtoMessage() {}

func (x *ServerActionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[27]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerActionResponse.ProtoReflect.Descriptor instead.
func (*ServerActionResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{27}
}

type VproxyStatusRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SshConfig     *ServerSshConfig       `protobuf:"bytes,1,opt,name=ssh_config,json=sshConfig,proto3" json:"ssh_config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VproxyStatusRequest) Reset() {
	*x = VproxyStatusRequest{}
	mi := &file_app_api_api_proto_msgTypes[28]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VproxyStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VproxyStatusRequest) ProtoMessage() {}

func (x *VproxyStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[28]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VproxyStatusRequest.ProtoReflect.Descriptor instead.
func (*VproxyStatusRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{28}
}

func (x *VproxyStatusRequest) GetSshConfig() *ServerSshConfig {
	if x != nil {
		return x.SshConfig
	}
	return nil
}

type VproxyStatusResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Installed     bool                   `protobuf:"varint,1,opt,name=installed,proto3" json:"installed,omitempty"`
	Version       string                 `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	StartTime     string                 `protobuf:"bytes,3,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	Memory        float32                `protobuf:"fixed32,4,opt,name=memory,proto3" json:"memory,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VproxyStatusResponse) Reset() {
	*x = VproxyStatusResponse{}
	mi := &file_app_api_api_proto_msgTypes[29]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VproxyStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VproxyStatusResponse) ProtoMessage() {}

func (x *VproxyStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[29]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VproxyStatusResponse.ProtoReflect.Descriptor instead.
func (*VproxyStatusResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{29}
}

func (x *VproxyStatusResponse) GetInstalled() bool {
	if x != nil {
		return x.Installed
	}
	return false
}

func (x *VproxyStatusResponse) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *VproxyStatusResponse) GetStartTime() string {
	if x != nil {
		return x.StartTime
	}
	return ""
}

func (x *VproxyStatusResponse) GetMemory() float32 {
	if x != nil {
		return x.Memory
	}
	return 0
}

type VXRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SshConfig     *ServerSshConfig       `protobuf:"bytes,1,opt,name=ssh_config,json=sshConfig,proto3" json:"ssh_config,omitempty"`
	Start         bool                   `protobuf:"varint,2,opt,name=start,proto3" json:"start,omitempty"`
	Stop          bool                   `protobuf:"varint,3,opt,name=stop,proto3" json:"stop,omitempty"`
	Restart       bool                   `protobuf:"varint,4,opt,name=restart,proto3" json:"restart,omitempty"`
	Install       bool                   `protobuf:"varint,5,opt,name=install,proto3" json:"install,omitempty"`
	Uninstall     bool                   `protobuf:"varint,6,opt,name=uninstall,proto3" json:"uninstall,omitempty"`
	Update        bool                   `protobuf:"varint,7,opt,name=update,proto3" json:"update,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VXRequest) Reset() {
	*x = VXRequest{}
	mi := &file_app_api_api_proto_msgTypes[30]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VXRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VXRequest) ProtoMessage() {}

func (x *VXRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[30]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VXRequest.ProtoReflect.Descriptor instead.
func (*VXRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{30}
}

func (x *VXRequest) GetSshConfig() *ServerSshConfig {
	if x != nil {
		return x.SshConfig
	}
	return nil
}

func (x *VXRequest) GetStart() bool {
	if x != nil {
		return x.Start
	}
	return false
}

func (x *VXRequest) GetStop() bool {
	if x != nil {
		return x.Stop
	}
	return false
}

func (x *VXRequest) GetRestart() bool {
	if x != nil {
		return x.Restart
	}
	return false
}

func (x *VXRequest) GetInstall() bool {
	if x != nil {
		return x.Install
	}
	return false
}

func (x *VXRequest) GetUninstall() bool {
	if x != nil {
		return x.Uninstall
	}
	return false
}

func (x *VXRequest) GetUpdate() bool {
	if x != nil {
		return x.Update
	}
	return false
}

type ServerConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SshConfig     *ServerSshConfig       `protobuf:"bytes,1,opt,name=ssh_config,json=sshConfig,proto3" json:"ssh_config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerConfigRequest) Reset() {
	*x = ServerConfigRequest{}
	mi := &file_app_api_api_proto_msgTypes[31]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfigRequest) ProtoMessage() {}

func (x *ServerConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[31]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerConfigRequest.ProtoReflect.Descriptor instead.
func (*ServerConfigRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{31}
}

func (x *ServerConfigRequest) GetSshConfig() *ServerSshConfig {
	if x != nil {
		return x.SshConfig
	}
	return nil
}

type ServerConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Config        *server.ServerConfig   `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerConfigResponse) Reset() {
	*x = ServerConfigResponse{}
	mi := &file_app_api_api_proto_msgTypes[32]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfigResponse) ProtoMessage() {}

func (x *ServerConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[32]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerConfigResponse.ProtoReflect.Descriptor instead.
func (*ServerConfigResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{32}
}

func (x *ServerConfigResponse) GetConfig() *server.ServerConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

type UpdateServerConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SshConfig     *ServerSshConfig       `protobuf:"bytes,1,opt,name=ssh_config,json=sshConfig,proto3" json:"ssh_config,omitempty"`
	Config        *server.ServerConfig   `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateServerConfigRequest) Reset() {
	*x = UpdateServerConfigRequest{}
	mi := &file_app_api_api_proto_msgTypes[33]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateServerConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateServerConfigRequest) ProtoMessage() {}

func (x *UpdateServerConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[33]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateServerConfigRequest.ProtoReflect.Descriptor instead.
func (*UpdateServerConfigRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{33}
}

func (x *UpdateServerConfigRequest) GetSshConfig() *ServerSshConfig {
	if x != nil {
		return x.SshConfig
	}
	return nil
}

func (x *UpdateServerConfigRequest) GetConfig() *server.ServerConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

type UpdateServerConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateServerConfigResponse) Reset() {
	*x = UpdateServerConfigResponse{}
	mi := &file_app_api_api_proto_msgTypes[34]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateServerConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateServerConfigResponse) ProtoMessage() {}

func (x *UpdateServerConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[34]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateServerConfigResponse.ProtoReflect.Descriptor instead.
func (*UpdateServerConfigResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{34}
}

type ProcessGeoFilesRequest struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	GeositeCodes   []string               `protobuf:"bytes,1,rep,name=geosite_codes,json=geositeCodes,proto3" json:"geosite_codes,omitempty"`
	GeoipCodes     []string               `protobuf:"bytes,2,rep,name=geoip_codes,json=geoipCodes,proto3" json:"geoip_codes,omitempty"`
	GeositePath    string                 `protobuf:"bytes,3,opt,name=geosite_path,json=geositePath,proto3" json:"geosite_path,omitempty"`
	GeoipPath      string                 `protobuf:"bytes,4,opt,name=geoip_path,json=geoipPath,proto3" json:"geoip_path,omitempty"`
	DstGeositePath string                 `protobuf:"bytes,5,opt,name=dst_geosite_path,json=dstGeositePath,proto3" json:"dst_geosite_path,omitempty"`
	DstGeoipPath   string                 `protobuf:"bytes,6,opt,name=dst_geoip_path,json=dstGeoipPath,proto3" json:"dst_geoip_path,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *ProcessGeoFilesRequest) Reset() {
	*x = ProcessGeoFilesRequest{}
	mi := &file_app_api_api_proto_msgTypes[35]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProcessGeoFilesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProcessGeoFilesRequest) ProtoMessage() {}

func (x *ProcessGeoFilesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[35]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProcessGeoFilesRequest.ProtoReflect.Descriptor instead.
func (*ProcessGeoFilesRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{35}
}

func (x *ProcessGeoFilesRequest) GetGeositeCodes() []string {
	if x != nil {
		return x.GeositeCodes
	}
	return nil
}

func (x *ProcessGeoFilesRequest) GetGeoipCodes() []string {
	if x != nil {
		return x.GeoipCodes
	}
	return nil
}

func (x *ProcessGeoFilesRequest) GetGeositePath() string {
	if x != nil {
		return x.GeositePath
	}
	return ""
}

func (x *ProcessGeoFilesRequest) GetGeoipPath() string {
	if x != nil {
		return x.GeoipPath
	}
	return ""
}

func (x *ProcessGeoFilesRequest) GetDstGeositePath() string {
	if x != nil {
		return x.DstGeositePath
	}
	return ""
}

func (x *ProcessGeoFilesRequest) GetDstGeoipPath() string {
	if x != nil {
		return x.DstGeoipPath
	}
	return ""
}

type ProcessGeoFilesResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProcessGeoFilesResponse) Reset() {
	*x = ProcessGeoFilesResponse{}
	mi := &file_app_api_api_proto_msgTypes[36]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProcessGeoFilesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProcessGeoFilesResponse) ProtoMessage() {}

func (x *ProcessGeoFilesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[36]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProcessGeoFilesResponse.ProtoReflect.Descriptor instead.
func (*ProcessGeoFilesResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{36}
}

type DecodeRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Data          string                 `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DecodeRequest) Reset() {
	*x = DecodeRequest{}
	mi := &file_app_api_api_proto_msgTypes[37]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DecodeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DecodeRequest) ProtoMessage() {}

func (x *DecodeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[37]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DecodeRequest.ProtoReflect.Descriptor instead.
func (*DecodeRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{37}
}

func (x *DecodeRequest) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

type DecodeResponse struct {
	state         protoimpl.MessageState           `protogen:"open.v1"`
	Handlers      []*configs.OutboundHandlerConfig `protobuf:"bytes,1,rep,name=handlers,proto3" json:"handlers,omitempty"`
	FailedNodes   []string                         `protobuf:"bytes,2,rep,name=failed_nodes,json=failedNodes,proto3" json:"failed_nodes,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DecodeResponse) Reset() {
	*x = DecodeResponse{}
	mi := &file_app_api_api_proto_msgTypes[38]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DecodeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DecodeResponse) ProtoMessage() {}

func (x *DecodeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[38]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DecodeResponse.ProtoReflect.Descriptor instead.
func (*DecodeResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{38}
}

func (x *DecodeResponse) GetHandlers() []*configs.OutboundHandlerConfig {
	if x != nil {
		return x.Handlers
	}
	return nil
}

func (x *DecodeResponse) GetFailedNodes() []string {
	if x != nil {
		return x.FailedNodes
	}
	return nil
}

type GetServerPublicKeyRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SshConfig     *ServerSshConfig       `protobuf:"bytes,1,opt,name=ssh_config,json=sshConfig,proto3" json:"ssh_config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetServerPublicKeyRequest) Reset() {
	*x = GetServerPublicKeyRequest{}
	mi := &file_app_api_api_proto_msgTypes[39]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetServerPublicKeyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetServerPublicKeyRequest) ProtoMessage() {}

func (x *GetServerPublicKeyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[39]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetServerPublicKeyRequest.ProtoReflect.Descriptor instead.
func (*GetServerPublicKeyRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{39}
}

func (x *GetServerPublicKeyRequest) GetSshConfig() *ServerSshConfig {
	if x != nil {
		return x.SshConfig
	}
	return nil
}

type GetServerPublicKeyResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	PublicKey     []byte                 `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetServerPublicKeyResponse) Reset() {
	*x = GetServerPublicKeyResponse{}
	mi := &file_app_api_api_proto_msgTypes[40]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetServerPublicKeyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetServerPublicKeyResponse) ProtoMessage() {}

func (x *GetServerPublicKeyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[40]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetServerPublicKeyResponse.ProtoReflect.Descriptor instead.
func (*GetServerPublicKeyResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{40}
}

func (x *GetServerPublicKeyResponse) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

type GenerateCertRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Domain        string                 `protobuf:"bytes,1,opt,name=domain,proto3" json:"domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateCertRequest) Reset() {
	*x = GenerateCertRequest{}
	mi := &file_app_api_api_proto_msgTypes[41]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateCertRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateCertRequest) ProtoMessage() {}

func (x *GenerateCertRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[41]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenerateCertRequest.ProtoReflect.Descriptor instead.
func (*GenerateCertRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{41}
}

func (x *GenerateCertRequest) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

type GenerateCertResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// pem format
	Cert []byte `protobuf:"bytes,1,opt,name=cert,proto3" json:"cert,omitempty"`
	// pem format
	Key           []byte `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	CertHash      []byte `protobuf:"bytes,3,opt,name=cert_hash,json=certHash,proto3" json:"cert_hash,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateCertResponse) Reset() {
	*x = GenerateCertResponse{}
	mi := &file_app_api_api_proto_msgTypes[42]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateCertResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateCertResponse) ProtoMessage() {}

func (x *GenerateCertResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[42]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenerateCertResponse.ProtoReflect.Descriptor instead.
func (*GenerateCertResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{42}
}

func (x *GenerateCertResponse) GetCert() []byte {
	if x != nil {
		return x.Cert
	}
	return nil
}

func (x *GenerateCertResponse) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *GenerateCertResponse) GetCertHash() []byte {
	if x != nil {
		return x.CertHash
	}
	return nil
}

type GetCertDomainRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Cert          []byte                 `protobuf:"bytes,1,opt,name=cert,proto3" json:"cert,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetCertDomainRequest) Reset() {
	*x = GetCertDomainRequest{}
	mi := &file_app_api_api_proto_msgTypes[43]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetCertDomainRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCertDomainRequest) ProtoMessage() {}

func (x *GetCertDomainRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[43]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCertDomainRequest.ProtoReflect.Descriptor instead.
func (*GetCertDomainRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{43}
}

func (x *GetCertDomainRequest) GetCert() []byte {
	if x != nil {
		return x.Cert
	}
	return nil
}

type GetCertDomainResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Domain        string                 `protobuf:"bytes,1,opt,name=domain,proto3" json:"domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetCertDomainResponse) Reset() {
	*x = GetCertDomainResponse{}
	mi := &file_app_api_api_proto_msgTypes[44]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetCertDomainResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCertDomainResponse) ProtoMessage() {}

func (x *GetCertDomainResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[44]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCertDomainResponse.ProtoReflect.Descriptor instead.
func (*GetCertDomainResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{44}
}

func (x *GetCertDomainResponse) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

type AddInboundRequest struct {
	state         protoimpl.MessageState      `protogen:"open.v1"`
	Inbound       *configs.ProxyInboundConfig `protobuf:"bytes,1,opt,name=inbound,proto3" json:"inbound,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddInboundRequest) Reset() {
	*x = AddInboundRequest{}
	mi := &file_app_api_api_proto_msgTypes[45]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddInboundRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddInboundRequest) ProtoMessage() {}

func (x *AddInboundRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[45]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddInboundRequest.ProtoReflect.Descriptor instead.
func (*AddInboundRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{45}
}

func (x *AddInboundRequest) GetInbound() *configs.ProxyInboundConfig {
	if x != nil {
		return x.Inbound
	}
	return nil
}

type AddInboundResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddInboundResponse) Reset() {
	*x = AddInboundResponse{}
	mi := &file_app_api_api_proto_msgTypes[46]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddInboundResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddInboundResponse) ProtoMessage() {}

func (x *AddInboundResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[46]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddInboundResponse.ProtoReflect.Descriptor instead.
func (*AddInboundResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{46}
}

type UploadLogRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Body          string                 `protobuf:"bytes,1,opt,name=body,proto3" json:"body,omitempty"`
	Version       string                 `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Secret        string                 `protobuf:"bytes,3,opt,name=secret,proto3" json:"secret,omitempty"`
	Ca            []byte                 `protobuf:"bytes,4,opt,name=ca,proto3" json:"ca,omitempty"`
	Url           string                 `protobuf:"bytes,5,opt,name=url,proto3" json:"url,omitempty"`
	Sni           string                 `protobuf:"bytes,6,opt,name=sni,proto3" json:"sni,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UploadLogRequest) Reset() {
	*x = UploadLogRequest{}
	mi := &file_app_api_api_proto_msgTypes[47]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UploadLogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadLogRequest) ProtoMessage() {}

func (x *UploadLogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[47]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadLogRequest.ProtoReflect.Descriptor instead.
func (*UploadLogRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{47}
}

func (x *UploadLogRequest) GetBody() string {
	if x != nil {
		return x.Body
	}
	return ""
}

func (x *UploadLogRequest) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *UploadLogRequest) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

func (x *UploadLogRequest) GetCa() []byte {
	if x != nil {
		return x.Ca
	}
	return nil
}

func (x *UploadLogRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *UploadLogRequest) GetSni() string {
	if x != nil {
		return x.Sni
	}
	return ""
}

type UploadLogResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UploadLogResponse) Reset() {
	*x = UploadLogResponse{}
	mi := &file_app_api_api_proto_msgTypes[48]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UploadLogResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadLogResponse) ProtoMessage() {}

func (x *UploadLogResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[48]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadLogResponse.ProtoReflect.Descriptor instead.
func (*UploadLogResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{48}
}

type DefaultNICHasGlobalV6Request struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DefaultNICHasGlobalV6Request) Reset() {
	*x = DefaultNICHasGlobalV6Request{}
	mi := &file_app_api_api_proto_msgTypes[49]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DefaultNICHasGlobalV6Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DefaultNICHasGlobalV6Request) ProtoMessage() {}

func (x *DefaultNICHasGlobalV6Request) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[49]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DefaultNICHasGlobalV6Request.ProtoReflect.Descriptor instead.
func (*DefaultNICHasGlobalV6Request) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{49}
}

type DefaultNICHasGlobalV6Response struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	HasGlobalV6   bool                   `protobuf:"varint,1,opt,name=has_global_v6,json=hasGlobalV6,proto3" json:"has_global_v6,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DefaultNICHasGlobalV6Response) Reset() {
	*x = DefaultNICHasGlobalV6Response{}
	mi := &file_app_api_api_proto_msgTypes[50]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DefaultNICHasGlobalV6Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DefaultNICHasGlobalV6Response) ProtoMessage() {}

func (x *DefaultNICHasGlobalV6Response) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[50]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DefaultNICHasGlobalV6Response.ProtoReflect.Descriptor instead.
func (*DefaultNICHasGlobalV6Response) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{50}
}

func (x *DefaultNICHasGlobalV6Response) GetHasGlobalV6() bool {
	if x != nil {
		return x.HasGlobalV6
	}
	return false
}

type UpdateTmStatusRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	On            bool                   `protobuf:"varint,1,opt,name=on,proto3" json:"on,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateTmStatusRequest) Reset() {
	*x = UpdateTmStatusRequest{}
	mi := &file_app_api_api_proto_msgTypes[51]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateTmStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateTmStatusRequest) ProtoMessage() {}

func (x *UpdateTmStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[51]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateTmStatusRequest.ProtoReflect.Descriptor instead.
func (*UpdateTmStatusRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{51}
}

func (x *UpdateTmStatusRequest) GetOn() bool {
	if x != nil {
		return x.On
	}
	return false
}

type Receipt struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Receipt) Reset() {
	*x = Receipt{}
	mi := &file_app_api_api_proto_msgTypes[52]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Receipt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Receipt) ProtoMessage() {}

func (x *Receipt) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[52]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Receipt.ProtoReflect.Descriptor instead.
func (*Receipt) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{52}
}

type ParseClashRuleFileRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Content       []byte                 `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ParseClashRuleFileRequest) Reset() {
	*x = ParseClashRuleFileRequest{}
	mi := &file_app_api_api_proto_msgTypes[53]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ParseClashRuleFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseClashRuleFileRequest) ProtoMessage() {}

func (x *ParseClashRuleFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[53]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseClashRuleFileRequest.ProtoReflect.Descriptor instead.
func (*ParseClashRuleFileRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{53}
}

func (x *ParseClashRuleFileRequest) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

type ParseClashRuleFileResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Domains       []*geo.Domain          `protobuf:"bytes,1,rep,name=domains,proto3" json:"domains,omitempty"`
	Cidrs         []*geo.CIDR            `protobuf:"bytes,2,rep,name=cidrs,proto3" json:"cidrs,omitempty"`
	AppIds        []*configs.AppId       `protobuf:"bytes,3,rep,name=app_ids,json=appIds,proto3" json:"app_ids,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ParseClashRuleFileResponse) Reset() {
	*x = ParseClashRuleFileResponse{}
	mi := &file_app_api_api_proto_msgTypes[54]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ParseClashRuleFileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseClashRuleFileResponse) ProtoMessage() {}

func (x *ParseClashRuleFileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[54]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseClashRuleFileResponse.ProtoReflect.Descriptor instead.
func (*ParseClashRuleFileResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{54}
}

func (x *ParseClashRuleFileResponse) GetDomains() []*geo.Domain {
	if x != nil {
		return x.Domains
	}
	return nil
}

func (x *ParseClashRuleFileResponse) GetCidrs() []*geo.CIDR {
	if x != nil {
		return x.Cidrs
	}
	return nil
}

func (x *ParseClashRuleFileResponse) GetAppIds() []*configs.AppId {
	if x != nil {
		return x.AppIds
	}
	return nil
}

type ParseGeositeConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Config        *configs.GeositeConfig `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ParseGeositeConfigRequest) Reset() {
	*x = ParseGeositeConfigRequest{}
	mi := &file_app_api_api_proto_msgTypes[55]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ParseGeositeConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseGeositeConfigRequest) ProtoMessage() {}

func (x *ParseGeositeConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[55]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseGeositeConfigRequest.ProtoReflect.Descriptor instead.
func (*ParseGeositeConfigRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{55}
}

func (x *ParseGeositeConfigRequest) GetConfig() *configs.GeositeConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

type ParseGeositeConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Domains       []*geo.Domain          `protobuf:"bytes,1,rep,name=domains,proto3" json:"domains,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ParseGeositeConfigResponse) Reset() {
	*x = ParseGeositeConfigResponse{}
	mi := &file_app_api_api_proto_msgTypes[56]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ParseGeositeConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseGeositeConfigResponse) ProtoMessage() {}

func (x *ParseGeositeConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[56]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseGeositeConfigResponse.ProtoReflect.Descriptor instead.
func (*ParseGeositeConfigResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{56}
}

func (x *ParseGeositeConfigResponse) GetDomains() []*geo.Domain {
	if x != nil {
		return x.Domains
	}
	return nil
}

type ParseGeoIPConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Config        *configs.GeoIPConfig   `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ParseGeoIPConfigRequest) Reset() {
	*x = ParseGeoIPConfigRequest{}
	mi := &file_app_api_api_proto_msgTypes[57]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ParseGeoIPConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseGeoIPConfigRequest) ProtoMessage() {}

func (x *ParseGeoIPConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[57]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseGeoIPConfigRequest.ProtoReflect.Descriptor instead.
func (*ParseGeoIPConfigRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{57}
}

func (x *ParseGeoIPConfigRequest) GetConfig() *configs.GeoIPConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

type ParseGeoIPConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Cidrs         []*geo.CIDR            `protobuf:"bytes,1,rep,name=cidrs,proto3" json:"cidrs,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ParseGeoIPConfigResponse) Reset() {
	*x = ParseGeoIPConfigResponse{}
	mi := &file_app_api_api_proto_msgTypes[58]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ParseGeoIPConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseGeoIPConfigResponse) ProtoMessage() {}

func (x *ParseGeoIPConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[58]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseGeoIPConfigResponse.ProtoReflect.Descriptor instead.
func (*ParseGeoIPConfigResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{58}
}

func (x *ParseGeoIPConfigResponse) GetCidrs() []*geo.CIDR {
	if x != nil {
		return x.Cidrs
	}
	return nil
}

type RunRealiScannerRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Addr          string                 `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RunRealiScannerRequest) Reset() {
	*x = RunRealiScannerRequest{}
	mi := &file_app_api_api_proto_msgTypes[59]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RunRealiScannerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunRealiScannerRequest) ProtoMessage() {}

func (x *RunRealiScannerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[59]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunRealiScannerRequest.ProtoReflect.Descriptor instead.
func (*RunRealiScannerRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{59}
}

func (x *RunRealiScannerRequest) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

type RunRealiScannerResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Results       []*RealiScannerResult  `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RunRealiScannerResponse) Reset() {
	*x = RunRealiScannerResponse{}
	mi := &file_app_api_api_proto_msgTypes[60]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RunRealiScannerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunRealiScannerResponse) ProtoMessage() {}

func (x *RunRealiScannerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[60]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunRealiScannerResponse.ProtoReflect.Descriptor instead.
func (*RunRealiScannerResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{60}
}

func (x *RunRealiScannerResponse) GetResults() []*RealiScannerResult {
	if x != nil {
		return x.Results
	}
	return nil
}

type RealiScannerResult struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Ip            string                 `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Domain        string                 `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RealiScannerResult) Reset() {
	*x = RealiScannerResult{}
	mi := &file_app_api_api_proto_msgTypes[61]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RealiScannerResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RealiScannerResult) ProtoMessage() {}

func (x *RealiScannerResult) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[61]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RealiScannerResult.ProtoReflect.Descriptor instead.
func (*RealiScannerResult) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{61}
}

func (x *RealiScannerResult) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *RealiScannerResult) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

type GenerateX25519KeyPairRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateX25519KeyPairRequest) Reset() {
	*x = GenerateX25519KeyPairRequest{}
	mi := &file_app_api_api_proto_msgTypes[62]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateX25519KeyPairRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateX25519KeyPairRequest) ProtoMessage() {}

func (x *GenerateX25519KeyPairRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[62]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenerateX25519KeyPairRequest.ProtoReflect.Descriptor instead.
func (*GenerateX25519KeyPairRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{62}
}

type GenerateX25519KeyPairResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Pub           string                 `protobuf:"bytes,1,opt,name=pub,proto3" json:"pub,omitempty"`
	Pri           string                 `protobuf:"bytes,2,opt,name=pri,proto3" json:"pri,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateX25519KeyPairResponse) Reset() {
	*x = GenerateX25519KeyPairResponse{}
	mi := &file_app_api_api_proto_msgTypes[63]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateX25519KeyPairResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateX25519KeyPairResponse) ProtoMessage() {}

func (x *GenerateX25519KeyPairResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[63]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenerateX25519KeyPairResponse.ProtoReflect.Descriptor instead.
func (*GenerateX25519KeyPairResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{63}
}

func (x *GenerateX25519KeyPairResponse) GetPub() string {
	if x != nil {
		return x.Pub
	}
	return ""
}

func (x *GenerateX25519KeyPairResponse) GetPri() string {
	if x != nil {
		return x.Pri
	}
	return ""
}

type StartMacSystemProxyRequest struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	HttpProxyAddress  string                 `protobuf:"bytes,1,opt,name=http_proxy_address,json=httpProxyAddress,proto3" json:"http_proxy_address,omitempty"`
	HttpProxyPort     uint32                 `protobuf:"varint,2,opt,name=http_proxy_port,json=httpProxyPort,proto3" json:"http_proxy_port,omitempty"`
	HttpsProxyAddress string                 `protobuf:"bytes,3,opt,name=https_proxy_address,json=httpsProxyAddress,proto3" json:"https_proxy_address,omitempty"`
	HttpsProxyPort    uint32                 `protobuf:"varint,4,opt,name=https_proxy_port,json=httpsProxyPort,proto3" json:"https_proxy_port,omitempty"`
	SocksProxyAddress string                 `protobuf:"bytes,5,opt,name=socks_proxy_address,json=socksProxyAddress,proto3" json:"socks_proxy_address,omitempty"`
	SocksProxyPort    uint32                 `protobuf:"varint,6,opt,name=socks_proxy_port,json=socksProxyPort,proto3" json:"socks_proxy_port,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *StartMacSystemProxyRequest) Reset() {
	*x = StartMacSystemProxyRequest{}
	mi := &file_app_api_api_proto_msgTypes[64]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StartMacSystemProxyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartMacSystemProxyRequest) ProtoMessage() {}

func (x *StartMacSystemProxyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[64]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartMacSystemProxyRequest.ProtoReflect.Descriptor instead.
func (*StartMacSystemProxyRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{64}
}

func (x *StartMacSystemProxyRequest) GetHttpProxyAddress() string {
	if x != nil {
		return x.HttpProxyAddress
	}
	return ""
}

func (x *StartMacSystemProxyRequest) GetHttpProxyPort() uint32 {
	if x != nil {
		return x.HttpProxyPort
	}
	return 0
}

func (x *StartMacSystemProxyRequest) GetHttpsProxyAddress() string {
	if x != nil {
		return x.HttpsProxyAddress
	}
	return ""
}

func (x *StartMacSystemProxyRequest) GetHttpsProxyPort() uint32 {
	if x != nil {
		return x.HttpsProxyPort
	}
	return 0
}

func (x *StartMacSystemProxyRequest) GetSocksProxyAddress() string {
	if x != nil {
		return x.SocksProxyAddress
	}
	return ""
}

func (x *StartMacSystemProxyRequest) GetSocksProxyPort() uint32 {
	if x != nil {
		return x.SocksProxyPort
	}
	return 0
}

type StopMacSystemProxyRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StopMacSystemProxyRequest) Reset() {
	*x = StopMacSystemProxyRequest{}
	mi := &file_app_api_api_proto_msgTypes[65]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StopMacSystemProxyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopMacSystemProxyRequest) ProtoMessage() {}

func (x *StopMacSystemProxyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[65]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopMacSystemProxyRequest.ProtoReflect.Descriptor instead.
func (*StopMacSystemProxyRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{65}
}

type CloseDbRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CloseDbRequest) Reset() {
	*x = CloseDbRequest{}
	mi := &file_app_api_api_proto_msgTypes[66]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CloseDbRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloseDbRequest) ProtoMessage() {}

func (x *CloseDbRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[66]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloseDbRequest.ProtoReflect.Descriptor instead.
func (*CloseDbRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{66}
}

type OpenDbRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Path          string                 `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OpenDbRequest) Reset() {
	*x = OpenDbRequest{}
	mi := &file_app_api_api_proto_msgTypes[67]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OpenDbRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OpenDbRequest) ProtoMessage() {}

func (x *OpenDbRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[67]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OpenDbRequest.ProtoReflect.Descriptor instead.
func (*OpenDbRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{67}
}

func (x *OpenDbRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type InboundConfigToOutboundConfigRequest struct {
	state         protoimpl.MessageState           `protogen:"open.v1"`
	Inbound       *configs.ProxyInboundConfig      `protobuf:"bytes,1,opt,name=inbound,proto3" json:"inbound,omitempty"`
	MultiInbound  *configs.MultiProxyInboundConfig `protobuf:"bytes,2,opt,name=multi_inbound,json=multiInbound,proto3" json:"multi_inbound,omitempty"`
	ServerAddress string                           `protobuf:"bytes,3,opt,name=server_address,json=serverAddress,proto3" json:"server_address,omitempty"`
	ServerName    string                           `protobuf:"bytes,4,opt,name=server_name,json=serverName,proto3" json:"server_name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InboundConfigToOutboundConfigRequest) Reset() {
	*x = InboundConfigToOutboundConfigRequest{}
	mi := &file_app_api_api_proto_msgTypes[68]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InboundConfigToOutboundConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InboundConfigToOutboundConfigRequest) ProtoMessage() {}

func (x *InboundConfigToOutboundConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[68]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InboundConfigToOutboundConfigRequest.ProtoReflect.Descriptor instead.
func (*InboundConfigToOutboundConfigRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{68}
}

func (x *InboundConfigToOutboundConfigRequest) GetInbound() *configs.ProxyInboundConfig {
	if x != nil {
		return x.Inbound
	}
	return nil
}

func (x *InboundConfigToOutboundConfigRequest) GetMultiInbound() *configs.MultiProxyInboundConfig {
	if x != nil {
		return x.MultiInbound
	}
	return nil
}

func (x *InboundConfigToOutboundConfigRequest) GetServerAddress() string {
	if x != nil {
		return x.ServerAddress
	}
	return ""
}

func (x *InboundConfigToOutboundConfigRequest) GetServerName() string {
	if x != nil {
		return x.ServerName
	}
	return ""
}

type InboundConfigToOutboundConfigResponse struct {
	state           protoimpl.MessageState           `protogen:"open.v1"`
	OutboundConfigs []*configs.OutboundHandlerConfig `protobuf:"bytes,1,rep,name=outbound_configs,json=outboundConfigs,proto3" json:"outbound_configs,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *InboundConfigToOutboundConfigResponse) Reset() {
	*x = InboundConfigToOutboundConfigResponse{}
	mi := &file_app_api_api_proto_msgTypes[69]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InboundConfigToOutboundConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InboundConfigToOutboundConfigResponse) ProtoMessage() {}

func (x *InboundConfigToOutboundConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[69]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InboundConfigToOutboundConfigResponse.ProtoReflect.Descriptor instead.
func (*InboundConfigToOutboundConfigResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{69}
}

func (x *InboundConfigToOutboundConfigResponse) GetOutboundConfigs() []*configs.OutboundHandlerConfig {
	if x != nil {
		return x.OutboundConfigs
	}
	return nil
}

type ToUrlRequest struct {
	state           protoimpl.MessageState           `protogen:"open.v1"`
	OutboundConfogs []*configs.OutboundHandlerConfig `protobuf:"bytes,1,rep,name=outbound_confogs,json=outboundConfogs,proto3" json:"outbound_confogs,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *ToUrlRequest) Reset() {
	*x = ToUrlRequest{}
	mi := &file_app_api_api_proto_msgTypes[70]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToUrlRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToUrlRequest) ProtoMessage() {}

func (x *ToUrlRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[70]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToUrlRequest.ProtoReflect.Descriptor instead.
func (*ToUrlRequest) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{70}
}

func (x *ToUrlRequest) GetOutboundConfogs() []*configs.OutboundHandlerConfig {
	if x != nil {
		return x.OutboundConfogs
	}
	return nil
}

type ToUrlResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Urls          []string               `protobuf:"bytes,1,rep,name=urls,proto3" json:"urls,omitempty"`
	FailedNodes   []string               `protobuf:"bytes,2,rep,name=failed_nodes,json=failedNodes,proto3" json:"failed_nodes,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ToUrlResponse) Reset() {
	*x = ToUrlResponse{}
	mi := &file_app_api_api_proto_msgTypes[71]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToUrlResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToUrlResponse) ProtoMessage() {}

func (x *ToUrlResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_api_api_proto_msgTypes[71]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToUrlResponse.ProtoReflect.Descriptor instead.
func (*ToUrlResponse) Descriptor() ([]byte, []int) {
	return file_app_api_api_proto_rawDescGZIP(), []int{71}
}

func (x *ToUrlResponse) GetUrls() []string {
	if x != nil {
		return x.Urls
	}
	return nil
}

func (x *ToUrlResponse) GetFailedNodes() []string {
	if x != nil {
		return x.FailedNodes
	}
	return nil
}

var File_app_api_api_proto protoreflect.FileDescriptor

const file_app_api_api_proto_rawDesc = "" +
	"\n" +
	"\x11app/api/api.proto\x12\x05x.api\x1a\x15protos/outbound.proto\x1a\x14protos/inbound.proto\x1a\x13protos/router.proto\x1a\x14common/geo/geo.proto\x1a\x10protos/geo.proto\x1a\x1aprotos/server/server.proto\"\xb8\x02\n" +
	"\x0fApiServerConfig\x12\x1f\n" +
	"\vlisten_addr\x18\x01 \x01(\tR\n" +
	"listenAddr\x12\x1d\n" +
	"\n" +
	"geoip_path\x18\x02 \x01(\tR\tgeoipPath\x12\x19\n" +
	"\btun_name\x18\x03 \x01(\tR\atunName\x12\x1b\n" +
	"\tlog_level\x18\x04 \x01(\rR\blogLevel\x12\x17\n" +
	"\adb_path\x18\x05 \x01(\tR\x06dbPath\x12(\n" +
	"\x10last_update_time\x18\x06 \x01(\rR\x0elastUpdateTime\x12\x1a\n" +
	"\binterval\x18\a \x01(\rR\binterval\x12-\n" +
	"\x13bind_to_default_nic\x18\b \x01(\bR\x10bindToDefaultNic\x12\x1f\n" +
	"\vclient_cert\x18\t \x01(\fR\n" +
	"clientCert\"\xb5\x01\n" +
	"\x1aXStatusChangeNotifyRequest\x12@\n" +
	"\x06status\x18\x01 \x01(\x0e2(.x.api.XStatusChangeNotifyRequest.StatusR\x06status\"U\n" +
	"\x06Status\x12\x10\n" +
	"\fSTATUS_START\x10\x00\x12\x0f\n" +
	"\vSTATUS_STOP\x10\x01\x12\x13\n" +
	"\x0fSTATUS_STARTING\x10\x02\x12\x13\n" +
	"\x0fSTATUS_STOPPING\x10\x03\"\x1d\n" +
	"\x1bXStatusChangeNotifyResponse\"<\n" +
	"\x1eSetSubscriptionIntervalRequest\x12\x1a\n" +
	"\binterval\x18\x01 \x01(\x05R\binterval\"!\n" +
	"\x1fSetSubscriptionIntervalResponse\"k\n" +
	"\x19UpdateSubscriptionRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x10\n" +
	"\x03all\x18\x02 \x01(\bR\x03all\x12,\n" +
	"\bhandlers\x18\x03 \x03(\v2\x10.x.HandlerConfigR\bhandlers\"\xad\x02\n" +
	"\x1aUpdateSubscriptionResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\x05R\asuccess\x12\x12\n" +
	"\x04fail\x18\x02 \x01(\x05R\x04fail\x12#\n" +
	"\rsuccess_nodes\x18\x03 \x01(\x05R\fsuccessNodes\x12X\n" +
	"\rerror_reasons\x18\x05 \x03(\v23.x.api.UpdateSubscriptionResponse.ErrorReasonsEntryR\ferrorReasons\x12!\n" +
	"\ffailed_nodes\x18\x06 \x03(\tR\vfailedNodes\x1a?\n" +
	"\x11ErrorReasonsEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\".\n" +
	"\x11SetTunNameRequest\x12\x19\n" +
	"\btun_name\x18\x01 \x01(\tR\atunName\"\x14\n" +
	"\x12SetTunNameResponse\"e\n" +
	"\x0fDownloadRequest\x12\x10\n" +
	"\x03url\x18\x01 \x01(\tR\x03url\x12,\n" +
	"\bhandlers\x18\x02 \x03(\v2\x10.x.HandlerConfigR\bhandlers\x12\x12\n" +
	"\x04dest\x18\x03 \x01(\tR\x04dest\"\x9a\x01\n" +
	"\x10DownloadResponse\x128\n" +
	"\x05usage\x18\x01 \x03(\v2\".x.api.DownloadResponse.UsageEntryR\x05usage\x12\x12\n" +
	"\x04data\x18\x03 \x01(\fR\x04data\x1a8\n" +
	"\n" +
	"UsageEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\rR\x05value:\x028\x01\"F\n" +
	"\x10HandlerIpRequest\x122\n" +
	"\ahandler\x18\x01 \x01(\v2\x18.x.OutboundHandlerConfigR\ahandler\"8\n" +
	"\x0eRttTestRequest\x12\x12\n" +
	"\x04addr\x18\x01 \x01(\tR\x04addr\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\"%\n" +
	"\x0fRttTestResponse\x12\x12\n" +
	"\x04ping\x18\x01 \x01(\rR\x04ping\"7\n" +
	"\x11HandlerIpResponse\x12\x10\n" +
	"\x03ip6\x18\x01 \x01(\tR\x03ip6\x12\x10\n" +
	"\x03ip4\x18\x02 \x01(\tR\x03ip4\"B\n" +
	"\x14HandlerUsableRequest\x12*\n" +
	"\ahandler\x18\x01 \x01(\v2\x10.x.HandlerConfigR\ahandler\";\n" +
	"\x15HandlerUsableResponse\x12\x12\n" +
	"\x04ping\x18\x01 \x01(\x05R\x04ping\x12\x0e\n" +
	"\x02ip\x18\x02 \x01(\tR\x02ip\"T\n" +
	"\x10SpeedTestRequest\x12,\n" +
	"\bhandlers\x18\x01 \x03(\v2\x10.x.HandlerConfigR\bhandlers\x12\x12\n" +
	"\x04size\x18\x02 \x01(\rR\x04size\"X\n" +
	"\x11SpeedTestResponse\x12\x12\n" +
	"\x04down\x18\x03 \x01(\x05R\x04down\x12\x10\n" +
	"\x03tag\x18\x04 \x01(\tR\x03tag\x12\x1d\n" +
	"\n" +
	"usage_down\x18\a \x01(\rR\tusageDown\" \n" +
	"\fGeoIPRequest\x12\x10\n" +
	"\x03ips\x18\x01 \x03(\tR\x03ips\"-\n" +
	"\rGeoIPResponse\x12\x1c\n" +
	"\tcountries\x18\x01 \x03(\tR\tcountries\"\x8f\x02\n" +
	"\x0fServerSshConfig\x12\x18\n" +
	"\aaddress\x18\x01 \x01(\tR\aaddress\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\x12\x1a\n" +
	"\busername\x18\x03 \x01(\tR\busername\x12#\n" +
	"\rsudo_password\x18\x04 \x01(\tR\fsudoPassword\x12\x17\n" +
	"\assh_key\x18\x06 \x01(\fR\x06sshKey\x12 \n" +
	"\fssh_key_path\x18\a \x01(\tR\n" +
	"sshKeyPath\x12,\n" +
	"\x12ssh_key_passphrase\x18\b \x01(\tR\x10sshKeyPassphrase\x12$\n" +
	"\x0eserver_pub_key\x18\t \x01(\fR\fserverPubKey\"i\n" +
	"\x14MonitorServerRequest\x125\n" +
	"\n" +
	"ssh_config\x18\x01 \x01(\v2\x16.x.api.ServerSshConfigR\tsshConfig\x12\x1a\n" +
	"\binterval\x18\x02 \x01(\rR\binterval\"\xb5\x02\n" +
	"\x15MonitorServerResponse\x12\x10\n" +
	"\x03cpu\x18\x01 \x01(\rR\x03cpu\x12\x1f\n" +
	"\vused_memory\x18\x02 \x01(\x04R\n" +
	"usedMemory\x12!\n" +
	"\ftotal_memory\x18\x03 \x01(\x04R\vtotalMemory\x12\x1b\n" +
	"\tused_disk\x18\x04 \x01(\rR\busedDisk\x12\x1d\n" +
	"\n" +
	"total_disk\x18\x05 \x01(\rR\ttotalDisk\x12 \n" +
	"\fnet_in_speed\x18\x06 \x01(\rR\n" +
	"netInSpeed\x12\"\n" +
	"\rnet_out_speed\x18\a \x01(\rR\vnetOutSpeed\x12 \n" +
	"\fnet_in_usage\x18\b \x01(\x04R\n" +
	"netInUsage\x12\"\n" +
	"\rnet_out_usage\x18\t \x01(\x04R\vnetOutUsage\"\xaf\x02\n" +
	"\rDeployRequest\x125\n" +
	"\n" +
	"ssh_config\x18\x01 \x01(\v2\x16.x.api.ServerSshConfigR\tsshConfig\x12'\n" +
	"\x0fhysteria_config\x18\x02 \x01(\fR\x0ehysteriaConfig\x12\x1f\n" +
	"\vxray_config\x18\x03 \x01(\fR\n" +
	"xrayConfig\x125\n" +
	"\x05files\x18\x04 \x03(\v2\x1f.x.api.DeployRequest.FilesEntryR\x05files\x12,\n" +
	"\tvx_config\x18\x05 \x01(\v2\x0f.x.ServerConfigR\bvxConfig\x1a8\n" +
	"\n" +
	"FilesEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\fR\x05value:\x028\x01\"\x10\n" +
	"\x0eDeployResponse\"\xba\x01\n" +
	"\x13ServerActionRequest\x129\n" +
	"\x06action\x18\x01 \x01(\x0e2!.x.api.ServerActionRequest.ActionR\x06action\x125\n" +
	"\n" +
	"ssh_config\x18\x02 \x01(\v2\x16.x.api.ServerSshConfigR\tsshConfig\"1\n" +
	"\x06Action\x12\x13\n" +
	"\x0fACTION_SHUTDOWN\x10\x00\x12\x12\n" +
	"\x0eACTION_RESTART\x10\x01\"\x16\n" +
	"\x14ServerActionResponse\"L\n" +
	"\x13VproxyStatusRequest\x125\n" +
	"\n" +
	"ssh_config\x18\x01 \x01(\v2\x16.x.api.ServerSshConfigR\tsshConfig\"\x85\x01\n" +
	"\x14VproxyStatusResponse\x12\x1c\n" +
	"\tinstalled\x18\x01 \x01(\bR\tinstalled\x12\x18\n" +
	"\aversion\x18\x02 \x01(\tR\aversion\x12\x1d\n" +
	"\n" +
	"start_time\x18\x03 \x01(\tR\tstartTime\x12\x16\n" +
	"\x06memory\x18\x04 \x01(\x02R\x06memory\"\xd6\x01\n" +
	"\tVXRequest\x125\n" +
	"\n" +
	"ssh_config\x18\x01 \x01(\v2\x16.x.api.ServerSshConfigR\tsshConfig\x12\x14\n" +
	"\x05start\x18\x02 \x01(\bR\x05start\x12\x12\n" +
	"\x04stop\x18\x03 \x01(\bR\x04stop\x12\x18\n" +
	"\arestart\x18\x04 \x01(\bR\arestart\x12\x18\n" +
	"\ainstall\x18\x05 \x01(\bR\ainstall\x12\x1c\n" +
	"\tuninstall\x18\x06 \x01(\bR\tuninstall\x12\x16\n" +
	"\x06update\x18\a \x01(\bR\x06update\"L\n" +
	"\x13ServerConfigRequest\x125\n" +
	"\n" +
	"ssh_config\x18\x01 \x01(\v2\x16.x.api.ServerSshConfigR\tsshConfig\"?\n" +
	"\x14ServerConfigResponse\x12'\n" +
	"\x06config\x18\x01 \x01(\v2\x0f.x.ServerConfigR\x06config\"{\n" +
	"\x19UpdateServerConfigRequest\x125\n" +
	"\n" +
	"ssh_config\x18\x01 \x01(\v2\x16.x.api.ServerSshConfigR\tsshConfig\x12'\n" +
	"\x06config\x18\x02 \x01(\v2\x0f.x.ServerConfigR\x06config\"\x1c\n" +
	"\x1aUpdateServerConfigResponse\"\xf0\x01\n" +
	"\x16ProcessGeoFilesRequest\x12#\n" +
	"\rgeosite_codes\x18\x01 \x03(\tR\fgeositeCodes\x12\x1f\n" +
	"\vgeoip_codes\x18\x02 \x03(\tR\n" +
	"geoipCodes\x12!\n" +
	"\fgeosite_path\x18\x03 \x01(\tR\vgeositePath\x12\x1d\n" +
	"\n" +
	"geoip_path\x18\x04 \x01(\tR\tgeoipPath\x12(\n" +
	"\x10dst_geosite_path\x18\x05 \x01(\tR\x0edstGeositePath\x12$\n" +
	"\x0edst_geoip_path\x18\x06 \x01(\tR\fdstGeoipPath\"\x19\n" +
	"\x17ProcessGeoFilesResponse\"#\n" +
	"\rDecodeRequest\x12\x12\n" +
	"\x04data\x18\x01 \x01(\tR\x04data\"i\n" +
	"\x0eDecodeResponse\x124\n" +
	"\bhandlers\x18\x01 \x03(\v2\x18.x.OutboundHandlerConfigR\bhandlers\x12!\n" +
	"\ffailed_nodes\x18\x02 \x03(\tR\vfailedNodes\"R\n" +
	"\x19GetServerPublicKeyRequest\x125\n" +
	"\n" +
	"ssh_config\x18\x01 \x01(\v2\x16.x.api.ServerSshConfigR\tsshConfig\";\n" +
	"\x1aGetServerPublicKeyResponse\x12\x1d\n" +
	"\n" +
	"public_key\x18\x01 \x01(\fR\tpublicKey\"-\n" +
	"\x13GenerateCertRequest\x12\x16\n" +
	"\x06domain\x18\x01 \x01(\tR\x06domain\"Y\n" +
	"\x14GenerateCertResponse\x12\x12\n" +
	"\x04cert\x18\x01 \x01(\fR\x04cert\x12\x10\n" +
	"\x03key\x18\x02 \x01(\fR\x03key\x12\x1b\n" +
	"\tcert_hash\x18\x03 \x01(\fR\bcertHash\"*\n" +
	"\x14GetCertDomainRequest\x12\x12\n" +
	"\x04cert\x18\x01 \x01(\fR\x04cert\"/\n" +
	"\x15GetCertDomainResponse\x12\x16\n" +
	"\x06domain\x18\x01 \x01(\tR\x06domain\"D\n" +
	"\x11AddInboundRequest\x12/\n" +
	"\ainbound\x18\x01 \x01(\v2\x15.x.ProxyInboundConfigR\ainbound\"\x14\n" +
	"\x12AddInboundResponse\"\x8c\x01\n" +
	"\x10UploadLogRequest\x12\x12\n" +
	"\x04body\x18\x01 \x01(\tR\x04body\x12\x18\n" +
	"\aversion\x18\x02 \x01(\tR\aversion\x12\x16\n" +
	"\x06secret\x18\x03 \x01(\tR\x06secret\x12\x0e\n" +
	"\x02ca\x18\x04 \x01(\fR\x02ca\x12\x10\n" +
	"\x03url\x18\x05 \x01(\tR\x03url\x12\x10\n" +
	"\x03sni\x18\x06 \x01(\tR\x03sni\"\x13\n" +
	"\x11UploadLogResponse\"\x1e\n" +
	"\x1cDefaultNICHasGlobalV6Request\"C\n" +
	"\x1dDefaultNICHasGlobalV6Response\x12\"\n" +
	"\rhas_global_v6\x18\x01 \x01(\bR\vhasGlobalV6\"'\n" +
	"\x15UpdateTmStatusRequest\x12\x0e\n" +
	"\x02on\x18\x01 \x01(\bR\x02on\"\t\n" +
	"\aReceipt\"5\n" +
	"\x19ParseClashRuleFileRequest\x12\x18\n" +
	"\acontent\x18\x01 \x01(\fR\acontent\"\x99\x01\n" +
	"\x1aParseClashRuleFileResponse\x12.\n" +
	"\adomains\x18\x01 \x03(\v2\x14.x.common.geo.DomainR\adomains\x12(\n" +
	"\x05cidrs\x18\x02 \x03(\v2\x12.x.common.geo.CIDRR\x05cidrs\x12!\n" +
	"\aapp_ids\x18\x03 \x03(\v2\b.x.AppIdR\x06appIds\"E\n" +
	"\x19ParseGeositeConfigRequest\x12(\n" +
	"\x06config\x18\x01 \x01(\v2\x10.x.GeositeConfigR\x06config\"L\n" +
	"\x1aParseGeositeConfigResponse\x12.\n" +
	"\adomains\x18\x01 \x03(\v2\x14.x.common.geo.DomainR\adomains\"A\n" +
	"\x17ParseGeoIPConfigRequest\x12&\n" +
	"\x06config\x18\x01 \x01(\v2\x0e.x.GeoIPConfigR\x06config\"D\n" +
	"\x18ParseGeoIPConfigResponse\x12(\n" +
	"\x05cidrs\x18\x01 \x03(\v2\x12.x.common.geo.CIDRR\x05cidrs\",\n" +
	"\x16RunRealiScannerRequest\x12\x12\n" +
	"\x04addr\x18\x01 \x01(\tR\x04addr\"N\n" +
	"\x17RunRealiScannerResponse\x123\n" +
	"\aresults\x18\x01 \x03(\v2\x19.x.api.RealiScannerResultR\aresults\"<\n" +
	"\x12RealiScannerResult\x12\x0e\n" +
	"\x02ip\x18\x01 \x01(\tR\x02ip\x12\x16\n" +
	"\x06domain\x18\x02 \x01(\tR\x06domain\"\x1e\n" +
	"\x1cGenerateX25519KeyPairRequest\"C\n" +
	"\x1dGenerateX25519KeyPairResponse\x12\x10\n" +
	"\x03pub\x18\x01 \x01(\tR\x03pub\x12\x10\n" +
	"\x03pri\x18\x02 \x01(\tR\x03pri\"\xa6\x02\n" +
	"\x1aStartMacSystemProxyRequest\x12,\n" +
	"\x12http_proxy_address\x18\x01 \x01(\tR\x10httpProxyAddress\x12&\n" +
	"\x0fhttp_proxy_port\x18\x02 \x01(\rR\rhttpProxyPort\x12.\n" +
	"\x13https_proxy_address\x18\x03 \x01(\tR\x11httpsProxyAddress\x12(\n" +
	"\x10https_proxy_port\x18\x04 \x01(\rR\x0ehttpsProxyPort\x12.\n" +
	"\x13socks_proxy_address\x18\x05 \x01(\tR\x11socksProxyAddress\x12(\n" +
	"\x10socks_proxy_port\x18\x06 \x01(\rR\x0esocksProxyPort\"\x1b\n" +
	"\x19StopMacSystemProxyRequest\"\x10\n" +
	"\x0eCloseDbRequest\"#\n" +
	"\rOpenDbRequest\x12\x12\n" +
	"\x04path\x18\x01 \x01(\tR\x04path\"\xe0\x01\n" +
	"$InboundConfigToOutboundConfigRequest\x12/\n" +
	"\ainbound\x18\x01 \x01(\v2\x15.x.ProxyInboundConfigR\ainbound\x12?\n" +
	"\rmulti_inbound\x18\x02 \x01(\v2\x1a.x.MultiProxyInboundConfigR\fmultiInbound\x12%\n" +
	"\x0eserver_address\x18\x03 \x01(\tR\rserverAddress\x12\x1f\n" +
	"\vserver_name\x18\x04 \x01(\tR\n" +
	"serverName\"l\n" +
	"%InboundConfigToOutboundConfigResponse\x12C\n" +
	"\x10outbound_configs\x18\x01 \x03(\v2\x18.x.OutboundHandlerConfigR\x0foutboundConfigs\"S\n" +
	"\fToUrlRequest\x12C\n" +
	"\x10outbound_confogs\x18\x01 \x03(\v2\x18.x.OutboundHandlerConfigR\x0foutboundConfogs\"F\n" +
	"\rToUrlResponse\x12\x12\n" +
	"\x04urls\x18\x01 \x03(\tR\x04urls\x12!\n" +
	"\ffailed_nodes\x18\x02 \x03(\tR\vfailedNodes2\x81\x13\n" +
	"\x03Api\x12>\n" +
	"\x0eUpdateTmStatus\x12\x1c.x.api.UpdateTmStatusRequest\x1a\x0e.x.api.Receipt\x12;\n" +
	"\bDownload\x12\x16.x.api.DownloadRequest\x1a\x17.x.api.DownloadResponse\x12J\n" +
	"\rHandlerUsable\x12\x1b.x.api.HandlerUsableRequest\x1a\x1c.x.api.HandlerUsableResponse\x12@\n" +
	"\tSpeedTest\x12\x17.x.api.SpeedTestRequest\x1a\x18.x.api.SpeedTestResponse0\x01\x128\n" +
	"\aRttTest\x12\x15.x.api.RttTestRequest\x1a\x16.x.api.RttTestResponse\x122\n" +
	"\x05GeoIP\x12\x13.x.api.GeoIPRequest\x1a\x14.x.api.GeoIPResponse\x12Y\n" +
	"\x12GetServerPublicKey\x12 .x.api.GetServerPublicKeyRequest\x1a!.x.api.GetServerPublicKeyResponse\x12L\n" +
	"\rMonitorServer\x12\x1b.x.api.MonitorServerRequest\x1a\x1c.x.api.MonitorServerResponse0\x01\x12G\n" +
	"\fServerAction\x12\x1a.x.api.ServerActionRequest\x1a\x1b.x.api.ServerActionResponse\x12G\n" +
	"\fVproxyStatus\x12\x1a.x.api.VproxyStatusRequest\x1a\x1b.x.api.VproxyStatusResponse\x12&\n" +
	"\x02VX\x12\x10.x.api.VXRequest\x1a\x0e.x.api.Receipt\x12G\n" +
	"\fServerConfig\x12\x1a.x.api.ServerConfigRequest\x1a\x1b.x.api.ServerConfigResponse\x12Y\n" +
	"\x12UpdateServerConfig\x12 .x.api.UpdateServerConfigRequest\x1a!.x.api.UpdateServerConfigResponse\x12Y\n" +
	"\x12UpdateSubscription\x12 .x.api.UpdateSubscriptionRequest\x1a!.x.api.UpdateSubscriptionResponse\x12P\n" +
	"\x0fProcessGeoFiles\x12\x1d.x.api.ProcessGeoFilesRequest\x1a\x1e.x.api.ProcessGeoFilesResponse\x125\n" +
	"\x06Decode\x12\x14.x.api.DecodeRequest\x1a\x15.x.api.DecodeResponse\x125\n" +
	"\x06Deploy\x12\x14.x.api.DeployRequest\x1a\x15.x.api.DeployResponse\x12G\n" +
	"\fGenerateCert\x12\x1a.x.api.GenerateCertRequest\x1a\x1b.x.api.GenerateCertResponse\x12J\n" +
	"\rGetCertDomain\x12\x1b.x.api.GetCertDomainRequest\x1a\x1c.x.api.GetCertDomainResponse\x12A\n" +
	"\n" +
	"AddInbound\x12\x18.x.api.AddInboundRequest\x1a\x19.x.api.AddInboundResponse\x12>\n" +
	"\tUploadLog\x12\x17.x.api.UploadLogRequest\x1a\x18.x.api.UploadLogResponse\x12b\n" +
	"\x15DefaultNICHasGlobalV6\x12#.x.api.DefaultNICHasGlobalV6Request\x1a$.x.api.DefaultNICHasGlobalV6Response\x12Y\n" +
	"\x12ParseClashRuleFile\x12 .x.api.ParseClashRuleFileRequest\x1a!.x.api.ParseClashRuleFileResponse\x12Y\n" +
	"\x12ParseGeositeConfig\x12 .x.api.ParseGeositeConfigRequest\x1a!.x.api.ParseGeositeConfigResponse\x12S\n" +
	"\x10ParseGeoIPConfig\x12\x1e.x.api.ParseGeoIPConfigRequest\x1a\x1f.x.api.ParseGeoIPConfigResponse\x12P\n" +
	"\x0fRunRealiScanner\x12\x1d.x.api.RunRealiScannerRequest\x1a\x1e.x.api.RunRealiScannerResponse\x12b\n" +
	"\x15GenerateX25519KeyPair\x12#.x.api.GenerateX25519KeyPairRequest\x1a$.x.api.GenerateX25519KeyPairResponse\x12H\n" +
	"\x13StartMacSystemProxy\x12!.x.api.StartMacSystemProxyRequest\x1a\x0e.x.api.Receipt\x12F\n" +
	"\x12StopMacSystemProxy\x12 .x.api.StopMacSystemProxyRequest\x1a\x0e.x.api.Receipt\x120\n" +
	"\aCloseDb\x12\x15.x.api.CloseDbRequest\x1a\x0e.x.api.Receipt\x12.\n" +
	"\x06OpenDb\x12\x14.x.api.OpenDbRequest\x1a\x0e.x.api.Receipt\x12z\n" +
	"\x1dInboundConfigToOutboundConfig\x12+.x.api.InboundConfigToOutboundConfigRequest\x1a,.x.api.InboundConfigToOutboundConfigResponse\x122\n" +
	"\x05ToUrl\x12\x13.x.api.ToUrlRequest\x1a\x14.x.api.ToUrlResponseB&Z$github.com/5vnetwork/vx-core/app/apib\x06proto3"

var (
	file_app_api_api_proto_rawDescOnce sync.Once
	file_app_api_api_proto_rawDescData []byte
)

func file_app_api_api_proto_rawDescGZIP() []byte {
	file_app_api_api_proto_rawDescOnce.Do(func() {
		file_app_api_api_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_api_api_proto_rawDesc), len(file_app_api_api_proto_rawDesc)))
	})
	return file_app_api_api_proto_rawDescData
}

var file_app_api_api_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_app_api_api_proto_msgTypes = make([]protoimpl.MessageInfo, 75)
var file_app_api_api_proto_goTypes = []any{
	(XStatusChangeNotifyRequest_Status)(0),        // 0: x.api.XStatusChangeNotifyRequest.Status
	(ServerActionRequest_Action)(0),               // 1: x.api.ServerActionRequest.Action
	(*ApiServerConfig)(nil),                       // 2: x.api.ApiServerConfig
	(*XStatusChangeNotifyRequest)(nil),            // 3: x.api.XStatusChangeNotifyRequest
	(*XStatusChangeNotifyResponse)(nil),           // 4: x.api.XStatusChangeNotifyResponse
	(*SetSubscriptionIntervalRequest)(nil),        // 5: x.api.SetSubscriptionIntervalRequest
	(*SetSubscriptionIntervalResponse)(nil),       // 6: x.api.SetSubscriptionIntervalResponse
	(*UpdateSubscriptionRequest)(nil),             // 7: x.api.UpdateSubscriptionRequest
	(*UpdateSubscriptionResponse)(nil),            // 8: x.api.UpdateSubscriptionResponse
	(*SetTunNameRequest)(nil),                     // 9: x.api.SetTunNameRequest
	(*SetTunNameResponse)(nil),                    // 10: x.api.SetTunNameResponse
	(*DownloadRequest)(nil),                       // 11: x.api.DownloadRequest
	(*DownloadResponse)(nil),                      // 12: x.api.DownloadResponse
	(*HandlerIpRequest)(nil),                      // 13: x.api.HandlerIpRequest
	(*RttTestRequest)(nil),                        // 14: x.api.RttTestRequest
	(*RttTestResponse)(nil),                       // 15: x.api.RttTestResponse
	(*HandlerIpResponse)(nil),                     // 16: x.api.HandlerIpResponse
	(*HandlerUsableRequest)(nil),                  // 17: x.api.HandlerUsableRequest
	(*HandlerUsableResponse)(nil),                 // 18: x.api.HandlerUsableResponse
	(*SpeedTestRequest)(nil),                      // 19: x.api.SpeedTestRequest
	(*SpeedTestResponse)(nil),                     // 20: x.api.SpeedTestResponse
	(*GeoIPRequest)(nil),                          // 21: x.api.GeoIPRequest
	(*GeoIPResponse)(nil),                         // 22: x.api.GeoIPResponse
	(*ServerSshConfig)(nil),                       // 23: x.api.ServerSshConfig
	(*MonitorServerRequest)(nil),                  // 24: x.api.MonitorServerRequest
	(*MonitorServerResponse)(nil),                 // 25: x.api.MonitorServerResponse
	(*DeployRequest)(nil),                         // 26: x.api.DeployRequest
	(*DeployResponse)(nil),                        // 27: x.api.DeployResponse
	(*ServerActionRequest)(nil),                   // 28: x.api.ServerActionRequest
	(*ServerActionResponse)(nil),                  // 29: x.api.ServerActionResponse
	(*VproxyStatusRequest)(nil),                   // 30: x.api.VproxyStatusRequest
	(*VproxyStatusResponse)(nil),                  // 31: x.api.VproxyStatusResponse
	(*VXRequest)(nil),                             // 32: x.api.VXRequest
	(*ServerConfigRequest)(nil),                   // 33: x.api.ServerConfigRequest
	(*ServerConfigResponse)(nil),                  // 34: x.api.ServerConfigResponse
	(*UpdateServerConfigRequest)(nil),             // 35: x.api.UpdateServerConfigRequest
	(*UpdateServerConfigResponse)(nil),            // 36: x.api.UpdateServerConfigResponse
	(*ProcessGeoFilesRequest)(nil),                // 37: x.api.ProcessGeoFilesRequest
	(*ProcessGeoFilesResponse)(nil),               // 38: x.api.ProcessGeoFilesResponse
	(*DecodeRequest)(nil),                         // 39: x.api.DecodeRequest
	(*DecodeResponse)(nil),                        // 40: x.api.DecodeResponse
	(*GetServerPublicKeyRequest)(nil),             // 41: x.api.GetServerPublicKeyRequest
	(*GetServerPublicKeyResponse)(nil),            // 42: x.api.GetServerPublicKeyResponse
	(*GenerateCertRequest)(nil),                   // 43: x.api.GenerateCertRequest
	(*GenerateCertResponse)(nil),                  // 44: x.api.GenerateCertResponse
	(*GetCertDomainRequest)(nil),                  // 45: x.api.GetCertDomainRequest
	(*GetCertDomainResponse)(nil),                 // 46: x.api.GetCertDomainResponse
	(*AddInboundRequest)(nil),                     // 47: x.api.AddInboundRequest
	(*AddInboundResponse)(nil),                    // 48: x.api.AddInboundResponse
	(*UploadLogRequest)(nil),                      // 49: x.api.UploadLogRequest
	(*UploadLogResponse)(nil),                     // 50: x.api.UploadLogResponse
	(*DefaultNICHasGlobalV6Request)(nil),          // 51: x.api.DefaultNICHasGlobalV6Request
	(*DefaultNICHasGlobalV6Response)(nil),         // 52: x.api.DefaultNICHasGlobalV6Response
	(*UpdateTmStatusRequest)(nil),                 // 53: x.api.UpdateTmStatusRequest
	(*Receipt)(nil),                               // 54: x.api.Receipt
	(*ParseClashRuleFileRequest)(nil),             // 55: x.api.ParseClashRuleFileRequest
	(*ParseClashRuleFileResponse)(nil),            // 56: x.api.ParseClashRuleFileResponse
	(*ParseGeositeConfigRequest)(nil),             // 57: x.api.ParseGeositeConfigRequest
	(*ParseGeositeConfigResponse)(nil),            // 58: x.api.ParseGeositeConfigResponse
	(*ParseGeoIPConfigRequest)(nil),               // 59: x.api.ParseGeoIPConfigRequest
	(*ParseGeoIPConfigResponse)(nil),              // 60: x.api.ParseGeoIPConfigResponse
	(*RunRealiScannerRequest)(nil),                // 61: x.api.RunRealiScannerRequest
	(*RunRealiScannerResponse)(nil),               // 62: x.api.RunRealiScannerResponse
	(*RealiScannerResult)(nil),                    // 63: x.api.RealiScannerResult
	(*GenerateX25519KeyPairRequest)(nil),          // 64: x.api.GenerateX25519KeyPairRequest
	(*GenerateX25519KeyPairResponse)(nil),         // 65: x.api.GenerateX25519KeyPairResponse
	(*StartMacSystemProxyRequest)(nil),            // 66: x.api.StartMacSystemProxyRequest
	(*StopMacSystemProxyRequest)(nil),             // 67: x.api.StopMacSystemProxyRequest
	(*CloseDbRequest)(nil),                        // 68: x.api.CloseDbRequest
	(*OpenDbRequest)(nil),                         // 69: x.api.OpenDbRequest
	(*InboundConfigToOutboundConfigRequest)(nil),  // 70: x.api.InboundConfigToOutboundConfigRequest
	(*InboundConfigToOutboundConfigResponse)(nil), // 71: x.api.InboundConfigToOutboundConfigResponse
	(*ToUrlRequest)(nil),                          // 72: x.api.ToUrlRequest
	(*ToUrlResponse)(nil),                         // 73: x.api.ToUrlResponse
	nil,                                           // 74: x.api.UpdateSubscriptionResponse.ErrorReasonsEntry
	nil,                                           // 75: x.api.DownloadResponse.UsageEntry
	nil,                                           // 76: x.api.DeployRequest.FilesEntry
	(*configs.HandlerConfig)(nil),                 // 77: x.HandlerConfig
	(*configs.OutboundHandlerConfig)(nil),         // 78: x.OutboundHandlerConfig
	(*server.ServerConfig)(nil),                   // 79: x.ServerConfig
	(*configs.ProxyInboundConfig)(nil),            // 80: x.ProxyInboundConfig
	(*geo.Domain)(nil),                            // 81: x.common.geo.Domain
	(*geo.CIDR)(nil),                              // 82: x.common.geo.CIDR
	(*configs.AppId)(nil),                         // 83: x.AppId
	(*configs.GeositeConfig)(nil),                 // 84: x.GeositeConfig
	(*configs.GeoIPConfig)(nil),                   // 85: x.GeoIPConfig
	(*configs.MultiProxyInboundConfig)(nil),       // 86: x.MultiProxyInboundConfig
}
var file_app_api_api_proto_depIdxs = []int32{
	0,  // 0: x.api.XStatusChangeNotifyRequest.status:type_name -> x.api.XStatusChangeNotifyRequest.Status
	77, // 1: x.api.UpdateSubscriptionRequest.handlers:type_name -> x.HandlerConfig
	74, // 2: x.api.UpdateSubscriptionResponse.error_reasons:type_name -> x.api.UpdateSubscriptionResponse.ErrorReasonsEntry
	77, // 3: x.api.DownloadRequest.handlers:type_name -> x.HandlerConfig
	75, // 4: x.api.DownloadResponse.usage:type_name -> x.api.DownloadResponse.UsageEntry
	78, // 5: x.api.HandlerIpRequest.handler:type_name -> x.OutboundHandlerConfig
	77, // 6: x.api.HandlerUsableRequest.handler:type_name -> x.HandlerConfig
	77, // 7: x.api.SpeedTestRequest.handlers:type_name -> x.HandlerConfig
	23, // 8: x.api.MonitorServerRequest.ssh_config:type_name -> x.api.ServerSshConfig
	23, // 9: x.api.DeployRequest.ssh_config:type_name -> x.api.ServerSshConfig
	76, // 10: x.api.DeployRequest.files:type_name -> x.api.DeployRequest.FilesEntry
	79, // 11: x.api.DeployRequest.vx_config:type_name -> x.ServerConfig
	1,  // 12: x.api.ServerActionRequest.action:type_name -> x.api.ServerActionRequest.Action
	23, // 13: x.api.ServerActionRequest.ssh_config:type_name -> x.api.ServerSshConfig
	23, // 14: x.api.VproxyStatusRequest.ssh_config:type_name -> x.api.ServerSshConfig
	23, // 15: x.api.VXRequest.ssh_config:type_name -> x.api.ServerSshConfig
	23, // 16: x.api.ServerConfigRequest.ssh_config:type_name -> x.api.ServerSshConfig
	79, // 17: x.api.ServerConfigResponse.config:type_name -> x.ServerConfig
	23, // 18: x.api.UpdateServerConfigRequest.ssh_config:type_name -> x.api.ServerSshConfig
	79, // 19: x.api.UpdateServerConfigRequest.config:type_name -> x.ServerConfig
	78, // 20: x.api.DecodeResponse.handlers:type_name -> x.OutboundHandlerConfig
	23, // 21: x.api.GetServerPublicKeyRequest.ssh_config:type_name -> x.api.ServerSshConfig
	80, // 22: x.api.AddInboundRequest.inbound:type_name -> x.ProxyInboundConfig
	81, // 23: x.api.ParseClashRuleFileResponse.domains:type_name -> x.common.geo.Domain
	82, // 24: x.api.ParseClashRuleFileResponse.cidrs:type_name -> x.common.geo.CIDR
	83, // 25: x.api.ParseClashRuleFileResponse.app_ids:type_name -> x.AppId
	84, // 26: x.api.ParseGeositeConfigRequest.config:type_name -> x.GeositeConfig
	81, // 27: x.api.ParseGeositeConfigResponse.domains:type_name -> x.common.geo.Domain
	85, // 28: x.api.ParseGeoIPConfigRequest.config:type_name -> x.GeoIPConfig
	82, // 29: x.api.ParseGeoIPConfigResponse.cidrs:type_name -> x.common.geo.CIDR
	63, // 30: x.api.RunRealiScannerResponse.results:type_name -> x.api.RealiScannerResult
	80, // 31: x.api.InboundConfigToOutboundConfigRequest.inbound:type_name -> x.ProxyInboundConfig
	86, // 32: x.api.InboundConfigToOutboundConfigRequest.multi_inbound:type_name -> x.MultiProxyInboundConfig
	78, // 33: x.api.InboundConfigToOutboundConfigResponse.outbound_configs:type_name -> x.OutboundHandlerConfig
	78, // 34: x.api.ToUrlRequest.outbound_confogs:type_name -> x.OutboundHandlerConfig
	53, // 35: x.api.Api.UpdateTmStatus:input_type -> x.api.UpdateTmStatusRequest
	11, // 36: x.api.Api.Download:input_type -> x.api.DownloadRequest
	17, // 37: x.api.Api.HandlerUsable:input_type -> x.api.HandlerUsableRequest
	19, // 38: x.api.Api.SpeedTest:input_type -> x.api.SpeedTestRequest
	14, // 39: x.api.Api.RttTest:input_type -> x.api.RttTestRequest
	21, // 40: x.api.Api.GeoIP:input_type -> x.api.GeoIPRequest
	41, // 41: x.api.Api.GetServerPublicKey:input_type -> x.api.GetServerPublicKeyRequest
	24, // 42: x.api.Api.MonitorServer:input_type -> x.api.MonitorServerRequest
	28, // 43: x.api.Api.ServerAction:input_type -> x.api.ServerActionRequest
	30, // 44: x.api.Api.VproxyStatus:input_type -> x.api.VproxyStatusRequest
	32, // 45: x.api.Api.VX:input_type -> x.api.VXRequest
	33, // 46: x.api.Api.ServerConfig:input_type -> x.api.ServerConfigRequest
	35, // 47: x.api.Api.UpdateServerConfig:input_type -> x.api.UpdateServerConfigRequest
	7,  // 48: x.api.Api.UpdateSubscription:input_type -> x.api.UpdateSubscriptionRequest
	37, // 49: x.api.Api.ProcessGeoFiles:input_type -> x.api.ProcessGeoFilesRequest
	39, // 50: x.api.Api.Decode:input_type -> x.api.DecodeRequest
	26, // 51: x.api.Api.Deploy:input_type -> x.api.DeployRequest
	43, // 52: x.api.Api.GenerateCert:input_type -> x.api.GenerateCertRequest
	45, // 53: x.api.Api.GetCertDomain:input_type -> x.api.GetCertDomainRequest
	47, // 54: x.api.Api.AddInbound:input_type -> x.api.AddInboundRequest
	49, // 55: x.api.Api.UploadLog:input_type -> x.api.UploadLogRequest
	51, // 56: x.api.Api.DefaultNICHasGlobalV6:input_type -> x.api.DefaultNICHasGlobalV6Request
	55, // 57: x.api.Api.ParseClashRuleFile:input_type -> x.api.ParseClashRuleFileRequest
	57, // 58: x.api.Api.ParseGeositeConfig:input_type -> x.api.ParseGeositeConfigRequest
	59, // 59: x.api.Api.ParseGeoIPConfig:input_type -> x.api.ParseGeoIPConfigRequest
	61, // 60: x.api.Api.RunRealiScanner:input_type -> x.api.RunRealiScannerRequest
	64, // 61: x.api.Api.GenerateX25519KeyPair:input_type -> x.api.GenerateX25519KeyPairRequest
	66, // 62: x.api.Api.StartMacSystemProxy:input_type -> x.api.StartMacSystemProxyRequest
	67, // 63: x.api.Api.StopMacSystemProxy:input_type -> x.api.StopMacSystemProxyRequest
	68, // 64: x.api.Api.CloseDb:input_type -> x.api.CloseDbRequest
	69, // 65: x.api.Api.OpenDb:input_type -> x.api.OpenDbRequest
	70, // 66: x.api.Api.InboundConfigToOutboundConfig:input_type -> x.api.InboundConfigToOutboundConfigRequest
	72, // 67: x.api.Api.ToUrl:input_type -> x.api.ToUrlRequest
	54, // 68: x.api.Api.UpdateTmStatus:output_type -> x.api.Receipt
	12, // 69: x.api.Api.Download:output_type -> x.api.DownloadResponse
	18, // 70: x.api.Api.HandlerUsable:output_type -> x.api.HandlerUsableResponse
	20, // 71: x.api.Api.SpeedTest:output_type -> x.api.SpeedTestResponse
	15, // 72: x.api.Api.RttTest:output_type -> x.api.RttTestResponse
	22, // 73: x.api.Api.GeoIP:output_type -> x.api.GeoIPResponse
	42, // 74: x.api.Api.GetServerPublicKey:output_type -> x.api.GetServerPublicKeyResponse
	25, // 75: x.api.Api.MonitorServer:output_type -> x.api.MonitorServerResponse
	29, // 76: x.api.Api.ServerAction:output_type -> x.api.ServerActionResponse
	31, // 77: x.api.Api.VproxyStatus:output_type -> x.api.VproxyStatusResponse
	54, // 78: x.api.Api.VX:output_type -> x.api.Receipt
	34, // 79: x.api.Api.ServerConfig:output_type -> x.api.ServerConfigResponse
	36, // 80: x.api.Api.UpdateServerConfig:output_type -> x.api.UpdateServerConfigResponse
	8,  // 81: x.api.Api.UpdateSubscription:output_type -> x.api.UpdateSubscriptionResponse
	38, // 82: x.api.Api.ProcessGeoFiles:output_type -> x.api.ProcessGeoFilesResponse
	40, // 83: x.api.Api.Decode:output_type -> x.api.DecodeResponse
	27, // 84: x.api.Api.Deploy:output_type -> x.api.DeployResponse
	44, // 85: x.api.Api.GenerateCert:output_type -> x.api.GenerateCertResponse
	46, // 86: x.api.Api.GetCertDomain:output_type -> x.api.GetCertDomainResponse
	48, // 87: x.api.Api.AddInbound:output_type -> x.api.AddInboundResponse
	50, // 88: x.api.Api.UploadLog:output_type -> x.api.UploadLogResponse
	52, // 89: x.api.Api.DefaultNICHasGlobalV6:output_type -> x.api.DefaultNICHasGlobalV6Response
	56, // 90: x.api.Api.ParseClashRuleFile:output_type -> x.api.ParseClashRuleFileResponse
	58, // 91: x.api.Api.ParseGeositeConfig:output_type -> x.api.ParseGeositeConfigResponse
	60, // 92: x.api.Api.ParseGeoIPConfig:output_type -> x.api.ParseGeoIPConfigResponse
	62, // 93: x.api.Api.RunRealiScanner:output_type -> x.api.RunRealiScannerResponse
	65, // 94: x.api.Api.GenerateX25519KeyPair:output_type -> x.api.GenerateX25519KeyPairResponse
	54, // 95: x.api.Api.StartMacSystemProxy:output_type -> x.api.Receipt
	54, // 96: x.api.Api.StopMacSystemProxy:output_type -> x.api.Receipt
	54, // 97: x.api.Api.CloseDb:output_type -> x.api.Receipt
	54, // 98: x.api.Api.OpenDb:output_type -> x.api.Receipt
	71, // 99: x.api.Api.InboundConfigToOutboundConfig:output_type -> x.api.InboundConfigToOutboundConfigResponse
	73, // 100: x.api.Api.ToUrl:output_type -> x.api.ToUrlResponse
	68, // [68:101] is the sub-list for method output_type
	35, // [35:68] is the sub-list for method input_type
	35, // [35:35] is the sub-list for extension type_name
	35, // [35:35] is the sub-list for extension extendee
	0,  // [0:35] is the sub-list for field type_name
}

func init() { file_app_api_api_proto_init() }
func file_app_api_api_proto_init() {
	if File_app_api_api_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_api_api_proto_rawDesc), len(file_app_api_api_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   75,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_app_api_api_proto_goTypes,
		DependencyIndexes: file_app_api_api_proto_depIdxs,
		EnumInfos:         file_app_api_api_proto_enumTypes,
		MessageInfos:      file_app_api_api_proto_msgTypes,
	}.Build()
	File_app_api_api_proto = out.File
	file_app_api_api_proto_goTypes = nil
	file_app_api_api_proto_depIdxs = nil
}
