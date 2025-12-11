package userlogger

import (
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

type UserLogMessage struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Message:
	//
	//	*UserLogMessage_RouteMessage
	//	*UserLogMessage_ErrorMessage
	//	*UserLogMessage_SessionError
	//	*UserLogMessage_RejectMessage
	//	*UserLogMessage_Fallback
	Message       isUserLogMessage_Message `protobuf_oneof:"message"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UserLogMessage) Reset() {
	*x = UserLogMessage{}
	mi := &file_app_userlogger_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserLogMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserLogMessage) ProtoMessage() {}

func (x *UserLogMessage) ProtoReflect() protoreflect.Message {
	mi := &file_app_userlogger_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserLogMessage.ProtoReflect.Descriptor instead.
func (*UserLogMessage) Descriptor() ([]byte, []int) {
	return file_app_userlogger_config_proto_rawDescGZIP(), []int{0}
}

func (x *UserLogMessage) GetMessage() isUserLogMessage_Message {
	if x != nil {
		return x.Message
	}
	return nil
}

func (x *UserLogMessage) GetRouteMessage() *RouteMessage {
	if x != nil {
		if x, ok := x.Message.(*UserLogMessage_RouteMessage); ok {
			return x.RouteMessage
		}
	}
	return nil
}

func (x *UserLogMessage) GetErrorMessage() *ErrorMessage {
	if x != nil {
		if x, ok := x.Message.(*UserLogMessage_ErrorMessage); ok {
			return x.ErrorMessage
		}
	}
	return nil
}

func (x *UserLogMessage) GetSessionError() *SessionError {
	if x != nil {
		if x, ok := x.Message.(*UserLogMessage_SessionError); ok {
			return x.SessionError
		}
	}
	return nil
}

func (x *UserLogMessage) GetRejectMessage() *RejectMessage {
	if x != nil {
		if x, ok := x.Message.(*UserLogMessage_RejectMessage); ok {
			return x.RejectMessage
		}
	}
	return nil
}

func (x *UserLogMessage) GetFallback() *Fallback {
	if x != nil {
		if x, ok := x.Message.(*UserLogMessage_Fallback); ok {
			return x.Fallback
		}
	}
	return nil
}

type isUserLogMessage_Message interface {
	isUserLogMessage_Message()
}

type UserLogMessage_RouteMessage struct {
	RouteMessage *RouteMessage `protobuf:"bytes,1,opt,name=route_message,json=routeMessage,proto3,oneof"`
}

type UserLogMessage_ErrorMessage struct {
	ErrorMessage *ErrorMessage `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3,oneof"`
}

type UserLogMessage_SessionError struct {
	SessionError *SessionError `protobuf:"bytes,3,opt,name=session_error,json=sessionError,proto3,oneof"`
}

type UserLogMessage_RejectMessage struct {
	RejectMessage *RejectMessage `protobuf:"bytes,4,opt,name=reject_message,json=rejectMessage,proto3,oneof"`
}

type UserLogMessage_Fallback struct {
	Fallback *Fallback `protobuf:"bytes,5,opt,name=fallback,proto3,oneof"`
}

func (*UserLogMessage_RouteMessage) isUserLogMessage_Message() {}

func (*UserLogMessage_ErrorMessage) isUserLogMessage_Message() {}

func (*UserLogMessage_SessionError) isUserLogMessage_Message() {}

func (*UserLogMessage_RejectMessage) isUserLogMessage_Message() {}

func (*UserLogMessage_Fallback) isUserLogMessage_Message() {}

