package clientgrpc

import (
	configs "github.com/5vnetwork/vx-core/app/configs"
	userlogger "github.com/5vnetwork/vx-core/app/userlogger"
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

type RttTestRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Addr          string                 `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	Port          uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RttTestRequest) Reset() {
	*x = RttTestRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RttTestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RttTestRequest) ProtoMessage() {}

func (x *RttTestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[0]
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
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{0}
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
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RttTestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RttTestResponse) ProtoMessage() {}

func (x *RttTestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[1]
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
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{1}
}

func (x *RttTestResponse) GetPing() uint32 {
	if x != nil {
		return x.Ping
	}
	return 0
}

type Receipt struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Receipt) Reset() {
	*x = Receipt{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Receipt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Receipt) ProtoMessage() {}

func (x *Receipt) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[2]
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
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{2}
}

type CommunicateRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CommunicateRequest) Reset() {
	*x = CommunicateRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CommunicateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommunicateRequest) ProtoMessage() {}

func (x *CommunicateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommunicateRequest.ProtoReflect.Descriptor instead.
func (*CommunicateRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{3}
}

type CommunicateMessage struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Message:
	//
	//	*CommunicateMessage_HandlerError
	//	*CommunicateMessage_SubscriptionUpdate
	//	*CommunicateMessage_HandlerBeingUsed
	//	*CommunicateMessage_HandlerUpdated
	Message       isCommunicateMessage_Message `protobuf_oneof:"message"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CommunicateMessage) Reset() {
	*x = CommunicateMessage{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CommunicateMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommunicateMessage) ProtoMessage() {}

func (x *CommunicateMessage) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommunicateMessage.ProtoReflect.Descriptor instead.
func (*CommunicateMessage) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{4}
}

func (x *CommunicateMessage) GetMessage() isCommunicateMessage_Message {
	if x != nil {
		return x.Message
	}
	return nil
}

func (x *CommunicateMessage) GetHandlerError() *HandlerError {
	if x != nil {
		if x, ok := x.Message.(*CommunicateMessage_HandlerError); ok {
			return x.HandlerError
		}
	}
	return nil
}

func (x *CommunicateMessage) GetSubscriptionUpdate() *SubscriptionUpdated {
	if x != nil {
		if x, ok := x.Message.(*CommunicateMessage_SubscriptionUpdate); ok {
			return x.SubscriptionUpdate
		}
	}
	return nil
}

func (x *CommunicateMessage) GetHandlerBeingUsed() *HandlerBeingUsed {
	if x != nil {
		if x, ok := x.Message.(*CommunicateMessage_HandlerBeingUsed); ok {
			return x.HandlerBeingUsed
		}
	}
	return nil
}

func (x *CommunicateMessage) GetHandlerUpdated() *HandlerUpdated {
	if x != nil {
		if x, ok := x.Message.(*CommunicateMessage_HandlerUpdated); ok {
			return x.HandlerUpdated
		}
	}
	return nil
}

type isCommunicateMessage_Message interface {
	isCommunicateMessage_Message()
}

type CommunicateMessage_HandlerError struct {
	HandlerError *HandlerError `protobuf:"bytes,1,opt,name=handler_error,json=handlerError,proto3,oneof"`
}

type CommunicateMessage_SubscriptionUpdate struct {
	SubscriptionUpdate *SubscriptionUpdated `protobuf:"bytes,2,opt,name=subscription_update,json=subscriptionUpdate,proto3,oneof"`
}

type CommunicateMessage_HandlerBeingUsed struct {
	HandlerBeingUsed *HandlerBeingUsed `protobuf:"bytes,3,opt,name=handler_being_used,json=handlerBeingUsed,proto3,oneof"`
}

type CommunicateMessage_HandlerUpdated struct {
	HandlerUpdated *HandlerUpdated `protobuf:"bytes,4,opt,name=handler_updated,json=handlerUpdated,proto3,oneof"`
}

func (*CommunicateMessage_HandlerError) isCommunicateMessage_Message() {}

func (*CommunicateMessage_SubscriptionUpdate) isCommunicateMessage_Message() {}

func (*CommunicateMessage_HandlerBeingUsed) isCommunicateMessage_Message() {}

func (*CommunicateMessage_HandlerUpdated) isCommunicateMessage_Message() {}

type HandlerError struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerError) Reset() {
	*x = HandlerError{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerError) ProtoMessage() {}

func (x *HandlerError) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerError.ProtoReflect.Descriptor instead.
func (*HandlerError) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{5}
}

func (x *HandlerError) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *HandlerError) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type HandlerBeingUsed struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag4          string                 `protobuf:"bytes,1,opt,name=tag4,proto3" json:"tag4,omitempty"`
	Tag6          string                 `protobuf:"bytes,2,opt,name=tag6,proto3" json:"tag6,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerBeingUsed) Reset() {
	*x = HandlerBeingUsed{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerBeingUsed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerBeingUsed) ProtoMessage() {}

func (x *HandlerBeingUsed) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerBeingUsed.ProtoReflect.Descriptor instead.
func (*HandlerBeingUsed) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{6}
}

func (x *HandlerBeingUsed) GetTag4() string {
	if x != nil {
		return x.Tag4
	}
	return ""
}

func (x *HandlerBeingUsed) GetTag6() string {
	if x != nil {
		return x.Tag6
	}
	return ""
}

type HandlerUpdated struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerUpdated) Reset() {
	*x = HandlerUpdated{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerUpdated) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerUpdated) ProtoMessage() {}

func (x *HandlerUpdated) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerUpdated.ProtoReflect.Descriptor instead.
func (*HandlerUpdated) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{7}
}

func (x *HandlerUpdated) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type SubscriptionUpdated struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubscriptionUpdated) Reset() {
	*x = SubscriptionUpdated{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubscriptionUpdated) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscriptionUpdated) ProtoMessage() {}

func (x *SubscriptionUpdated) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscriptionUpdated.ProtoReflect.Descriptor instead.
func (*SubscriptionUpdated) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{8}
}

// inbound
type AddInboundRequest struct {
	state         protoimpl.MessageState      `protogen:"open.v1"`
	HandlerConfig *configs.ProxyInboundConfig `protobuf:"bytes,1,opt,name=handler_config,json=handlerConfig,proto3" json:"handler_config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddInboundRequest) Reset() {
	*x = AddInboundRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddInboundRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddInboundRequest) ProtoMessage() {}

func (x *AddInboundRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[9]
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
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{9}
}

func (x *AddInboundRequest) GetHandlerConfig() *configs.ProxyInboundConfig {
	if x != nil {
		return x.HandlerConfig
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
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddInboundResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddInboundResponse) ProtoMessage() {}

func (x *AddInboundResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[10]
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
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{10}
}

type RemoveInboundRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RemoveInboundRequest) Reset() {
	*x = RemoveInboundRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveInboundRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveInboundRequest) ProtoMessage() {}

func (x *RemoveInboundRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveInboundRequest.ProtoReflect.Descriptor instead.
func (*RemoveInboundRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{11}
}

func (x *RemoveInboundRequest) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

type RemoveInboundResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RemoveInboundResponse) Reset() {
	*x = RemoveInboundResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveInboundResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveInboundResponse) ProtoMessage() {}

