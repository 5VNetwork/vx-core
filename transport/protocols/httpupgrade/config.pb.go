package httpupgrade

import (
	websocket "github.com/5vnetwork/vx-core/transport/protocols/websocket"
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

type HttpUpgradeConfig struct {
	state         protoimpl.MessageState     `protogen:"open.v1"`
	Config        *websocket.WebsocketConfig `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HttpUpgradeConfig) Reset() {
	*x = HttpUpgradeConfig{}
	mi := &file_transport_protocols_httpupgrade_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HttpUpgradeConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HttpUpgradeConfig) ProtoMessage() {}

func (x *HttpUpgradeConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_httpupgrade_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HttpUpgradeConfig.ProtoReflect.Descriptor instead.
func (*HttpUpgradeConfig) Descriptor() ([]byte, []int) {
	return file_transport_protocols_httpupgrade_config_proto_rawDescGZIP(), []int{0}
}

func (x *HttpUpgradeConfig) GetConfig() *websocket.WebsocketConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

var File_transport_protocols_httpupgrade_config_proto protoreflect.FileDescriptor

const file_transport_protocols_httpupgrade_config_proto_rawDesc = "" +
	"\n" +
	",transport/protocols/httpupgrade/config.proto\x12!x.transport.protocols.httpupgrade\x1a*transport/protocols/websocket/config.proto\"]\n" +
	"\x11HttpUpgradeConfig\x12H\n" +
	"\x06config\x18\x01 \x01(\v20.x.transport.protocols.websocket.WebsocketConfigR\x06configB>Z<github.com/5vnetwork/vx-core/transport/protocols/httpupgradeb\x06proto3"

var (
	file_transport_protocols_httpupgrade_config_proto_rawDescOnce sync.Once
	file_transport_protocols_httpupgrade_config_proto_rawDescData []byte
)

func file_transport_protocols_httpupgrade_config_proto_rawDescGZIP() []byte {
	file_transport_protocols_httpupgrade_config_proto_rawDescOnce.Do(func() {
		file_transport_protocols_httpupgrade_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_protocols_httpupgrade_config_proto_rawDesc), len(file_transport_protocols_httpupgrade_config_proto_rawDesc)))
	})
	return file_transport_protocols_httpupgrade_config_proto_rawDescData
}

var file_transport_protocols_httpupgrade_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_protocols_httpupgrade_config_proto_goTypes = []any{
	(*HttpUpgradeConfig)(nil),         // 0: x.transport.protocols.httpupgrade.HttpUpgradeConfig
	(*websocket.WebsocketConfig)(nil), // 1: x.transport.protocols.websocket.WebsocketConfig
}
var file_transport_protocols_httpupgrade_config_proto_depIdxs = []int32{
	1, // 0: x.transport.protocols.httpupgrade.HttpUpgradeConfig.config:type_name -> x.transport.protocols.websocket.WebsocketConfig
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_protocols_httpupgrade_config_proto_init() }
func file_transport_protocols_httpupgrade_config_proto_init() {
	if File_transport_protocols_httpupgrade_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_protocols_httpupgrade_config_proto_rawDesc), len(file_transport_protocols_httpupgrade_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_protocols_httpupgrade_config_proto_goTypes,
		DependencyIndexes: file_transport_protocols_httpupgrade_config_proto_depIdxs,
		MessageInfos:      file_transport_protocols_httpupgrade_config_proto_msgTypes,
	}.Build()
	File_transport_protocols_httpupgrade_config_proto = out.File
	file_transport_protocols_httpupgrade_config_proto_goTypes = nil
	file_transport_protocols_httpupgrade_config_proto_depIdxs = nil
}