type RouteMessage struct {
	state     protoimpl.MessageState `protogen:"open.v1"`
	Dst       string                 `protobuf:"bytes,1,opt,name=dst,proto3" json:"dst,omitempty"`
	Tag       string                 `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	Timestamp int64                  `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	AppId     string                 `protobuf:"bytes,4,opt,name=app_id,json=appId,proto3" json:"app_id,omitempty"`
	// when dst is ip, this field contains
	// sniffed domain
	SniffDomain   string `protobuf:"bytes,5,opt,name=sniff_domain,json=sniffDomain,proto3" json:"sniff_domain,omitempty"`
	Sid           uint32 `protobuf:"varint,6,opt,name=sid,proto3" json:"sid,omitempty"`
	IpToDomain    string `protobuf:"bytes,7,opt,name=ip_to_domain,json=ipToDomain,proto3" json:"ip_to_domain,omitempty"`
	SelectorTag   string `protobuf:"bytes,8,opt,name=selector_tag,json=selectorTag,proto3" json:"selector_tag,omitempty"`
	MatchedRule   string `protobuf:"bytes,9,opt,name=matched_rule,json=matchedRule,proto3" json:"matched_rule,omitempty"`
	InboundTag    string `protobuf:"bytes,10,opt,name=inbound_tag,json=inboundTag,proto3" json:"inbound_tag,omitempty"`
	Network       string `protobuf:"bytes,11,opt,name=network,proto3" json:"network,omitempty"`
	SniffProtofol string `protobuf:"bytes,12,opt,name=sniff_protofol,json=sniffProtofol,proto3" json:"sniff_protofol,omitempty"`
	Source        string `protobuf:"bytes,13,opt,name=source,proto3" json:"source,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RouteMessage) Reset() {
	*x = RouteMessage{}
	mi := &file_app_userlogger_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RouteMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RouteMessage) ProtoMessage() {}

func (x *RouteMessage) ProtoReflect() protoreflect.Message {
	mi := &file_app_userlogger_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RouteMessage.ProtoReflect.Descriptor instead.
func (*RouteMessage) Descriptor() ([]byte, []int) {
	return file_app_userlogger_config_proto_rawDescGZIP(), []int{1}
}

func (x *RouteMessage) GetDst() string {
	if x != nil {
		return x.Dst
	}
	return ""
}

func (x *RouteMessage) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *RouteMessage) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *RouteMessage) GetAppId() string {
	if x != nil {
		return x.AppId
	}
	return ""
}

func (x *RouteMessage) GetSniffDomain() string {
	if x != nil {
		return x.SniffDomain
	}
	return ""
}

func (x *RouteMessage) GetSid() uint32 {
	if x != nil {
		return x.Sid
	}
	return 0
}

func (x *RouteMessage) GetIpToDomain() string {
	if x != nil {
		return x.IpToDomain
	}
	return ""
}

func (x *RouteMessage) GetSelectorTag() string {
	if x != nil {
		return x.SelectorTag
	}
	return ""
}

func (x *RouteMessage) GetMatchedRule() string {
	if x != nil {
		return x.MatchedRule
	}
	return ""
}

func (x *RouteMessage) GetInboundTag() string {
	if x != nil {
		return x.InboundTag
	}
	return ""
}

func (x *RouteMessage) GetNetwork() string {
	if x != nil {
		return x.Network
	}
	return ""
}

func (x *RouteMessage) GetSniffProtofol() string {
	if x != nil {
		return x.SniffProtofol
	}
	return ""
}

func (x *RouteMessage) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

type ErrorMessage struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Timestamp     int64                  `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ErrorMessage) Reset() {
	*x = ErrorMessage{}
	mi := &file_app_userlogger_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ErrorMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ErrorMessage) ProtoMessage() {}

func (x *ErrorMessage) ProtoReflect() protoreflect.Message {
	mi := &file_app_userlogger_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ErrorMessage.ProtoReflect.Descriptor instead.
func (*ErrorMessage) Descriptor() ([]byte, []int) {
	return file_app_userlogger_config_proto_rawDescGZIP(), []int{2}
}

func (x *ErrorMessage) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *ErrorMessage) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type Fallback struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Sid   uint32                 `protobuf:"varint,1,opt,name=sid,proto3" json:"sid,omitempty"`
	// the handler that eventually handle this session
	Tag           string `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Fallback) Reset() {
	*x = Fallback{}
	mi := &file_app_userlogger_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Fallback) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fallback) ProtoMessage() {}

func (x *Fallback) ProtoReflect() protoreflect.Message {
	mi := &file_app_userlogger_config_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fallback.ProtoReflect.Descriptor instead.
func (*Fallback) Descriptor() ([]byte, []int) {
	return file_app_userlogger_config_proto_rawDescGZIP(), []int{3}
}

func (x *Fallback) GetSid() uint32 {
	if x != nil {
		return x.Sid
	}
	return 0
}

func (x *Fallback) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

type SessionError struct {
	state   protoimpl.MessageState `protogen:"open.v1"`
	Message string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	// int64 timestamp = 2;
	Sid  uint32 `protobuf:"varint,3,opt,name=sid,proto3" json:"sid,omitempty"`
	Up   uint32 `protobuf:"varint,4,opt,name=up,proto3" json:"up,omitempty"`
	Down uint32 `protobuf:"varint,5,opt,name=down,proto3" json:"down,omitempty"`
	// if the session dst is ip, this is the dns server
	// handles the dns query that resolves to the ip
	Dns           string `protobuf:"bytes,6,opt,name=dns,proto3" json:"dns,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SessionError) Reset() {
	*x = SessionError{}
	mi := &file_app_userlogger_config_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SessionError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SessionError) ProtoMessage() {}

