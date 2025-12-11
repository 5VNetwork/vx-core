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

type VlessClientConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// ID of the account, in the form of a UUID, e.g.,
	// "66ad4540-b58c-4ad2-9926-ea63445a9b57".
	Id string `protobuf:"bytes,5,opt,name=id,proto3" json:"id,omitempty"`
	// Flow settings. May be "xtls-rprx-vision".
	Flow string `protobuf:"bytes,6,opt,name=flow,proto3" json:"flow,omitempty"`
	// Encryption settings. Only applies to client side, and only accepts "none"
	// for now.
	Encryption    string `protobuf:"bytes,7,opt,name=encryption,proto3" json:"encryption,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VlessClientConfig) Reset() {
	*x = VlessClientConfig{}
	mi := &file_protos_proxy_vless_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VlessClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VlessClientConfig) ProtoMessage() {}

func (x *VlessClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_vless_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VlessClientConfig.ProtoReflect.Descriptor instead.
func (*VlessClientConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_vless_proto_rawDescGZIP(), []int{0}
}

func (x *VlessClientConfig) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *VlessClientConfig) GetFlow() string {
	if x != nil {
		return x.Flow
	}
	return ""
}

func (x *VlessClientConfig) GetEncryption() string {
	if x != nil {
		return x.Encryption
	}
	return ""
}

type VlessServerConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Users         []*configs.UserConfig  `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VlessServerConfig) Reset() {
	*x = VlessServerConfig{}
	mi := &file_protos_proxy_vless_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VlessServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VlessServerConfig) ProtoMessage() {}

func (x *VlessServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_vless_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VlessServerConfig.ProtoReflect.Descriptor instead.
func (*VlessServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_vless_proto_rawDescGZIP(), []int{1}
}

func (x *VlessServerConfig) GetUsers() []*configs.UserConfig {
	if x != nil {
		return x.Users
	}
	return nil
}

var File_protos_proxy_vless_proto protoreflect.FileDescriptor

const file_protos_proxy_vless_proto_rawDesc = "" +
	"\n" +
	"\x18protos/proxy/vless.proto\x12\ax.proxy\x1a\x11protos/user.proto\"W\n" +
	"\x11VlessClientConfig\x12\x0e\n" +
	"\x02id\x18\x05 \x01(\tR\x02id\x12\x12\n" +
	"\x04flow\x18\x06 \x01(\tR\x04flow\x12\x1e\n" +
	"\n" +
	"encryption\x18\a \x01(\tR\n" +
	"encryption\"8\n" +
	"\x11VlessServerConfig\x12#\n" +
	"\x05users\x18\x01 \x03(\v2\r.x.UserConfigR\x05usersB0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_vless_proto_rawDescOnce sync.Once
	file_protos_proxy_vless_proto_rawDescData []byte
)

func file_protos_proxy_vless_proto_rawDescGZIP() []byte {
	file_protos_proxy_vless_proto_rawDescOnce.Do(func() {
		file_protos_proxy_vless_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_vless_proto_rawDesc), len(file_protos_proxy_vless_proto_rawDesc)))
	})
	return file_protos_proxy_vless_proto_rawDescData
}

var file_protos_proxy_vless_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protos_proxy_vless_proto_goTypes = []any{
	(*VlessClientConfig)(nil),  // 0: x.proxy.VlessClientConfig
	(*VlessServerConfig)(nil),  // 1: x.proxy.VlessServerConfig
	(*configs.UserConfig)(nil), // 2: x.UserConfig
}
var file_protos_proxy_vless_proto_depIdxs = []int32{
	2, // 0: x.proxy.VlessServerConfig.users:type_name -> x.UserConfig
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protos_proxy_vless_proto_init() }
func file_protos_proxy_vless_proto_init() {
	if File_protos_proxy_vless_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_vless_proto_rawDesc), len(file_protos_proxy_vless_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_vless_proto_goTypes,
		DependencyIndexes: file_protos_proxy_vless_proto_depIdxs,
		MessageInfos:      file_protos_proxy_vless_proto_msgTypes,
	}.Build()
	File_protos_proxy_vless_proto = out.File
	file_protos_proxy_vless_proto_goTypes = nil
	file_protos_proxy_vless_proto_depIdxs = nil
}
