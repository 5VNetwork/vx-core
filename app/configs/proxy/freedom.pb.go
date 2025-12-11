package proxy

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

type FreedomConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FreedomConfig) Reset() {
	*x = FreedomConfig{}
	mi := &file_protos_proxy_freedom_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FreedomConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FreedomConfig) ProtoMessage() {}

func (x *FreedomConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_freedom_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FreedomConfig.ProtoReflect.Descriptor instead.
func (*FreedomConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_freedom_proto_rawDescGZIP(), []int{0}
}

var File_protos_proxy_freedom_proto protoreflect.FileDescriptor

const file_protos_proxy_freedom_proto_rawDesc = "" +
	"\n" +
	"\x1aprotos/proxy/freedom.proto\x12\ax.proxy\"\x0f\n" +
	"\rFreedomConfigB0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_freedom_proto_rawDescOnce sync.Once
	file_protos_proxy_freedom_proto_rawDescData []byte
)

func file_protos_proxy_freedom_proto_rawDescGZIP() []byte {
	file_protos_proxy_freedom_proto_rawDescOnce.Do(func() {
		file_protos_proxy_freedom_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_freedom_proto_rawDesc), len(file_protos_proxy_freedom_proto_rawDesc)))
	})
	return file_protos_proxy_freedom_proto_rawDescData
}

var file_protos_proxy_freedom_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_proxy_freedom_proto_goTypes = []any{
	(*FreedomConfig)(nil), // 0: x.proxy.FreedomConfig
}
var file_protos_proxy_freedom_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protos_proxy_freedom_proto_init() }
func file_protos_proxy_freedom_proto_init() {
	if File_protos_proxy_freedom_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_freedom_proto_rawDesc), len(file_protos_proxy_freedom_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_freedom_proto_goTypes,
		DependencyIndexes: file_protos_proxy_freedom_proto_depIdxs,
		MessageInfos:      file_protos_proxy_freedom_proto_msgTypes,
	}.Build()
	File_protos_proxy_freedom_proto = out.File
	file_protos_proxy_freedom_proto_goTypes = nil
	file_protos_proxy_freedom_proto_depIdxs = nil
}
