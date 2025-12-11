package quic

import (
	protocol "github.com/5vnetwork/vx-core/common/protocol"
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

type QuicConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Key           string                 `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Security      protocol.SecurityType  `protobuf:"varint,2,opt,name=security,proto3,enum=x.common.protocol.SecurityType" json:"security,omitempty"`
	Header        *anypb.Any             `protobuf:"bytes,3,opt,name=header,proto3" json:"header,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QuicConfig) Reset() {
	*x = QuicConfig{}
	mi := &file_transport_protocols_quic_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QuicConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QuicConfig) ProtoMessage() {}

func (x *QuicConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_quic_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QuicConfig.ProtoReflect.Descriptor instead.
func (*QuicConfig) Descriptor() ([]byte, []int) {
	return file_transport_protocols_quic_config_proto_rawDescGZIP(), []int{0}
}

func (x *QuicConfig) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *QuicConfig) GetSecurity() protocol.SecurityType {
	if x != nil {
		return x.Security
	}
	return protocol.SecurityType(0)
}

func (x *QuicConfig) GetHeader() *anypb.Any {
	if x != nil {
		return x.Header
	}
	return nil
}

var File_transport_protocols_quic_config_proto protoreflect.FileDescriptor

const file_transport_protocols_quic_config_proto_rawDesc = "" +
	"\n" +
	"%transport/protocols/quic/config.proto\x12\x1ax.transport.protocols.quic\x1a\x19google/protobuf/any.proto\x1a\x1ecommon/protocol/protocol.proto\"\x89\x01\n" +
	"\n" +
	"QuicConfig\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12;\n" +
	"\bsecurity\x18\x02 \x01(\x0e2\x1f.x.common.protocol.SecurityTypeR\bsecurity\x12,\n" +
	"\x06header\x18\x03 \x01(\v2\x14.google.protobuf.AnyR\x06headerB7Z5github.com/5vnetwork/vx-core/transport/protocols/quicb\x06proto3"

var (
	file_transport_protocols_quic_config_proto_rawDescOnce sync.Once
	file_transport_protocols_quic_config_proto_rawDescData []byte
)

func file_transport_protocols_quic_config_proto_rawDescGZIP() []byte {
	file_transport_protocols_quic_config_proto_rawDescOnce.Do(func() {
		file_transport_protocols_quic_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_protocols_quic_config_proto_rawDesc), len(file_transport_protocols_quic_config_proto_rawDesc)))
	})
	return file_transport_protocols_quic_config_proto_rawDescData
}

var file_transport_protocols_quic_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_protocols_quic_config_proto_goTypes = []any{
	(*QuicConfig)(nil),         // 0: x.transport.protocols.quic.QuicConfig
	(protocol.SecurityType)(0), // 1: x.common.protocol.SecurityType
	(*anypb.Any)(nil),          // 2: google.protobuf.Any
}
var file_transport_protocols_quic_config_proto_depIdxs = []int32{
	1, // 0: x.transport.protocols.quic.QuicConfig.security:type_name -> x.common.protocol.SecurityType
	2, // 1: x.transport.protocols.quic.QuicConfig.header:type_name -> google.protobuf.Any
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_transport_protocols_quic_config_proto_init() }
func file_transport_protocols_quic_config_proto_init() {
	if File_transport_protocols_quic_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_protocols_quic_config_proto_rawDesc), len(file_transport_protocols_quic_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_protocols_quic_config_proto_goTypes,
		DependencyIndexes: file_transport_protocols_quic_config_proto_depIdxs,
		MessageInfos:      file_transport_protocols_quic_config_proto_msgTypes,
	}.Build()
	File_transport_protocols_quic_config_proto = out.File
	file_transport_protocols_quic_config_proto_goTypes = nil
	file_transport_protocols_quic_config_proto_depIdxs = nil
}