func (x *SessionError) ProtoReflect() protoreflect.Message {
	mi := &file_app_userlogger_config_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SessionError.ProtoReflect.Descriptor instead.
func (*SessionError) Descriptor() ([]byte, []int) {
	return file_app_userlogger_config_proto_rawDescGZIP(), []int{4}
}

func (x *SessionError) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *SessionError) GetSid() uint32 {
	if x != nil {
		return x.Sid
	}
	return 0
}

func (x *SessionError) GetUp() uint32 {
	if x != nil {
		return x.Up
	}
	return 0
}

func (x *SessionError) GetDown() uint32 {
	if x != nil {
		return x.Down
	}
	return 0
}

func (x *SessionError) GetDns() string {
	if x != nil {
		return x.Dns
	}
	return ""
}

type RejectMessage struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Dst           string                 `protobuf:"bytes,1,opt,name=dst,proto3" json:"dst,omitempty"`
	Timestamp     int64                  `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Domain        string                 `protobuf:"bytes,3,opt,name=domain,proto3" json:"domain,omitempty"`
	Reason        string                 `protobuf:"bytes,4,opt,name=reason,proto3" json:"reason,omitempty"`
	AppId         string                 `protobuf:"bytes,5,opt,name=app_id,json=appId,proto3" json:"app_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RejectMessage) Reset() {
	*x = RejectMessage{}
	mi := &file_app_userlogger_config_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RejectMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RejectMessage) ProtoMessage() {}

func (x *RejectMessage) ProtoReflect() protoreflect.Message {
	mi := &file_app_userlogger_config_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RejectMessage.ProtoReflect.Descriptor instead.
func (*RejectMessage) Descriptor() ([]byte, []int) {
	return file_app_userlogger_config_proto_rawDescGZIP(), []int{5}
}

func (x *RejectMessage) GetDst() string {
	if x != nil {
		return x.Dst
	}
	return ""
}

func (x *RejectMessage) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *RejectMessage) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *RejectMessage) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *RejectMessage) GetAppId() string {
	if x != nil {
		return x.AppId
	}
	return ""
}

var File_app_userlogger_config_proto protoreflect.FileDescriptor

const file_app_userlogger_config_proto_rawDesc = "" +
	"\n" +
	"\x1bapp/userlogger/config.proto\x12\fx.userlogger\"\xe0\x02\n" +
	"\x0eUserLogMessage\x12A\n" +
	"\rroute_message\x18\x01 \x01(\v2\x1a.x.userlogger.RouteMessageH\x00R\frouteMessage\x12A\n" +
	"\rerror_message\x18\x02 \x01(\v2\x1a.x.userlogger.ErrorMessageH\x00R\ferrorMessage\x12A\n" +
	"\rsession_error\x18\x03 \x01(\v2\x1a.x.userlogger.SessionErrorH\x00R\fsessionError\x12D\n" +
	"\x0ereject_message\x18\x04 \x01(\v2\x1b.x.userlogger.RejectMessageH\x00R\rrejectMessage\x124\n" +
	"\bfallback\x18\x05 \x01(\v2\x16.x.userlogger.FallbackH\x00R\bfallbackB\t\n" +
	"\amessage\"\xfe\x02\n" +
	"\fRouteMessage\x12\x10\n" +
	"\x03dst\x18\x01 \x01(\tR\x03dst\x12\x10\n" +
	"\x03tag\x18\x02 \x01(\tR\x03tag\x12\x1c\n" +
	"\ttimestamp\x18\x03 \x01(\x03R\ttimestamp\x12\x15\n" +
	"\x06app_id\x18\x04 \x01(\tR\x05appId\x12!\n" +
	"\fsniff_domain\x18\x05 \x01(\tR\vsniffDomain\x12\x10\n" +
	"\x03sid\x18\x06 \x01(\rR\x03sid\x12 \n" +
	"\fip_to_domain\x18\a \x01(\tR\n" +
	"ipToDomain\x12!\n" +
	"\fselector_tag\x18\b \x01(\tR\vselectorTag\x12!\n" +
	"\fmatched_rule\x18\t \x01(\tR\vmatchedRule\x12\x1f\n" +
	"\vinbound_tag\x18\n" +
	" \x01(\tR\n" +
	"inboundTag\x12\x18\n" +
	"\anetwork\x18\v \x01(\tR\anetwork\x12%\n" +
	"\x0esniff_protofol\x18\f \x01(\tR\rsniffProtofol\x12\x16\n" +
	"\x06source\x18\r \x01(\tR\x06source\"F\n" +
	"\fErrorMessage\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage\x12\x1c\n" +
	"\ttimestamp\x18\x02 \x01(\x03R\ttimestamp\".\n" +
	"\bFallback\x12\x10\n" +
	"\x03sid\x18\x01 \x01(\rR\x03sid\x12\x10\n" +
	"\x03tag\x18\x02 \x01(\tR\x03tag\"p\n" +
	"\fSessionError\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage\x12\x10\n" +
	"\x03sid\x18\x03 \x01(\rR\x03sid\x12\x0e\n" +
	"\x02up\x18\x04 \x01(\rR\x02up\x12\x12\n" +
	"\x04down\x18\x05 \x01(\rR\x04down\x12\x10\n" +
	"\x03dns\x18\x06 \x01(\tR\x03dns\"\x86\x01\n" +
	"\rRejectMessage\x12\x10\n" +
	"\x03dst\x18\x01 \x01(\tR\x03dst\x12\x1c\n" +
	"\ttimestamp\x18\x02 \x01(\x03R\ttimestamp\x12\x16\n" +
	"\x06domain\x18\x03 \x01(\tR\x06domain\x12\x16\n" +
	"\x06reason\x18\x04 \x01(\tR\x06reason\x12\x15\n" +
	"\x06app_id\x18\x05 \x01(\tR\x05appIdB-Z+github.com/5vnetwork/vx-core/app/userloggerb\x06proto3"

