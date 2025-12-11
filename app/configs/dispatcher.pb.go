package configs

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

type DispatcherConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// SniffingConfig sniff = 1;
	FallbackToProxy      bool     `protobuf:"varint,2,opt,name=fallback_to_proxy,json=fallbackToProxy,proto3" json:"fallback_to_proxy,omitempty"`
	FallbackToDomain     bool     `protobuf:"varint,3,opt,name=fallback_to_domain,json=fallbackToDomain,proto3" json:"fallback_to_domain,omitempty"`
	DestinationOverride  []string `protobuf:"bytes,4,rep,name=destination_override,json=destinationOverride,proto3" json:"destination_override,omitempty"`
	Sniff                bool     `protobuf:"varint,5,opt,name=sniff,proto3" json:"sniff,omitempty"`
	Ipv6FallbackToDomain bool     `protobuf:"varint,6,opt,name=ipv6_fallback_to_domain,json=ipv6FallbackToDomain,proto3" json:"ipv6_fallback_to_domain,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *DispatcherConfig) Reset() {
	*x = DispatcherConfig{}
	mi := &file_protos_dispatcher_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DispatcherConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DispatcherConfig) ProtoMessage() {}

func (x *DispatcherConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dispatcher_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DispatcherConfig.ProtoReflect.Descriptor instead.
func (*DispatcherConfig) Descriptor() ([]byte, []int) {
	return file_protos_dispatcher_proto_rawDescGZIP(), []int{0}
}

func (x *DispatcherConfig) GetFallbackToProxy() bool {
	if x != nil {
		return x.FallbackToProxy
	}
	return false
}

func (x *DispatcherConfig) GetFallbackToDomain() bool {
	if x != nil {
		return x.FallbackToDomain
	}
	return false
}

func (x *DispatcherConfig) GetDestinationOverride() []string {
	if x != nil {
		return x.DestinationOverride
	}
	return nil
}

func (x *DispatcherConfig) GetSniff() bool {
	if x != nil {
		return x.Sniff
	}
	return false
}

func (x *DispatcherConfig) GetIpv6FallbackToDomain() bool {
	if x != nil {
		return x.Ipv6FallbackToDomain
	}
	return false
}

var File_protos_dispatcher_proto protoreflect.FileDescriptor

const file_protos_dispatcher_proto_rawDesc = "" +
	"\n" +
	"\x17protos/dispatcher.proto\x12\x01x\"\xec\x01\n" +
	"\x10DispatcherConfig\x12*\n" +
	"\x11fallback_to_proxy\x18\x02 \x01(\bR\x0ffallbackToProxy\x12,\n" +
	"\x12fallback_to_domain\x18\x03 \x01(\bR\x10fallbackToDomain\x121\n" +
	"\x14destination_override\x18\x04 \x03(\tR\x13destinationOverride\x12\x14\n" +
	"\x05sniff\x18\x05 \x01(\bR\x05sniff\x125\n" +
	"\x17ipv6_fallback_to_domain\x18\x06 \x01(\bR\x14ipv6FallbackToDomainB*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_dispatcher_proto_rawDescOnce sync.Once
	file_protos_dispatcher_proto_rawDescData []byte
)

func file_protos_dispatcher_proto_rawDescGZIP() []byte {
	file_protos_dispatcher_proto_rawDescOnce.Do(func() {
		file_protos_dispatcher_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_dispatcher_proto_rawDesc), len(file_protos_dispatcher_proto_rawDesc)))
	})
	return file_protos_dispatcher_proto_rawDescData
}

var file_protos_dispatcher_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_dispatcher_proto_goTypes = []any{
	(*DispatcherConfig)(nil), // 0: x.DispatcherConfig
}
var file_protos_dispatcher_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protos_dispatcher_proto_init() }
func file_protos_dispatcher_proto_init() {
	if File_protos_dispatcher_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_dispatcher_proto_rawDesc), len(file_protos_dispatcher_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_dispatcher_proto_goTypes,
		DependencyIndexes: file_protos_dispatcher_proto_depIdxs,
		MessageInfos:      file_protos_dispatcher_proto_msgTypes,
	}.Build()
	File_protos_dispatcher_proto = out.File
	file_protos_dispatcher_proto_goTypes = nil
	file_protos_dispatcher_proto_depIdxs = nil
}