func (x *RemoveInboundResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveInboundResponse.ProtoReflect.Descriptor instead.
func (*RemoveInboundResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{12}
}

type OutboundStats struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Up    uint64                 `protobuf:"varint,1,opt,name=up,proto3" json:"up,omitempty"`
	Down  uint64                 `protobuf:"varint,2,opt,name=down,proto3" json:"down,omitempty"`
	// download
	Rate          uint64 `protobuf:"varint,3,opt,name=rate,proto3" json:"rate,omitempty"`
	Ping          uint64 `protobuf:"varint,4,opt,name=ping,proto3" json:"ping,omitempty"`
	Id            string `protobuf:"bytes,5,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OutboundStats) Reset() {
	*x = OutboundStats{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OutboundStats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OutboundStats) ProtoMessage() {}

func (x *OutboundStats) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OutboundStats.ProtoReflect.Descriptor instead.
func (*OutboundStats) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{13}
}

func (x *OutboundStats) GetUp() uint64 {
	if x != nil {
		return x.Up
	}
	return 0
}

func (x *OutboundStats) GetDown() uint64 {
	if x != nil {
		return x.Down
	}
	return 0
}

func (x *OutboundStats) GetRate() uint64 {
	if x != nil {
		return x.Rate
	}
	return 0
}

func (x *OutboundStats) GetPing() uint64 {
	if x != nil {
		return x.Ping
	}
	return 0
}

func (x *OutboundStats) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetStatsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Interval      uint32                 `protobuf:"varint,1,opt,name=interval,proto3" json:"interval,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetStatsRequest) Reset() {
	*x = GetStatsRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetStatsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStatsRequest) ProtoMessage() {}

func (x *GetStatsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStatsRequest.ProtoReflect.Descriptor instead.
func (*GetStatsRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{14}
}

func (x *GetStatsRequest) GetInterval() uint32 {
	if x != nil {
		return x.Interval
	}
	return 0
}

type StatsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Stats         []*OutboundStats       `protobuf:"bytes,1,rep,name=stats,proto3" json:"stats,omitempty"`
	Connections   int32                  `protobuf:"varint,2,opt,name=connections,proto3" json:"connections,omitempty"`
	Memory        uint64                 `protobuf:"varint,3,opt,name=memory,proto3" json:"memory,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StatsResponse) Reset() {
	*x = StatsResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatsResponse) ProtoMessage() {}

func (x *StatsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatsResponse.ProtoReflect.Descriptor instead.
func (*StatsResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{15}
}

func (x *StatsResponse) GetStats() []*OutboundStats {
	if x != nil {
		return x.Stats
	}
	return nil
}

func (x *StatsResponse) GetConnections() int32 {
	if x != nil {
		return x.Connections
	}
	return 0
}

func (x *StatsResponse) GetMemory() uint64 {
	if x != nil {
		return x.Memory
	}
	return 0
}

type SetOutboundHandlerSpeedRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Speed         int32                  `protobuf:"varint,2,opt,name=speed,proto3" json:"speed,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SetOutboundHandlerSpeedRequest) Reset() {
	*x = SetOutboundHandlerSpeedRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[16]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetOutboundHandlerSpeedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetOutboundHandlerSpeedRequest) ProtoMessage() {}

func (x *SetOutboundHandlerSpeedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[16]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetOutboundHandlerSpeedRequest.ProtoReflect.Descriptor instead.
func (*SetOutboundHandlerSpeedRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{16}
}

func (x *SetOutboundHandlerSpeedRequest) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *SetOutboundHandlerSpeedRequest) GetSpeed() int32 {
	if x != nil {
		return x.Speed
	}
	return 0
}

type SetOutboundHandlerSpeedResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SetOutboundHandlerSpeedResponse) Reset() {
	*x = SetOutboundHandlerSpeedResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[17]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetOutboundHandlerSpeedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetOutboundHandlerSpeedResponse) ProtoMessage() {}

func (x *SetOutboundHandlerSpeedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[17]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetOutboundHandlerSpeedResponse.ProtoReflect.Descriptor instead.
func (*SetOutboundHandlerSpeedResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{17}
}

// log
// message ChangeLogLevelRequest { x.Level level = 1; }
// message ChangeLogLevelResponse {}
// message LogStreamRequest {}
// message LogMessage { string message = 1; }
type UserLogStreamRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UserLogStreamRequest) Reset() {
	*x = UserLogStreamRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[18]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserLogStreamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserLogStreamRequest) ProtoMessage() {}

func (x *UserLogStreamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[18]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserLogStreamRequest.ProtoReflect.Descriptor instead.
func (*UserLogStreamRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{18}
}

type ToggleUserLogRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Enable        bool                   `protobuf:"varint,1,opt,name=enable,proto3" json:"enable,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ToggleUserLogRequest) Reset() {
	*x = ToggleUserLogRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[19]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToggleUserLogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToggleUserLogRequest) ProtoMessage() {}

func (x *ToggleUserLogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[19]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToggleUserLogRequest.ProtoReflect.Descriptor instead.
func (*ToggleUserLogRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{19}
}

func (x *ToggleUserLogRequest) GetEnable() bool {
	if x != nil {
		return x.Enable
	}
	return false
}

type ToggleUserLogResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ToggleUserLogResponse) Reset() {
	*x = ToggleUserLogResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[20]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToggleUserLogResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToggleUserLogResponse) ProtoMessage() {}

func (x *ToggleUserLogResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[20]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToggleUserLogResponse.ProtoReflect.Descriptor instead.
func (*ToggleUserLogResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{20}
}

type ToggleLogAppIdRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Enable        bool                   `protobuf:"varint,1,opt,name=enable,proto3" json:"enable,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ToggleLogAppIdRequest) Reset() {
	*x = ToggleLogAppIdRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[21]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToggleLogAppIdRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToggleLogAppIdRequest) ProtoMessage() {}

func (x *ToggleLogAppIdRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[21]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToggleLogAppIdRequest.ProtoReflect.Descriptor instead.
func (*ToggleLogAppIdRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{21}
}

func (x *ToggleLogAppIdRequest) GetEnable() bool {
	if x != nil {
		return x.Enable
	}
	return false
}

type ToggleLogAppIdResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ToggleLogAppIdResponse) Reset() {
	*x = ToggleLogAppIdResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[22]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToggleLogAppIdResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToggleLogAppIdResponse) ProtoMessage() {}

func (x *ToggleLogAppIdResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[22]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToggleLogAppIdResponse.ProtoReflect.Descriptor instead.
func (*ToggleLogAppIdResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{22}
}

// outbound
type ChangeOutboundRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// handlers to be added
	Handlers []*configs.HandlerConfig `protobuf:"bytes,1,rep,name=handlers,proto3" json:"handlers,omitempty"`
	// handlers to be removed
	Tags []string `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty"`
	// delete all proxy outbound handlers
	DeleteAll     bool `protobuf:"varint,3,opt,name=delete_all,json=deleteAll,proto3" json:"delete_all,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChangeOutboundRequest) Reset() {
	*x = ChangeOutboundRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[23]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChangeOutboundRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeOutboundRequest) ProtoMessage() {}

func (x *ChangeOutboundRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[23]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeOutboundRequest.ProtoReflect.Descriptor instead.
func (*ChangeOutboundRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{23}
}

func (x *ChangeOutboundRequest) GetHandlers() []*configs.HandlerConfig {
	if x != nil {
		return x.Handlers
	}
	return nil
}

func (x *ChangeOutboundRequest) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *ChangeOutboundRequest) GetDeleteAll() bool {
	if x != nil {
		return x.DeleteAll
	}
	return false
}

type ChangeOutboundResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChangeOutboundResponse) Reset() {
	*x = ChangeOutboundResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[24]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChangeOutboundResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeOutboundResponse) ProtoMessage() {}

func (x *ChangeOutboundResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[24]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeOutboundResponse.ProtoReflect.Descriptor instead.
func (*ChangeOutboundResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{24}
}

type CurrentOutboundRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CurrentOutboundRequest) Reset() {
	*x = CurrentOutboundRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[25]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CurrentOutboundRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CurrentOutboundRequest) ProtoMessage() {}

func (x *CurrentOutboundRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[25]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CurrentOutboundRequest.ProtoReflect.Descriptor instead.
func (*CurrentOutboundRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{25}
}

type CurrentOutboundResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	OutboundTags  []string               `protobuf:"bytes,1,rep,name=outbound_tags,json=outboundTags,proto3" json:"outbound_tags,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CurrentOutboundResponse) Reset() {
	*x = CurrentOutboundResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[26]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CurrentOutboundResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CurrentOutboundResponse) ProtoMessage() {}