var (
	file_app_userlogger_config_proto_rawDescOnce sync.Once
	file_app_userlogger_config_proto_rawDescData []byte
)

func file_app_userlogger_config_proto_rawDescGZIP() []byte {
	file_app_userlogger_config_proto_rawDescOnce.Do(func() {
		file_app_userlogger_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_userlogger_config_proto_rawDesc), len(file_app_userlogger_config_proto_rawDesc)))
	})
	return file_app_userlogger_config_proto_rawDescData
}

var file_app_userlogger_config_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_app_userlogger_config_proto_goTypes = []any{
	(*UserLogMessage)(nil), // 0: x.userlogger.UserLogMessage
	(*RouteMessage)(nil),   // 1: x.userlogger.RouteMessage
	(*ErrorMessage)(nil),   // 2: x.userlogger.ErrorMessage
	(*Fallback)(nil),       // 3: x.userlogger.Fallback
	(*SessionError)(nil),   // 4: x.userlogger.SessionError
	(*RejectMessage)(nil),  // 5: x.userlogger.RejectMessage
}
var file_app_userlogger_config_proto_depIdxs = []int32{
	1, // 0: x.userlogger.UserLogMessage.route_message:type_name -> x.userlogger.RouteMessage
	2, // 1: x.userlogger.UserLogMessage.error_message:type_name -> x.userlogger.ErrorMessage
	4, // 2: x.userlogger.UserLogMessage.session_error:type_name -> x.userlogger.SessionError
	5, // 3: x.userlogger.UserLogMessage.reject_message:type_name -> x.userlogger.RejectMessage
	3, // 4: x.userlogger.UserLogMessage.fallback:type_name -> x.userlogger.Fallback
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_app_userlogger_config_proto_init() }
func file_app_userlogger_config_proto_init() {
	if File_app_userlogger_config_proto != nil {
		return
	}
	file_app_userlogger_config_proto_msgTypes[0].OneofWrappers = []any{
		(*UserLogMessage_RouteMessage)(nil),
		(*UserLogMessage_ErrorMessage)(nil),
		(*UserLogMessage_SessionError)(nil),
		(*UserLogMessage_RejectMessage)(nil),
		(*UserLogMessage_Fallback)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_userlogger_config_proto_rawDesc), len(file_app_userlogger_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_userlogger_config_proto_goTypes,
		DependencyIndexes: file_app_userlogger_config_proto_depIdxs,
		MessageInfos:      file_app_userlogger_config_proto_msgTypes,
	}.Build()
	File_app_userlogger_config_proto = out.File
	file_app_userlogger_config_proto_goTypes = nil
	file_app_userlogger_config_proto_depIdxs = nil
}
