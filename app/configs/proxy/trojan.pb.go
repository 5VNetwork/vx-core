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

type TrojanClientConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Password      string                 `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	Special       bool                   `protobuf:"varint,3,opt,name=special,proto3" json:"special,omitempty"`
	Vision        bool                   `protobuf:"varint,4,opt,name=vision,proto3" json:"vision,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TrojanClientConfig) Reset() {
	*x = TrojanClientConfig{}
	mi := &file_protos_proxy_trojan_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TrojanClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrojanClientConfig) ProtoMessage() {}

func (x *TrojanClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_trojan_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrojanClientConfig.ProtoReflect.Descriptor instead.
func (*TrojanClientConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_trojan_proto_rawDescGZIP(), []int{0}
}

func (x *TrojanClientConfig) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *TrojanClientConfig) GetSpecial() bool {
	if x != nil {
		return x.Special
	}
	return false
}

func (x *TrojanClientConfig) GetVision() bool {
	if x != nil {
		return x.Vision
	}
	return false
}

type TrojanServerConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Users         []*configs.UserConfig  `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
	Vision        bool                   `protobuf:"varint,2,opt,name=vision,proto3" json:"vision,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TrojanServerConfig) Reset() {
	*x = TrojanServerConfig{}
	mi := &file_protos_proxy_trojan_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TrojanServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrojanServerConfig) ProtoMessage() {}

func (x *TrojanServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_trojan_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrojanServerConfig.ProtoReflect.Descriptor instead.
func (*TrojanServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_trojan_proto_rawDescGZIP(), []int{1}
}

func (x *TrojanServerConfig) GetUsers() []*configs.UserConfig {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *TrojanServerConfig) GetVision() bool {
	if x != nil {
		return x.Vision
	}
	return false
}

var File_protos_proxy_trojan_proto protoreflect.FileDescriptor

const file_protos_proxy_trojan_proto_rawDesc = "" +
	"\n" +
	"\x19protos/proxy/trojan.proto\x12\ax.proxy\x1a\x11protos/user.proto\"b\n" +
	"\x12TrojanClientConfig\x12\x1a\n" +
	"\bpassword\x18\x02 \x01(\tR\bpassword\x12\x18\n" +
	"\aspecial\x18\x03 \x01(\bR\aspecial\x12\x16\n" +
	"\x06vision\x18\x04 \x01(\bR\x06vision\"W\n" +
	"\x12TrojanServerConfig\x12#\n" +
	"\x05users\x18\x01 \x03(\v2\r.x.UserConfigR\x05users\x12\x16\n" +
	"\x06vision\x18\x02 \x01(\bR\x06visionJ\x04\b\x03\x10\x04B0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_trojan_proto_rawDescOnce sync.Once
	file_protos_proxy_trojan_proto_rawDescData []byte
)

func file_protos_proxy_trojan_proto_rawDescGZIP() []byte {
	file_protos_proxy_trojan_proto_rawDescOnce.Do(func() {
		file_protos_proxy_trojan_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_trojan_proto_rawDesc), len(file_protos_proxy_trojan_proto_rawDesc)))
	})
	return file_protos_proxy_trojan_proto_rawDescData
}

var file_protos_proxy_trojan_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protos_proxy_trojan_proto_goTypes = []any{
	(*TrojanClientConfig)(nil), // 0: x.proxy.TrojanClientConfig
	(*TrojanServerConfig)(nil), // 1: x.proxy.TrojanServerConfig
	(*configs.UserConfig)(nil), // 2: x.UserConfig
}
var file_protos_proxy_trojan_proto_depIdxs = []int32{
	2, // 0: x.proxy.TrojanServerConfig.users:type_name -> x.UserConfig
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protos_proxy_trojan_proto_init() }
func file_protos_proxy_trojan_proto_init() {
	if File_protos_proxy_trojan_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_trojan_proto_rawDesc), len(file_protos_proxy_trojan_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_trojan_proto_goTypes,
		DependencyIndexes: file_protos_proxy_trojan_proto_depIdxs,
		MessageInfos:      file_protos_proxy_trojan_proto_msgTypes,
	}.Build()
	File_protos_proxy_trojan_proto = out.File
	file_protos_proxy_trojan_proto_goTypes = nil
	file_protos_proxy_trojan_proto_depIdxs = nil
}