func (x *CurrentOutboundResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[26]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CurrentOutboundResponse.ProtoReflect.Descriptor instead.
func (*CurrentOutboundResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{26}
}

func (x *CurrentOutboundResponse) GetOutboundTags() []string {
	if x != nil {
		return x.OutboundTags
	}
	return nil
}

// routing
type ChangeRoutingModeRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	RouterConfig  *configs.RouterConfig  `protobuf:"bytes,1,opt,name=router_config,json=routerConfig,proto3" json:"router_config,omitempty"`
	GeoConfig     *configs.GeoConfig     `protobuf:"bytes,2,opt,name=geo_config,json=geoConfig,proto3" json:"geo_config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChangeRoutingModeRequest) Reset() {
	*x = ChangeRoutingModeRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[27]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChangeRoutingModeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeRoutingModeRequest) ProtoMessage() {}

func (x *ChangeRoutingModeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[27]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeRoutingModeRequest.ProtoReflect.Descriptor instead.
func (*ChangeRoutingModeRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{27}
}

func (x *ChangeRoutingModeRequest) GetRouterConfig() *configs.RouterConfig {
	if x != nil {
		return x.RouterConfig
	}
	return nil
}

func (x *ChangeRoutingModeRequest) GetGeoConfig() *configs.GeoConfig {
	if x != nil {
		return x.GeoConfig
	}
	return nil
}

type ChangeRoutingModeResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChangeRoutingModeResponse) Reset() {
	*x = ChangeRoutingModeResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[28]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChangeRoutingModeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeRoutingModeResponse) ProtoMessage() {}

func (x *ChangeRoutingModeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[28]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeRoutingModeResponse.ProtoReflect.Descriptor instead.
func (*ChangeRoutingModeResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{28}
}

type ChangeSelectorRequest struct {
	state            protoimpl.MessageState    `protogen:"open.v1"`
	SelectorsToAdd   []*configs.SelectorConfig `protobuf:"bytes,1,rep,name=selectors_to_add,json=selectorsToAdd,proto3" json:"selectors_to_add,omitempty"`
	SelectorToRemove string                    `protobuf:"bytes,2,opt,name=selector_to_remove,json=selectorToRemove,proto3" json:"selector_to_remove,omitempty"`
	DeleteAll        bool                      `protobuf:"varint,3,opt,name=delete_all,json=deleteAll,proto3" json:"delete_all,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *ChangeSelectorRequest) Reset() {
	*x = ChangeSelectorRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[29]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChangeSelectorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeSelectorRequest) ProtoMessage() {}

