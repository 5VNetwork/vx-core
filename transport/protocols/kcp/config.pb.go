package kcp

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

type KcpConfig struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Mtu              uint32                 `protobuf:"varint,1,opt,name=mtu,proto3" json:"mtu,omitempty"`
	Tti              uint32                 `protobuf:"varint,2,opt,name=tti,proto3" json:"tti,omitempty"`
	UplinkCapacity   uint32                 `protobuf:"varint,3,opt,name=uplink_capacity,json=uplinkCapacity,proto3" json:"uplink_capacity,omitempty"`
	DownlinkCapacity uint32                 `protobuf:"varint,4,opt,name=downlink_capacity,json=downlinkCapacity,proto3" json:"downlink_capacity,omitempty"`
	Congestion       bool                   `protobuf:"varint,5,opt,name=congestion,proto3" json:"congestion,omitempty"`
	WriteBuffer      uint32                 `protobuf:"varint,6,opt,name=write_buffer,json=writeBuffer,proto3" json:"write_buffer,omitempty"`
	ReadBuffer       uint32                 `protobuf:"varint,7,opt,name=read_buffer,json=readBuffer,proto3" json:"read_buffer,omitempty"`
	HeaderConfig     *anypb.Any             `protobuf:"bytes,8,opt,name=header_config,json=headerConfig,proto3" json:"header_config,omitempty"`
	Seed             string                 `protobuf:"bytes,10,opt,name=seed,proto3" json:"seed,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *KcpConfig) Reset() {
	*x = KcpConfig{}
	mi := &file_transport_protocols_kcp_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *KcpConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KcpConfig) ProtoMessage() {}

func (x *KcpConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_kcp_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KcpConfig.ProtoReflect.Descriptor instead.
func (*KcpConfig) Descriptor() ([]byte, []int) {
	return file_transport_protocols_kcp_config_proto_rawDescGZIP(), []int{0}
}

func (x *KcpConfig) GetMtu() uint32 {
	if x != nil {
		return x.Mtu
	}
	return 0
}

func (x *KcpConfig) GetTti() uint32 {
	if x != nil {
		return x.Tti
	}
	return 0
}

func (x *KcpConfig) GetUplinkCapacity() uint32 {
	if x != nil {
		return x.UplinkCapacity
	}
	return 0
}

func (x *KcpConfig) GetDownlinkCapacity() uint32 {
	if x != nil {
		return x.DownlinkCapacity
	}
	return 0
}

func (x *KcpConfig) GetCongestion() bool {
	if x != nil {
		return x.Congestion
	}
	return false
}

func (x *KcpConfig) GetWriteBuffer() uint32 {
	if x != nil {
		return x.WriteBuffer
	}
	return 0
}

func (x *KcpConfig) GetReadBuffer() uint32 {
	if x != nil {
		return x.ReadBuffer
	}
	return 0
}

func (x *KcpConfig) GetHeaderConfig() *anypb.Any {
	if x != nil {
		return x.HeaderConfig
	}
	return nil
}

func (x *KcpConfig) GetSeed() string {
	if x != nil {
		return x.Seed
	}
	return ""
}

var File_transport_protocols_kcp_config_proto protoreflect.FileDescriptor

const file_transport_protocols_kcp_config_proto_rawDesc = "" +
	"\n" +
	"$transport/protocols/kcp/config.proto\x12\x19x.transport.protocols.kcp\x1a\x19google/protobuf/any.proto\"\xbe\x02\n" +
	"\tKcpConfig\x12\x10\n" +
	"\x03mtu\x18\x01 \x01(\rR\x03mtu\x12\x10\n" +
	"\x03tti\x18\x02 \x01(\rR\x03tti\x12'\n" +
	"\x0fuplink_capacity\x18\x03 \x01(\rR\x0euplinkCapacity\x12+\n" +
	"\x11downlink_capacity\x18\x04 \x01(\rR\x10downlinkCapacity\x12\x1e\n" +
	"\n" +
	"congestion\x18\x05 \x01(\bR\n" +
	"congestion\x12!\n" +
	"\fwrite_buffer\x18\x06 \x01(\rR\vwriteBuffer\x12\x1f\n" +
	"\vread_buffer\x18\a \x01(\rR\n" +
	"readBuffer\x129\n" +
	"\rheader_config\x18\b \x01(\v2\x14.google.protobuf.AnyR\fheaderConfig\x12\x12\n" +
	"\x04seed\x18\n" +
	" \x01(\tR\x04seedJ\x04\b\t\x10\n" +
	"B6Z4github.com/5vnetwork/vx-core/transport/protocols/kcpb\x06proto3"

var (
	file_transport_protocols_kcp_config_proto_rawDescOnce sync.Once
	file_transport_protocols_kcp_config_proto_rawDescData []byte
)

func file_transport_protocols_kcp_config_proto_rawDescGZIP() []byte {
	file_transport_protocols_kcp_config_proto_rawDescOnce.Do(func() {
		file_transport_protocols_kcp_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_protocols_kcp_config_proto_rawDesc), len(file_transport_protocols_kcp_config_proto_rawDesc)))
	})
	return file_transport_protocols_kcp_config_proto_rawDescData
}

var file_transport_protocols_kcp_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_protocols_kcp_config_proto_goTypes = []any{
	(*KcpConfig)(nil), // 0: x.transport.protocols.kcp.KcpConfig
	(*anypb.Any)(nil), // 1: google.protobuf.Any
}
var file_transport_protocols_kcp_config_proto_depIdxs = []int32{
	1, // 0: x.transport.protocols.kcp.KcpConfig.header_config:type_name -> google.protobuf.Any
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_protocols_kcp_config_proto_init() }
func file_transport_protocols_kcp_config_proto_init() {
	if File_transport_protocols_kcp_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_protocols_kcp_config_proto_rawDesc), len(file_transport_protocols_kcp_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_protocols_kcp_config_proto_goTypes,
		DependencyIndexes: file_transport_protocols_kcp_config_proto_depIdxs,
		MessageInfos:      file_transport_protocols_kcp_config_proto_msgTypes,
	}.Build()
	File_transport_protocols_kcp_config_proto = out.File
	file_transport_protocols_kcp_config_proto_goTypes = nil
	file_transport_protocols_kcp_config_proto_depIdxs = nil
}
