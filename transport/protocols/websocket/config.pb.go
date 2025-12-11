package websocket

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

type Header struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Key           string                 `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value         string                 `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Header) Reset() {
	*x = Header{}
	mi := &file_transport_protocols_websocket_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_websocket_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_transport_protocols_websocket_config_proto_rawDescGZIP(), []int{0}
}

func (x *Header) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Header) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type WebsocketConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Host  string                 `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	// URL path to the WebSocket service. Empty value means root(/).
	Path                 string    `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Header               []*Header `protobuf:"bytes,3,rep,name=header,proto3" json:"header,omitempty"`
	MaxEarlyData         int32     `protobuf:"varint,5,opt,name=max_early_data,json=maxEarlyData,proto3" json:"max_early_data,omitempty"`
	UseBrowserForwarding bool      `protobuf:"varint,6,opt,name=use_browser_forwarding,json=useBrowserForwarding,proto3" json:"use_browser_forwarding,omitempty"`
	EarlyDataHeaderName  string    `protobuf:"bytes,7,opt,name=early_data_header_name,json=earlyDataHeaderName,proto3" json:"early_data_header_name,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *WebsocketConfig) Reset() {
	*x = WebsocketConfig{}
	mi := &file_transport_protocols_websocket_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WebsocketConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebsocketConfig) ProtoMessage() {}

func (x *WebsocketConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_websocket_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebsocketConfig.ProtoReflect.Descriptor instead.
func (*WebsocketConfig) Descriptor() ([]byte, []int) {
	return file_transport_protocols_websocket_config_proto_rawDescGZIP(), []int{1}
}

func (x *WebsocketConfig) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *WebsocketConfig) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *WebsocketConfig) GetHeader() []*Header {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *WebsocketConfig) GetMaxEarlyData() int32 {
	if x != nil {
		return x.MaxEarlyData
	}
	return 0
}

func (x *WebsocketConfig) GetUseBrowserForwarding() bool {
	if x != nil {
		return x.UseBrowserForwarding
	}
	return false
}

func (x *WebsocketConfig) GetEarlyDataHeaderName() string {
	if x != nil {
		return x.EarlyDataHeaderName
	}
	return ""
}

var File_transport_protocols_websocket_config_proto protoreflect.FileDescriptor

const file_transport_protocols_websocket_config_proto_rawDesc = "" +
	"\n" +
	"*transport/protocols/websocket/config.proto\x12\x1fx.transport.protocols.websocket\"0\n" +
	"\x06Header\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value\"\x8b\x02\n" +
	"\x0fWebsocketConfig\x12\x12\n" +
	"\x04host\x18\x01 \x01(\tR\x04host\x12\x12\n" +
	"\x04path\x18\x02 \x01(\tR\x04path\x12?\n" +
	"\x06header\x18\x03 \x03(\v2'.x.transport.protocols.websocket.HeaderR\x06header\x12$\n" +
	"\x0emax_early_data\x18\x05 \x01(\x05R\fmaxEarlyData\x124\n" +
	"\x16use_browser_forwarding\x18\x06 \x01(\bR\x14useBrowserForwarding\x123\n" +
	"\x16early_data_header_name\x18\a \x01(\tR\x13earlyDataHeaderNameB<Z:github.com/5vnetwork/vx-core/transport/protocols/websocketb\x06proto3"

var (
	file_transport_protocols_websocket_config_proto_rawDescOnce sync.Once
	file_transport_protocols_websocket_config_proto_rawDescData []byte
)

func file_transport_protocols_websocket_config_proto_rawDescGZIP() []byte {
	file_transport_protocols_websocket_config_proto_rawDescOnce.Do(func() {
		file_transport_protocols_websocket_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_protocols_websocket_config_proto_rawDesc), len(file_transport_protocols_websocket_config_proto_rawDesc)))
	})
	return file_transport_protocols_websocket_config_proto_rawDescData
}

var file_transport_protocols_websocket_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_protocols_websocket_config_proto_goTypes = []any{
	(*Header)(nil),          // 0: x.transport.protocols.websocket.Header
	(*WebsocketConfig)(nil), // 1: x.transport.protocols.websocket.WebsocketConfig
}
var file_transport_protocols_websocket_config_proto_depIdxs = []int32{
	0, // 0: x.transport.protocols.websocket.WebsocketConfig.header:type_name -> x.transport.protocols.websocket.Header
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_protocols_websocket_config_proto_init() }
func file_transport_protocols_websocket_config_proto_init() {
	if File_transport_protocols_websocket_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_protocols_websocket_config_proto_rawDesc), len(file_transport_protocols_websocket_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_protocols_websocket_config_proto_goTypes,
		DependencyIndexes: file_transport_protocols_websocket_config_proto_depIdxs,
		MessageInfos:      file_transport_protocols_websocket_config_proto_msgTypes,
	}.Build()
	File_transport_protocols_websocket_config_proto = out.File
	file_transport_protocols_websocket_config_proto_goTypes = nil
	file_transport_protocols_websocket_config_proto_depIdxs = nil
}