func (x *ChangeSelectorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[29]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeSelectorRequest.ProtoReflect.Descriptor instead.
func (*ChangeSelectorRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{29}
}

func (x *ChangeSelectorRequest) GetSelectorsToAdd() []*configs.SelectorConfig {
	if x != nil {
		return x.SelectorsToAdd
	}
	return nil
}

func (x *ChangeSelectorRequest) GetSelectorToRemove() string {
	if x != nil {
		return x.SelectorToRemove
	}
	return ""
}

func (x *ChangeSelectorRequest) GetDeleteAll() bool {
	if x != nil {
		return x.DeleteAll
	}
	return false
}

type UpdateSelectorBalancerRequest struct {
	state           protoimpl.MessageState                 `protogen:"open.v1"`
	Tag             string                                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	BalanceStrategy configs.SelectorConfig_BalanceStrategy `protobuf:"varint,2,opt,name=balance_strategy,json=balanceStrategy,proto3,enum=x.SelectorConfig_BalanceStrategy" json:"balance_strategy,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *UpdateSelectorBalancerRequest) Reset() {
	*x = UpdateSelectorBalancerRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[30]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateSelectorBalancerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateSelectorBalancerRequest) ProtoMessage() {}

func (x *UpdateSelectorBalancerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[30]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateSelectorBalancerRequest.ProtoReflect.Descriptor instead.
func (*UpdateSelectorBalancerRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{30}
}

func (x *UpdateSelectorBalancerRequest) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *UpdateSelectorBalancerRequest) GetBalanceStrategy() configs.SelectorConfig_BalanceStrategy {
	if x != nil {
		return x.BalanceStrategy
	}
	return configs.SelectorConfig_BalanceStrategy(0)
}

type UpdateSelectorFilterRequest struct {
	state         protoimpl.MessageState         `protogen:"open.v1"`
	Tag           string                         `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Filter        *configs.SelectorConfig_Filter `protobuf:"bytes,2,opt,name=filter,proto3" json:"filter,omitempty"`
	SelectFromOm  bool                           `protobuf:"varint,3,opt,name=select_from_om,json=selectFromOm,proto3" json:"select_from_om,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateSelectorFilterRequest) Reset() {
	*x = UpdateSelectorFilterRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[31]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateSelectorFilterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateSelectorFilterRequest) ProtoMessage() {}

func (x *UpdateSelectorFilterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[31]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateSelectorFilterRequest.ProtoReflect.Descriptor instead.
func (*UpdateSelectorFilterRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{31}
}

func (x *UpdateSelectorFilterRequest) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *UpdateSelectorFilterRequest) GetFilter() *configs.SelectorConfig_Filter {
	if x != nil {
		return x.Filter
	}
	return nil
}

func (x *UpdateSelectorFilterRequest) GetSelectFromOm() bool {
	if x != nil {
		return x.SelectFromOm
	}
	return false
}

type ChangeSelectorResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChangeSelectorResponse) Reset() {
	*x = ChangeSelectorResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[32]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChangeSelectorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeSelectorResponse) ProtoMessage() {}

func (x *ChangeSelectorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[32]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeSelectorResponse.ProtoReflect.Descriptor instead.
func (*ChangeSelectorResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{32}
}

type HandlerChangeNotify struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerChangeNotify) Reset() {
	*x = HandlerChangeNotify{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[33]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerChangeNotify) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerChangeNotify) ProtoMessage() {}

func (x *HandlerChangeNotify) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[33]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerChangeNotify.ProtoReflect.Descriptor instead.
func (*HandlerChangeNotify) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{33}
}

type HandlerChangeNotifyResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerChangeNotifyResponse) Reset() {
	*x = HandlerChangeNotifyResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[34]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerChangeNotifyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerChangeNotifyResponse) ProtoMessage() {}

func (x *HandlerChangeNotifyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[34]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerChangeNotifyResponse.ProtoReflect.Descriptor instead.
func (*HandlerChangeNotifyResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{34}
}

// fake dns
type SwitchFakeDnsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Enable        bool                   `protobuf:"varint,1,opt,name=enable,proto3" json:"enable,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SwitchFakeDnsRequest) Reset() {
	*x = SwitchFakeDnsRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[35]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SwitchFakeDnsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SwitchFakeDnsRequest) ProtoMessage() {}

func (x *SwitchFakeDnsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[35]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SwitchFakeDnsRequest.ProtoReflect.Descriptor instead.
func (*SwitchFakeDnsRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{35}
}

func (x *SwitchFakeDnsRequest) GetEnable() bool {
	if x != nil {
		return x.Enable
	}
	return false
}

type SwitchFakeDnsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SwitchFakeDnsResponse) Reset() {
	*x = SwitchFakeDnsResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[36]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SwitchFakeDnsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SwitchFakeDnsResponse) ProtoMessage() {}

func (x *SwitchFakeDnsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[36]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SwitchFakeDnsResponse.ProtoReflect.Descriptor instead.
func (*SwitchFakeDnsResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{36}
}

// geo
type UpdateGeoRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Geo           *configs.GeoConfig     `protobuf:"bytes,1,opt,name=geo,proto3" json:"geo,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateGeoRequest) Reset() {
	*x = UpdateGeoRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[37]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateGeoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateGeoRequest) ProtoMessage() {}

func (x *UpdateGeoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[37]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateGeoRequest.ProtoReflect.Descriptor instead.
func (*UpdateGeoRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{37}
}

func (x *UpdateGeoRequest) GetGeo() *configs.GeoConfig {
	if x != nil {
		return x.Geo
	}
	return nil
}

type UpdateGeoResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateGeoResponse) Reset() {
	*x = UpdateGeoResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[38]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateGeoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateGeoResponse) ProtoMessage() {}

func (x *UpdateGeoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[38]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateGeoResponse.ProtoReflect.Descriptor instead.
func (*UpdateGeoResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{38}
}

type AddGeoDomainRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	DomainSetName string                 `protobuf:"bytes,1,opt,name=domain_set_name,json=domainSetName,proto3" json:"domain_set_name,omitempty"`
	Domain        *geo.Domain            `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddGeoDomainRequest) Reset() {
	*x = AddGeoDomainRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[39]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddGeoDomainRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddGeoDomainRequest) ProtoMessage() {}

func (x *AddGeoDomainRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[39]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddGeoDomainRequest.ProtoReflect.Descriptor instead.
func (*AddGeoDomainRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{39}
}

func (x *AddGeoDomainRequest) GetDomainSetName() string {
	if x != nil {
		return x.DomainSetName
	}
	return ""
}

func (x *AddGeoDomainRequest) GetDomain() *geo.Domain {
	if x != nil {
		return x.Domain
	}
	return nil
}

type RemoveGeoDomainRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	DomainSetName string                 `protobuf:"bytes,1,opt,name=domain_set_name,json=domainSetName,proto3" json:"domain_set_name,omitempty"`
	Domain        *geo.Domain            `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RemoveGeoDomainRequest) Reset() {
	*x = RemoveGeoDomainRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[40]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveGeoDomainRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveGeoDomainRequest) ProtoMessage() {}

func (x *RemoveGeoDomainRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[40]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveGeoDomainRequest.ProtoReflect.Descriptor instead.
func (*RemoveGeoDomainRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{40}
}

func (x *RemoveGeoDomainRequest) GetDomainSetName() string {
	if x != nil {
		return x.DomainSetName
	}
	return ""
}

func (x *RemoveGeoDomainRequest) GetDomain() *geo.Domain {
	if x != nil {
		return x.Domain
	}
	return nil
}

type ReplaceDomainSetRequest struct {
	state         protoimpl.MessageState         `protogen:"open.v1"`
	Set           *configs.AtomicDomainSetConfig `protobuf:"bytes,1,opt,name=set,proto3" json:"set,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReplaceDomainSetRequest) Reset() {
	*x = ReplaceDomainSetRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[41]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReplaceDomainSetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReplaceDomainSetRequest) ProtoMessage() {}

func (x *ReplaceDomainSetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[41]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReplaceDomainSetRequest.ProtoReflect.Descriptor instead.
func (*ReplaceDomainSetRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{41}
}

func (x *ReplaceDomainSetRequest) GetSet() *configs.AtomicDomainSetConfig {
	if x != nil {
		return x.Set
	}
	return nil
}

type ReplaceIPSetRequest struct {
	state         protoimpl.MessageState     `protogen:"open.v1"`
	Set           *configs.AtomicIPSetConfig `protobuf:"bytes,1,opt,name=set,proto3" json:"set,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReplaceIPSetRequest) Reset() {
	*x = ReplaceIPSetRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[42]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReplaceIPSetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReplaceIPSetRequest) ProtoMessage() {}

func (x *ReplaceIPSetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[42]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReplaceIPSetRequest.ProtoReflect.Descriptor instead.
func (*ReplaceIPSetRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{42}
}

func (x *ReplaceIPSetRequest) GetSet() *configs.AtomicIPSetConfig {
	if x != nil {
		return x.Set
	}
	return nil
}

// app id
type UpdateRouterRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	RouterConfig  *configs.RouterConfig  `protobuf:"bytes,1,opt,name=router_config,json=routerConfig,proto3" json:"router_config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateRouterRequest) Reset() {
	*x = UpdateRouterRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[43]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateRouterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRouterRequest) ProtoMessage() {}

func (x *UpdateRouterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[43]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRouterRequest.ProtoReflect.Descriptor instead.
func (*UpdateRouterRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{43}
}

func (x *UpdateRouterRequest) GetRouterConfig() *configs.RouterConfig {
	if x != nil {
		return x.RouterConfig
	}
	return nil
}

type UpdateRouterResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateRouterResponse) Reset() {
	*x = UpdateRouterResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[44]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateRouterResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRouterResponse) ProtoMessage() {}

func (x *UpdateRouterResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[44]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRouterResponse.ProtoReflect.Descriptor instead.
func (*UpdateRouterResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{44}
}

// subscription
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
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[45]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetSubscriptionIntervalRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetSubscriptionIntervalRequest) ProtoMessage() {}

func (x *SetSubscriptionIntervalRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[45]
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
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{45}
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
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[46]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetSubscriptionIntervalResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetSubscriptionIntervalResponse) ProtoMessage() {}

func (x *SetSubscriptionIntervalResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[46]
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
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{46}
}

type SetAutoSubscriptionUpdateRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Enable        bool                   `protobuf:"varint,1,opt,name=enable,proto3" json:"enable,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SetAutoSubscriptionUpdateRequest) Reset() {
	*x = SetAutoSubscriptionUpdateRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[47]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetAutoSubscriptionUpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetAutoSubscriptionUpdateRequest) ProtoMessage() {}

func (x *SetAutoSubscriptionUpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[47]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetAutoSubscriptionUpdateRequest.ProtoReflect.Descriptor instead.
func (*SetAutoSubscriptionUpdateRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{47}
}

func (x *SetAutoSubscriptionUpdateRequest) GetEnable() bool {
	if x != nil {
		return x.Enable
	}
	return false
}

type SetProxyShareRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Enable        bool                   `protobuf:"varint,1,opt,name=enable,proto3" json:"enable,omitempty"`
	ListenAddr    string                 `protobuf:"bytes,2,opt,name=listen_addr,json=listenAddr,proto3" json:"listen_addr,omitempty"`
	ListenPort    uint32                 `protobuf:"varint,3,opt,name=listen_port,json=listenPort,proto3" json:"listen_port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SetProxyShareRequest) Reset() {
	*x = SetProxyShareRequest{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[48]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetProxyShareRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetProxyShareRequest) ProtoMessage() {}

func (x *SetProxyShareRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[48]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetProxyShareRequest.ProtoReflect.Descriptor instead.
func (*SetProxyShareRequest) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{48}
}

func (x *SetProxyShareRequest) GetEnable() bool {
	if x != nil {
		return x.Enable
	}
	return false
}

func (x *SetProxyShareRequest) GetListenAddr() string {
	if x != nil {
		return x.ListenAddr
	}
	return ""
}

func (x *SetProxyShareRequest) GetListenPort() uint32 {
	if x != nil {
		return x.ListenPort
	}
	return 0
}

type SetProxyShareResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SetProxyShareResponse) Reset() {
	*x = SetProxyShareResponse{}
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[49]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SetProxyShareResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetProxyShareResponse) ProtoMessage() {}

func (x *SetProxyShareResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_clientgrpc_grpc_proto_msgTypes[49]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetProxyShareResponse.ProtoReflect.Descriptor instead.
func (*SetProxyShareResponse) Descriptor() ([]byte, []int) {
	return file_app_clientgrpc_grpc_proto_rawDescGZIP(), []int{49}
}

var File_app_clientgrpc_grpc_proto protoreflect.FileDescriptor

const file_app_clientgrpc_grpc_proto_rawDesc = "" +
	"\n" +
	"\x19app/clientgrpc/grpc.proto\x12\fx.clientgrpc\x1a\x14protos/inbound.proto\x1a\x15protos/outbound.proto\x1a\x13protos/router.proto\x1a\x10protos/geo.proto\x1a\x1bapp/userlogger/config.proto\x1a\x14common/geo/geo.proto\"8\n" +
	"\x0eRttTestRequest\x12\x12\n" +
	"\x04addr\x18\x01 \x01(\tR\x04addr\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\"%\n" +
	"\x0fRttTestResponse\x12\x12\n" +
	"\x04ping\x18\x01 \x01(\rR\x04ping\"\t\n" +
	"\aReceipt\"\x14\n" +
	"\x12CommunicateRequest\"\xd1\x02\n" +
	"\x12CommunicateMessage\x12A\n" +
	"\rhandler_error\x18\x01 \x01(\v2\x1a.x.clientgrpc.HandlerErrorH\x00R\fhandlerError\x12T\n" +
	"\x13subscription_update\x18\x02 \x01(\v2!.x.clientgrpc.SubscriptionUpdatedH\x00R\x12subscriptionUpdate\x12N\n" +
	"\x12handler_being_used\x18\x03 \x01(\v2\x1e.x.clientgrpc.HandlerBeingUsedH\x00R\x10handlerBeingUsed\x12G\n" +
	"\x0fhandler_updated\x18\x04 \x01(\v2\x1c.x.clientgrpc.HandlerUpdatedH\x00R\x0ehandlerUpdatedB\t\n" +
	"\amessage\"6\n" +
	"\fHandlerError\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error\":\n" +
	"\x10HandlerBeingUsed\x12\x12\n" +
	"\x04tag4\x18\x01 \x01(\tR\x04tag4\x12\x12\n" +
	"\x04tag6\x18\x02 \x01(\tR\x04tag6\" \n" +
	"\x0eHandlerUpdated\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\"\x15\n" +
	"\x13SubscriptionUpdated\"Q\n" +
	"\x11AddInboundRequest\x12<\n" +
	"\x0ehandler_config\x18\x01 \x01(\v2\x15.x.ProxyInboundConfigR\rhandlerConfig\"\x14\n" +
	"\x12AddInboundResponse\"(\n" +
	"\x14RemoveInboundRequest\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\"\x17\n" +
	"\x15RemoveInboundResponse\"k\n" +
	"\rOutboundStats\x12\x0e\n" +
	"\x02up\x18\x01 \x01(\x04R\x02up\x12\x12\n" +
	"\x04down\x18\x02 \x01(\x04R\x04down\x12\x12\n" +
	"\x04rate\x18\x03 \x01(\x04R\x04rate\x12\x12\n" +
	"\x04ping\x18\x04 \x01(\x04R\x04ping\x12\x0e\n" +
	"\x02id\x18\x05 \x01(\tR\x02id\"-\n" +
	"\x0fGetStatsRequest\x12\x1a\n" +
	"\binterval\x18\x01 \x01(\rR\binterval\"|\n" +
	"\rStatsResponse\x121\n" +
	"\x05stats\x18\x01 \x03(\v2\x1b.x.clientgrpc.OutboundStatsR\x05stats\x12 \n" +
	"\vconnections\x18\x02 \x01(\x05R\vconnections\x12\x16\n" +
	"\x06memory\x18\x03 \x01(\x04R\x06memory\"H\n" +
	"\x1eSetOutboundHandlerSpeedRequest\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12\x14\n" +
	"\x05speed\x18\x02 \x01(\x05R\x05speed\"!\n" +
	"\x1fSetOutboundHandlerSpeedResponse\"\x16\n" +
	"\x14UserLogStreamRequest\".\n" +
	"\x14ToggleUserLogRequest\x12\x16\n" +
	"\x06enable\x18\x01 \x01(\bR\x06enable\"\x17\n" +
	"\x15ToggleUserLogResponse\"/\n" +
	"\x15ToggleLogAppIdRequest\x12\x16\n" +
	"\x06enable\x18\x01 \x01(\bR\x06enable\"\x18\n" +
	"\x16ToggleLogAppIdResponse\"x\n" +
	"\x15ChangeOutboundRequest\x12,\n" +
	"\bhandlers\x18\x01 \x03(\v2\x10.x.HandlerConfigR\bhandlers\x12\x12\n" +
	"\x04tags\x18\x02 \x03(\tR\x04tags\x12\x1d\n" +
	"\n" +
	"delete_all\x18\x03 \x01(\bR\tdeleteAll\"\x18\n" +
	"\x16ChangeOutboundResponse\"\x18\n" +
	"\x16CurrentOutboundRequest\">\n" +
	"\x17CurrentOutboundResponse\x12#\n" +
	"\routbound_tags\x18\x01 \x03(\tR\foutboundTags\"}\n" +
	"\x18ChangeRoutingModeRequest\x124\n" +
	"\rrouter_config\x18\x01 \x01(\v2\x0f.x.RouterConfigR\frouterConfig\x12+\n" +
	"\n" +
	"geo_config\x18\x02 \x01(\v2\f.x.GeoConfigR\tgeoConfig\"\x1b\n" +
	"\x19ChangeRoutingModeResponse\"\xa1\x01\n" +
	"\x15ChangeSelectorRequest\x12;\n" +
	"\x10selectors_to_add\x18\x01 \x03(\v2\x11.x.SelectorConfigR\x0eselectorsToAdd\x12,\n" +
	"\x12selector_to_remove\x18\x02 \x01(\tR\x10selectorToRemove\x12\x1d\n" +
	"\n" +
	"delete_all\x18\x03 \x01(\bR\tdeleteAll\"\x7f\n" +
	"\x1dUpdateSelectorBalancerRequest\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12L\n" +
	"\x10balance_strategy\x18\x02 \x01(\x0e2!.x.SelectorConfig.BalanceStrategyR\x0fbalanceStrategy\"\x87\x01\n" +
	"\x1bUpdateSelectorFilterRequest\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x120\n" +
	"\x06filter\x18\x02 \x01(\v2\x18.x.SelectorConfig.FilterR\x06filter\x12$\n" +
	"\x0eselect_from_om\x18\x03 \x01(\bR\fselectFromOm\"\x18\n" +
	"\x16ChangeSelectorResponse\"\x15\n" +
	"\x13HandlerChangeNotify\"\x1d\n" +
	"\x1bHandlerChangeNotifyResponse\".\n" +
	"\x14SwitchFakeDnsRequest\x12\x16\n" +
	"\x06enable\x18\x01 \x01(\bR\x06enable\"\x17\n" +
	"\x15SwitchFakeDnsResponse\"2\n" +
	"\x10UpdateGeoRequest\x12\x1e\n" +
	"\x03geo\x18\x01 \x01(\v2\f.x.GeoConfigR\x03geo\"\x13\n" +
	"\x11UpdateGeoResponse\"k\n" +
	"\x13AddGeoDomainRequest\x12&\n" +
	"\x0fdomain_set_name\x18\x01 \x01(\tR\rdomainSetName\x12,\n" +
	"\x06domain\x18\x02 \x01(\v2\x14.x.common.geo.DomainR\x06domain\"n\n" +
	"\x16RemoveGeoDomainRequest\x12&\n" +
	"\x0fdomain_set_name\x18\x01 \x01(\tR\rdomainSetName\x12,\n" +
	"\x06domain\x18\x02 \x01(\v2\x14.x.common.geo.DomainR\x06domain\"E\n" +
	"\x17ReplaceDomainSetRequest\x12*\n" +
	"\x03set\x18\x01 \x01(\v2\x18.x.AtomicDomainSetConfigR\x03set\"=\n" +
	"\x13ReplaceIPSetRequest\x12&\n" +
	"\x03set\x18\x01 \x01(\v2\x14.x.AtomicIPSetConfigR\x03set\"K\n" +
	"\x13UpdateRouterRequest\x124\n" +
	"\rrouter_config\x18\x01 \x01(\v2\x0f.x.RouterConfigR\frouterConfig\"\x16\n" +
	"\x14UpdateRouterResponse\"<\n" +
	"\x1eSetSubscriptionIntervalRequest\x12\x1a\n" +
	"\binterval\x18\x01 \x01(\x05R\binterval\"!\n" +
	"\x1fSetSubscriptionIntervalResponse\":\n" +
	" SetAutoSubscriptionUpdateRequest\x12\x16\n" +
	"\x06enable\x18\x01 \x01(\bR\x06enable\"p\n" +
	"\x14SetProxyShareRequest\x12\x16\n" +
	"\x06enable\x18\x01 \x01(\bR\x06enable\x12\x1f\n" +
	"\vlisten_addr\x18\x02 \x01(\tR\n" +
	"listenAddr\x12\x1f\n" +
	"\vlisten_port\x18\x03 \x01(\rR\n" +
	"listenPort\"\x17\n" +
	"\x15SetProxyShareResponse2\xdd\x11\n" +
	"\rClientService\x12S\n" +
	"\vCommunicate\x12 .x.clientgrpc.CommunicateRequest\x1a .x.clientgrpc.CommunicateMessage0\x01\x12O\n" +
	"\n" +
	"AddInbound\x12\x1f.x.clientgrpc.AddInboundRequest\x1a .x.clientgrpc.AddInboundResponse\x12X\n" +
	"\rRemoveInbound\x12\".x.clientgrpc.RemoveInboundRequest\x1a#.x.clientgrpc.RemoveInboundResponse\x12P\n" +
	"\x0eGetStatsStream\x12\x1d.x.clientgrpc.GetStatsRequest\x1a\x1b.x.clientgrpc.StatsResponse\"\x000\x01\x12v\n" +
	"\x17SetOutboundHandlerSpeed\x12,.x.clientgrpc.SetOutboundHandlerSpeedRequest\x1a-.x.clientgrpc.SetOutboundHandlerSpeedResponse\x12S\n" +
	"\rUserLogStream\x12\".x.clientgrpc.UserLogStreamRequest\x1a\x1c.x.userlogger.UserLogMessage0\x01\x12X\n" +
	"\rToggleUserLog\x12\".x.clientgrpc.ToggleUserLogRequest\x1a#.x.clientgrpc.ToggleUserLogResponse\x12[\n" +
	"\x0eToggleLogAppId\x12#.x.clientgrpc.ToggleLogAppIdRequest\x1a$.x.clientgrpc.ToggleLogAppIdResponse\x12[\n" +
	"\x0eChangeOutbound\x12#.x.clientgrpc.ChangeOutboundRequest\x1a$.x.clientgrpc.ChangeOutboundResponse\x12^\n" +
	"\x0fCurrentOutbound\x12$.x.clientgrpc.CurrentOutboundRequest\x1a%.x.clientgrpc.CurrentOutboundResponse\x12d\n" +
	"\x11ChangeRoutingMode\x12&.x.clientgrpc.ChangeRoutingModeRequest\x1a'.x.clientgrpc.ChangeRoutingModeResponse\x12[\n" +
	"\x0eChangeSelector\x12#.x.clientgrpc.ChangeSelectorRequest\x1a$.x.clientgrpc.ChangeSelectorResponse\x12\\\n" +
	"\x16UpdateSelectorBalancer\x12+.x.clientgrpc.UpdateSelectorBalancerRequest\x1a\x15.x.clientgrpc.Receipt\x12X\n" +
	"\x14UpdateSelectorFilter\x12).x.clientgrpc.UpdateSelectorFilterRequest\x1a\x15.x.clientgrpc.Receipt\x12c\n" +
	"\x13NotifyHandlerChange\x12!.x.clientgrpc.HandlerChangeNotify\x1a).x.clientgrpc.HandlerChangeNotifyResponse\x12X\n" +
	"\rSwitchFakeDns\x12\".x.clientgrpc.SwitchFakeDnsRequest\x1a#.x.clientgrpc.SwitchFakeDnsResponse\x12L\n" +
	"\tUpdateGeo\x12\x1e.x.clientgrpc.UpdateGeoRequest\x1a\x1f.x.clientgrpc.UpdateGeoResponse\x12H\n" +
	"\fAddGeoDomain\x12!.x.clientgrpc.AddGeoDomainRequest\x1a\x15.x.clientgrpc.Receipt\x12N\n" +
	"\x0fRemoveGeoDomain\x12$.x.clientgrpc.RemoveGeoDomainRequest\x1a\x15.x.clientgrpc.Receipt\x12Q\n" +
	"\x11ReplaceGeoDomains\x12%.x.clientgrpc.ReplaceDomainSetRequest\x1a\x15.x.clientgrpc.Receipt\x12I\n" +
	"\rReplaceGeoIPs\x12!.x.clientgrpc.ReplaceIPSetRequest\x1a\x15.x.clientgrpc.Receipt\x12U\n" +
	"\fUpdateRouter\x12!.x.clientgrpc.UpdateRouterRequest\x1a\".x.clientgrpc.UpdateRouterResponse\x12v\n" +
	"\x17SetSubscriptionInterval\x12,.x.clientgrpc.SetSubscriptionIntervalRequest\x1a-.x.clientgrpc.SetSubscriptionIntervalResponse\x12b\n" +
	"\x19SetAutoSubscriptionUpdate\x12..x.clientgrpc.SetAutoSubscriptionUpdateRequest\x1a\x15.x.clientgrpc.Receipt\x12F\n" +
	"\aRttTest\x12\x1c.x.clientgrpc.RttTestRequest\x1a\x1d.x.clientgrpc.RttTestResponseB-Z+github.com/5vnetwork/vx-core/app/clientgrpcb\x06proto3"

var (
	file_app_clientgrpc_grpc_proto_rawDescOnce sync.Once
	file_app_clientgrpc_grpc_proto_rawDescData []byte
)

func file_app_clientgrpc_grpc_proto_rawDescGZIP() []byte {
	file_app_clientgrpc_grpc_proto_rawDescOnce.Do(func() {
		file_app_clientgrpc_grpc_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_clientgrpc_grpc_proto_rawDesc), len(file_app_clientgrpc_grpc_proto_rawDesc)))
	})
	return file_app_clientgrpc_grpc_proto_rawDescData
}

