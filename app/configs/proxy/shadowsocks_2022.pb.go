package proxy

import (
	configs "github.com/5vnetwork/vx-core/app/configs"
	net "github.com/5vnetwork/vx-core/common/net"
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

type Shadowsocks2022ServerConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Method        string                 `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	User          *configs.UserConfig    `protobuf:"bytes,2,opt,name=user,proto3" json:"user,omitempty"`
	Networks      []net.Network          `protobuf:"varint,5,rep,packed,name=networks,proto3,enum=x.common.net.Network" json:"networks,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Shadowsocks2022ServerConfig) Reset() {
	*x = Shadowsocks2022ServerConfig{}
	mi := &file_protos_proxy_shadowsocks_2022_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Shadowsocks2022ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Shadowsocks2022ServerConfig) ProtoMessage() {}

func (x *Shadowsocks2022ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_shadowsocks_2022_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Shadowsocks2022ServerConfig.ProtoReflect.Descriptor instead.
func (*Shadowsocks2022ServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_shadowsocks_2022_proto_rawDescGZIP(), []int{0}
}

func (x *Shadowsocks2022ServerConfig) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Shadowsocks2022ServerConfig) GetUser() *configs.UserConfig {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *Shadowsocks2022ServerConfig) GetNetworks() []net.Network {
	if x != nil {
		return x.Networks
	}
	return nil
}

type Shadowsocks2022ClientConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Method        string                 `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	Key           string                 `protobuf:"bytes,4,opt,name=key,proto3" json:"key,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Shadowsocks2022ClientConfig) Reset() {
	*x = Shadowsocks2022ClientConfig{}
	mi := &file_protos_proxy_shadowsocks_2022_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Shadowsocks2022ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Shadowsocks2022ClientConfig) ProtoMessage() {}

func (x *Shadowsocks2022ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_shadowsocks_2022_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Shadowsocks2022ClientConfig.ProtoReflect.Descriptor instead.
func (*Shadowsocks2022ClientConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_shadowsocks_2022_proto_rawDescGZIP(), []int{1}
}

func (x *Shadowsocks2022ClientConfig) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Shadowsocks2022ClientConfig) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

var File_protos_proxy_shadowsocks_2022_proto protoreflect.FileDescriptor

const file_protos_proxy_shadowsocks_2022_proto_rawDesc = "" +
	"\n" +
	"#protos/proxy/shadowsocks_2022.proto\x12\ax.proxy\x1a\x11protos/user.proto\x1a\x14common/net/net.proto\"\x8b\x01\n" +
	"\x1bShadowsocks2022ServerConfig\x12\x16\n" +
	"\x06method\x18\x01 \x01(\tR\x06method\x12!\n" +
	"\x04user\x18\x02 \x01(\v2\r.x.UserConfigR\x04user\x121\n" +
	"\bnetworks\x18\x05 \x03(\x0e2\x15.x.common.net.NetworkR\bnetworks\"G\n" +
	"\x1bShadowsocks2022ClientConfig\x12\x16\n" +
	"\x06method\x18\x03 \x01(\tR\x06method\x12\x10\n" +
	"\x03key\x18\x04 \x01(\tR\x03keyB0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_shadowsocks_2022_proto_rawDescOnce sync.Once
	file_protos_proxy_shadowsocks_2022_proto_rawDescData []byte
)

func file_protos_proxy_shadowsocks_2022_proto_rawDescGZIP() []byte {
	file_protos_proxy_shadowsocks_2022_proto_rawDescOnce.Do(func() {
		file_protos_proxy_shadowsocks_2022_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_shadowsocks_2022_proto_rawDesc), len(file_protos_proxy_shadowsocks_2022_proto_rawDesc)))
	})
	return file_protos_proxy_shadowsocks_2022_proto_rawDescData
}

var file_protos_proxy_shadowsocks_2022_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protos_proxy_shadowsocks_2022_proto_goTypes = []any{
	(*Shadowsocks2022ServerConfig)(nil), // 0: x.proxy.Shadowsocks2022ServerConfig
	(*Shadowsocks2022ClientConfig)(nil), // 1: x.proxy.Shadowsocks2022ClientConfig
	(*configs.UserConfig)(nil),          // 2: x.UserConfig
	(net.Network)(0),                    // 3: x.common.net.Network
}
var file_protos_proxy_shadowsocks_2022_proto_depIdxs = []int32{
	2, // 0: x.proxy.Shadowsocks2022ServerConfig.user:type_name -> x.UserConfig
	3, // 1: x.proxy.Shadowsocks2022ServerConfig.networks:type_name -> x.common.net.Network
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_protos_proxy_shadowsocks_2022_proto_init() }
func file_protos_proxy_shadowsocks_2022_proto_init() {
	if File_protos_proxy_shadowsocks_2022_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_shadowsocks_2022_proto_rawDesc), len(file_protos_proxy_shadowsocks_2022_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_shadowsocks_2022_proto_goTypes,
		DependencyIndexes: file_protos_proxy_shadowsocks_2022_proto_depIdxs,
		MessageInfos:      file_protos_proxy_shadowsocks_2022_proto_msgTypes,
	}.Build()
	File_protos_proxy_shadowsocks_2022_proto = out.File
	file_protos_proxy_shadowsocks_2022_proto_goTypes = nil
	file_protos_proxy_shadowsocks_2022_proto_depIdxs = nil
}
