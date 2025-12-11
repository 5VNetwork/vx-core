package tcp

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
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

type TcpConfig struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	HeaderSettings *anypb.Any             `protobuf:"bytes,2,opt,name=header_settings,json=headerSettings,proto3" json:"header_settings,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *TcpConfig) Reset() {
	*x = TcpConfig{}
	mi := &file_transport_protocols_tcp_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TcpConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TcpConfig) ProtoMessage() {}

func (x *TcpConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_tcp_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TcpConfig.ProtoReflect.Descriptor instead.
func (*TcpConfig) Descriptor() ([]byte, []int) {
	return file_transport_protocols_tcp_config_proto_rawDescGZIP(), []int{0}
}

func (x *TcpConfig) GetHeaderSettings() *anypb.Any {
	if x != nil {
		return x.HeaderSettings
	}
	return nil
}

var File_transport_protocols_tcp_config_proto protoreflect.FileDescriptor

const file_transport_protocols_tcp_config_proto_rawDesc = "" +
	"\n" +
	"$transport/protocols/tcp/config.proto\x12\x19x.transport.protocols.tcp\x1a\x19google/protobuf/any.proto\"J\n" +
	"\tTcpConfig\x12=\n" +
	"\x0fheader_settings\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x0eheaderSettingsB6Z4github.com/5vnetwork/vx-core/transport/protocols/tcpb\x06proto3"

var (
	file_transport_protocols_tcp_config_proto_rawDescOnce sync.Once
	file_transport_protocols_tcp_config_proto_rawDescData []byte
)

func file_transport_protocols_tcp_config_proto_rawDescGZIP() []byte {
	file_transport_protocols_tcp_config_proto_rawDescOnce.Do(func() {
		file_transport_protocols_tcp_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_protocols_tcp_config_proto_rawDesc), len(file_transport_protocols_tcp_config_proto_rawDesc)))
	})
	return file_transport_protocols_tcp_config_proto_rawDescData
}

var file_transport_protocols_tcp_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_protocols_tcp_config_proto_goTypes = []any{
	(*TcpConfig)(nil), // 0: x.transport.protocols.tcp.TcpConfig
	(*anypb.Any)(nil), // 1: google.protobuf.Any
}
var file_transport_protocols_tcp_config_proto_depIdxs = []int32{
	1, // 0: x.transport.protocols.tcp.TcpConfig.header_settings:type_name -> google.protobuf.Any
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_protocols_tcp_config_proto_init() }
func file_transport_protocols_tcp_config_proto_init() {
	if File_transport_protocols_tcp_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_protocols_tcp_config_proto_rawDesc), len(file_transport_protocols_tcp_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_protocols_tcp_config_proto_goTypes,
		DependencyIndexes: file_transport_protocols_tcp_config_proto_depIdxs,
		MessageInfos:      file_transport_protocols_tcp_config_proto_msgTypes,
	}.Build()
	File_transport_protocols_tcp_config_proto = out.File
	file_transport_protocols_tcp_config_proto_goTypes = nil
	file_transport_protocols_tcp_config_proto_depIdxs = nil
}