var file_app_clientgrpc_grpc_proto_msgTypes = make([]protoimpl.MessageInfo, 50)
var file_app_clientgrpc_grpc_proto_goTypes = []any{
	(*RttTestRequest)(nil),                      // 0: x.clientgrpc.RttTestRequest
	(*RttTestResponse)(nil),                     // 1: x.clientgrpc.RttTestResponse
	(*Receipt)(nil),                             // 2: x.clientgrpc.Receipt
	(*CommunicateRequest)(nil),                  // 3: x.clientgrpc.CommunicateRequest
	(*CommunicateMessage)(nil),                  // 4: x.clientgrpc.CommunicateMessage
	(*HandlerError)(nil),                        // 5: x.clientgrpc.HandlerError
	(*HandlerBeingUsed)(nil),                    // 6: x.clientgrpc.HandlerBeingUsed
	(*HandlerUpdated)(nil),                      // 7: x.clientgrpc.HandlerUpdated
	(*SubscriptionUpdated)(nil),                 // 8: x.clientgrpc.SubscriptionUpdated
	(*AddInboundRequest)(nil),                   // 9: x.clientgrpc.AddInboundRequest
	(*AddInboundResponse)(nil),                  // 10: x.clientgrpc.AddInboundResponse
	(*RemoveInboundRequest)(nil),                // 11: x.clientgrpc.RemoveInboundRequest
	(*RemoveInboundResponse)(nil),               // 12: x.clientgrpc.RemoveInboundResponse
	(*OutboundStats)(nil),                       // 13: x.clientgrpc.OutboundStats
	(*GetStatsRequest)(nil),                     // 14: x.clientgrpc.GetStatsRequest
	(*StatsResponse)(nil),                       // 15: x.clientgrpc.StatsResponse
	(*SetOutboundHandlerSpeedRequest)(nil),      // 16: x.clientgrpc.SetOutboundHandlerSpeedRequest
	(*SetOutboundHandlerSpeedResponse)(nil),     // 17: x.clientgrpc.SetOutboundHandlerSpeedResponse
	(*UserLogStreamRequest)(nil),                // 18: x.clientgrpc.UserLogStreamRequest
	(*ToggleUserLogRequest)(nil),                // 19: x.clientgrpc.ToggleUserLogRequest
	(*ToggleUserLogResponse)(nil),               // 20: x.clientgrpc.ToggleUserLogResponse
	(*ToggleLogAppIdRequest)(nil),               // 21: x.clientgrpc.ToggleLogAppIdRequest
	(*ToggleLogAppIdResponse)(nil),              // 22: x.clientgrpc.ToggleLogAppIdResponse
	(*ChangeOutboundRequest)(nil),               // 23: x.clientgrpc.ChangeOutboundRequest
	(*ChangeOutboundResponse)(nil),              // 24: x.clientgrpc.ChangeOutboundResponse
	(*CurrentOutboundRequest)(nil),              // 25: x.clientgrpc.CurrentOutboundRequest
	(*CurrentOutboundResponse)(nil),             // 26: x.clientgrpc.CurrentOutboundResponse
	(*ChangeRoutingModeRequest)(nil),            // 27: x.clientgrpc.ChangeRoutingModeRequest
	(*ChangeRoutingModeResponse)(nil),           // 28: x.clientgrpc.ChangeRoutingModeResponse
	(*ChangeSelectorRequest)(nil),               // 29: x.clientgrpc.ChangeSelectorRequest
	(*UpdateSelectorBalancerRequest)(nil),       // 30: x.clientgrpc.UpdateSelectorBalancerRequest
	(*UpdateSelectorFilterRequest)(nil),         // 31: x.clientgrpc.UpdateSelectorFilterRequest
	(*ChangeSelectorResponse)(nil),              // 32: x.clientgrpc.ChangeSelectorResponse
	(*HandlerChangeNotify)(nil),                 // 33: x.clientgrpc.HandlerChangeNotify
	(*HandlerChangeNotifyResponse)(nil),         // 34: x.clientgrpc.HandlerChangeNotifyResponse
	(*SwitchFakeDnsRequest)(nil),                // 35: x.clientgrpc.SwitchFakeDnsRequest
	(*SwitchFakeDnsResponse)(nil),               // 36: x.clientgrpc.SwitchFakeDnsResponse
	(*UpdateGeoRequest)(nil),                    // 37: x.clientgrpc.UpdateGeoRequest
	(*UpdateGeoResponse)(nil),                   // 38: x.clientgrpc.UpdateGeoResponse
	(*AddGeoDomainRequest)(nil),                 // 39: x.clientgrpc.AddGeoDomainRequest
	(*RemoveGeoDomainRequest)(nil),              // 40: x.clientgrpc.RemoveGeoDomainRequest
	(*ReplaceDomainSetRequest)(nil),             // 41: x.clientgrpc.ReplaceDomainSetRequest
	(*ReplaceIPSetRequest)(nil),                 // 42: x.clientgrpc.ReplaceIPSetRequest
	(*UpdateRouterRequest)(nil),                 // 43: x.clientgrpc.UpdateRouterRequest
	(*UpdateRouterResponse)(nil),                // 44: x.clientgrpc.UpdateRouterResponse
	(*SetSubscriptionIntervalRequest)(nil),      // 45: x.clientgrpc.SetSubscriptionIntervalRequest
	(*SetSubscriptionIntervalResponse)(nil),     // 46: x.clientgrpc.SetSubscriptionIntervalResponse
	(*SetAutoSubscriptionUpdateRequest)(nil),    // 47: x.clientgrpc.SetAutoSubscriptionUpdateRequest
	(*SetProxyShareRequest)(nil),                // 48: x.clientgrpc.SetProxyShareRequest
	(*SetProxyShareResponse)(nil),               // 49: x.clientgrpc.SetProxyShareResponse
	(*configs.ProxyInboundConfig)(nil),          // 50: x.ProxyInboundConfig
	(*configs.HandlerConfig)(nil),               // 51: x.HandlerConfig
	(*configs.RouterConfig)(nil),                // 52: x.RouterConfig
	(*configs.GeoConfig)(nil),                   // 53: x.GeoConfig
	(*configs.SelectorConfig)(nil),              // 54: x.SelectorConfig
	(configs.SelectorConfig_BalanceStrategy)(0), // 55: x.SelectorConfig.BalanceStrategy
	(*configs.SelectorConfig_Filter)(nil),       // 56: x.SelectorConfig.Filter
	(*geo.Domain)(nil),                          // 57: x.common.geo.Domain
	(*configs.AtomicDomainSetConfig)(nil),       // 58: x.AtomicDomainSetConfig
	(*configs.AtomicIPSetConfig)(nil),           // 59: x.AtomicIPSetConfig
	(*userlogger.UserLogMessage)(nil),           // 60: x.userlogger.UserLogMessage
}
var file_app_clientgrpc_grpc_proto_depIdxs = []int32{
	5,  // 0: x.clientgrpc.CommunicateMessage.handler_error:type_name -> x.clientgrpc.HandlerError
	8,  // 1: x.clientgrpc.CommunicateMessage.subscription_update:type_name -> x.clientgrpc.SubscriptionUpdated
	6,  // 2: x.clientgrpc.CommunicateMessage.handler_being_used:type_name -> x.clientgrpc.HandlerBeingUsed
	7,  // 3: x.clientgrpc.CommunicateMessage.handler_updated:type_name -> x.clientgrpc.HandlerUpdated
	50, // 4: x.clientgrpc.AddInboundRequest.handler_config:type_name -> x.ProxyInboundConfig
	13, // 5: x.clientgrpc.StatsResponse.stats:type_name -> x.clientgrpc.OutboundStats
	51, // 6: x.clientgrpc.ChangeOutboundRequest.handlers:type_name -> x.HandlerConfig
	52, // 7: x.clientgrpc.ChangeRoutingModeRequest.router_config:type_name -> x.RouterConfig
	53, // 8: x.clientgrpc.ChangeRoutingModeRequest.geo_config:type_name -> x.GeoConfig
	54, // 9: x.clientgrpc.ChangeSelectorRequest.selectors_to_add:type_name -> x.SelectorConfig
	55, // 10: x.clientgrpc.UpdateSelectorBalancerRequest.balance_strategy:type_name -> x.SelectorConfig.BalanceStrategy
	56, // 11: x.clientgrpc.UpdateSelectorFilterRequest.filter:type_name -> x.SelectorConfig.Filter
	53, // 12: x.clientgrpc.UpdateGeoRequest.geo:type_name -> x.GeoConfig
	57, // 13: x.clientgrpc.AddGeoDomainRequest.domain:type_name -> x.common.geo.Domain
	57, // 14: x.clientgrpc.RemoveGeoDomainRequest.domain:type_name -> x.common.geo.Domain
	58, // 15: x.clientgrpc.ReplaceDomainSetRequest.set:type_name -> x.AtomicDomainSetConfig
	59, // 16: x.clientgrpc.ReplaceIPSetRequest.set:type_name -> x.AtomicIPSetConfig
	52, // 17: x.clientgrpc.UpdateRouterRequest.router_config:type_name -> x.RouterConfig
	3,  // 18: x.clientgrpc.ClientService.Communicate:input_type -> x.clientgrpc.CommunicateRequest
	9,  // 19: x.clientgrpc.ClientService.AddInbound:input_type -> x.clientgrpc.AddInboundRequest
	11, // 20: x.clientgrpc.ClientService.RemoveInbound:input_type -> x.clientgrpc.RemoveInboundRequest
	14, // 21: x.clientgrpc.ClientService.GetStatsStream:input_type -> x.clientgrpc.GetStatsRequest
	16, // 22: x.clientgrpc.ClientService.SetOutboundHandlerSpeed:input_type -> x.clientgrpc.SetOutboundHandlerSpeedRequest
	18, // 23: x.clientgrpc.ClientService.UserLogStream:input_type -> x.clientgrpc.UserLogStreamRequest
	19, // 24: x.clientgrpc.ClientService.ToggleUserLog:input_type -> x.clientgrpc.ToggleUserLogRequest
	21, // 25: x.clientgrpc.ClientService.ToggleLogAppId:input_type -> x.clientgrpc.ToggleLogAppIdRequest
	23, // 26: x.clientgrpc.ClientService.ChangeOutbound:input_type -> x.clientgrpc.ChangeOutboundRequest
	25, // 27: x.clientgrpc.ClientService.CurrentOutbound:input_type -> x.clientgrpc.CurrentOutboundRequest
	27, // 28: x.clientgrpc.ClientService.ChangeRoutingMode:input_type -> x.clientgrpc.ChangeRoutingModeRequest
	29, // 29: x.clientgrpc.ClientService.ChangeSelector:input_type -> x.clientgrpc.ChangeSelectorRequest
	30, // 30: x.clientgrpc.ClientService.UpdateSelectorBalancer:input_type -> x.clientgrpc.UpdateSelectorBalancerRequest
	31, // 31: x.clientgrpc.ClientService.UpdateSelectorFilter:input_type -> x.clientgrpc.UpdateSelectorFilterRequest
	33, // 32: x.clientgrpc.ClientService.NotifyHandlerChange:input_type -> x.clientgrpc.HandlerChangeNotify
	35, // 33: x.clientgrpc.ClientService.SwitchFakeDns:input_type -> x.clientgrpc.SwitchFakeDnsRequest
	37, // 34: x.clientgrpc.ClientService.UpdateGeo:input_type -> x.clientgrpc.UpdateGeoRequest
	39, // 35: x.clientgrpc.ClientService.AddGeoDomain:input_type -> x.clientgrpc.AddGeoDomainRequest
	40, // 36: x.clientgrpc.ClientService.RemoveGeoDomain:input_type -> x.clientgrpc.RemoveGeoDomainRequest
	41, // 37: x.clientgrpc.ClientService.ReplaceGeoDomains:input_type -> x.clientgrpc.ReplaceDomainSetRequest
	42, // 38: x.clientgrpc.ClientService.ReplaceGeoIPs:input_type -> x.clientgrpc.ReplaceIPSetRequest
	43, // 39: x.clientgrpc.ClientService.UpdateRouter:input_type -> x.clientgrpc.UpdateRouterRequest
	45, // 40: x.clientgrpc.ClientService.SetSubscriptionInterval:input_type -> x.clientgrpc.SetSubscriptionIntervalRequest
	47, // 41: x.clientgrpc.ClientService.SetAutoSubscriptionUpdate:input_type -> x.clientgrpc.SetAutoSubscriptionUpdateRequest
	0,  // 42: x.clientgrpc.ClientService.RttTest:input_type -> x.clientgrpc.RttTestRequest
	4,  // 43: x.clientgrpc.ClientService.Communicate:output_type -> x.clientgrpc.CommunicateMessage
	10, // 44: x.clientgrpc.ClientService.AddInbound:output_type -> x.clientgrpc.AddInboundResponse
	12, // 45: x.clientgrpc.ClientService.RemoveInbound:output_type -> x.clientgrpc.RemoveInboundResponse
	15, // 46: x.clientgrpc.ClientService.GetStatsStream:output_type -> x.clientgrpc.StatsResponse
	17, // 47: x.clientgrpc.ClientService.SetOutboundHandlerSpeed:output_type -> x.clientgrpc.SetOutboundHandlerSpeedResponse
	60, // 48: x.clientgrpc.ClientService.UserLogStream:output_type -> x.userlogger.UserLogMessage
	20, // 49: x.clientgrpc.ClientService.ToggleUserLog:output_type -> x.clientgrpc.ToggleUserLogResponse
	22, // 50: x.clientgrpc.ClientService.ToggleLogAppId:output_type -> x.clientgrpc.ToggleLogAppIdResponse
	24, // 51: x.clientgrpc.ClientService.ChangeOutbound:output_type -> x.clientgrpc.ChangeOutboundResponse
	26, // 52: x.clientgrpc.ClientService.CurrentOutbound:output_type -> x.clientgrpc.CurrentOutboundResponse
	28, // 53: x.clientgrpc.ClientService.ChangeRoutingMode:output_type -> x.clientgrpc.ChangeRoutingModeResponse
	32, // 54: x.clientgrpc.ClientService.ChangeSelector:output_type -> x.clientgrpc.ChangeSelectorResponse
	2,  // 55: x.clientgrpc.ClientService.UpdateSelectorBalancer:output_type -> x.clientgrpc.Receipt
	2,  // 56: x.clientgrpc.ClientService.UpdateSelectorFilter:output_type -> x.clientgrpc.Receipt
	34, // 57: x.clientgrpc.ClientService.NotifyHandlerChange:output_type -> x.clientgrpc.HandlerChangeNotifyResponse
	36, // 58: x.clientgrpc.ClientService.SwitchFakeDns:output_type -> x.clientgrpc.SwitchFakeDnsResponse
	38, // 59: x.clientgrpc.ClientService.UpdateGeo:output_type -> x.clientgrpc.UpdateGeoResponse
	2,  // 60: x.clientgrpc.ClientService.AddGeoDomain:output_type -> x.clientgrpc.Receipt
	2,  // 61: x.clientgrpc.ClientService.RemoveGeoDomain:output_type -> x.clientgrpc.Receipt
	2,  // 62: x.clientgrpc.ClientService.ReplaceGeoDomains:output_type -> x.clientgrpc.Receipt
	2,  // 63: x.clientgrpc.ClientService.ReplaceGeoIPs:output_type -> x.clientgrpc.Receipt
	44, // 64: x.clientgrpc.ClientService.UpdateRouter:output_type -> x.clientgrpc.UpdateRouterResponse
	46, // 65: x.clientgrpc.ClientService.SetSubscriptionInterval:output_type -> x.clientgrpc.SetSubscriptionIntervalResponse
	2,  // 66: x.clientgrpc.ClientService.SetAutoSubscriptionUpdate:output_type -> x.clientgrpc.Receipt
	1,  // 67: x.clientgrpc.ClientService.RttTest:output_type -> x.clientgrpc.RttTestResponse
	43, // [43:68] is the sub-list for method output_type
	18, // [18:43] is the sub-list for method input_type
	18, // [18:18] is the sub-list for extension type_name
	18, // [18:18] is the sub-list for extension extendee
	0,  // [0:18] is the sub-list for field type_name
}

func init() { file_app_clientgrpc_grpc_proto_init() }
func file_app_clientgrpc_grpc_proto_init() {
	if File_app_clientgrpc_grpc_proto != nil {
		return
	}
	file_app_clientgrpc_grpc_proto_msgTypes[4].OneofWrappers = []any{
		(*CommunicateMessage_HandlerError)(nil),
		(*CommunicateMessage_SubscriptionUpdate)(nil),
		(*CommunicateMessage_HandlerBeingUsed)(nil),
		(*CommunicateMessage_HandlerUpdated)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_clientgrpc_grpc_proto_rawDesc), len(file_app_clientgrpc_grpc_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   50,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_app_clientgrpc_grpc_proto_goTypes,
		DependencyIndexes: file_app_clientgrpc_grpc_proto_depIdxs,
		MessageInfos:      file_app_clientgrpc_grpc_proto_msgTypes,
	}.Build()
	File_app_clientgrpc_grpc_proto = out.File
	file_app_clientgrpc_grpc_proto_goTypes = nil
	file_app_clientgrpc_grpc_proto_depIdxs = nil
}
