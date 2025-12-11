package proxy

import (
	configs "github.com/5vnetwork/vx-core/app/configs"
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

type ViClientConfig struct {
	state   protoimpl.MessageState `protogen:"open.v1"`
	Address string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port    uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// uuid
	Secret        string `protobuf:"bytes,3,opt,name=secret,proto3" json:"secret,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ViClientConfig) Reset() {
	*x = ViClientConfig{}
	mi := &file_protos_proxy_vi_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ViClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ViClientConfig) ProtoMessage() {}

func (x *ViClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_vi_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ViClientConfig.ProtoReflect.Descriptor instead.
func (*ViClientConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_vi_proto_rawDescGZIP(), []int{0}
}

func (x *ViClientConfig) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ViClientConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ViClientConfig) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

type ViServerConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Users         []*configs.UserConfig  `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ViServerConfig) Reset() {
	*x = ViServerConfig{}
	mi := &file_protos_proxy_vi_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ViServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ViServerConfig) ProtoMessage() {}

func (x *ViServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_vi_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ViServerConfig.ProtoReflect.Descriptor instead.
func (*ViServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_vi_proto_rawDescGZIP(), []int{1}
}

func (x *ViServerConfig) GetUsers() []*configs.UserConfig {
	if x != nil {
		return x.Users
	}
	return nil
}

var File_protos_proxy_vi_proto protoreflect.FileDescriptor

const file_protos_proxy_vi_proto_rawDesc = "" +
	"\n" +
	"\x15protos/proxy/vi.proto\x12\ax.proxy\x1a\x11protos/user.proto\"V\n" +
	"\x0eViClientConfig\x12\x18\n" +
	"\aaddress\x18\x01 \x01(\tR\aaddress\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\x12\x16\n" +
	"\x06secret\x18\x03 \x01(\tR\x06secret\"5\n" +
	"\x0eViServerConfig\x12#\n" +
	"\x05users\x18\x01 \x03(\v2\r.x.UserConfigR\x05usersB0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_vi_proto_rawDescOnce sync.Once
	file_protos_proxy_vi_proto_rawDescData []byte
)

func file_protos_proxy_vi_proto_rawDescGZIP() []byte {
	file_protos_proxy_vi_proto_rawDescOnce.Do(func() {
		file_protos_proxy_vi_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_vi_proto_rawDesc), len(file_protos_proxy_vi_proto_rawDesc)))
	})
	return file_protos_proxy_vi_proto_rawDescData
}

var file_protos_proxy_vi_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protos_proxy_vi_proto_goTypes = []any{
	(*ViClientConfig)(nil),     // 0: x.proxy.ViClientConfig
	(*ViServerConfig)(nil),     // 1: x.proxy.ViServerConfig
	(*configs.UserConfig)(nil), // 2: x.UserConfig
}
var file_protos_proxy_vi_proto_depIdxs = []int32{
	2, // 0: x.proxy.ViServerConfig.users:type_name -> x.UserConfig
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protos_proxy_vi_proto_init() }
func file_protos_proxy_vi_proto_init() {
	if File_protos_proxy_vi_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_vi_proto_rawDesc), len(file_protos_proxy_vi_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_vi_proto_goTypes,
		DependencyIndexes: file_protos_proxy_vi_proto_depIdxs,
		MessageInfos:      file_protos_proxy_vi_proto_msgTypes,
	}.Build()
	File_protos_proxy_vi_proto = out.File
	file_protos_proxy_vi_proto_goTypes = nil
	file_protos_proxy_vi_proto_depIdxs = nil
}
