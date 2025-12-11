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

type AuthType int32

const (
	// NO_AUTH is for anonymous authentication.
	AuthType_NO_AUTH AuthType = 0
	// PASSWORD is for username/password authentication.
	AuthType_PASSWORD AuthType = 1
)

// Enum value maps for AuthType.
var (
	AuthType_name = map[int32]string{
		0: "NO_AUTH",
		1: "PASSWORD",
	}
	AuthType_value = map[string]int32{
		"NO_AUTH":  0,
		"PASSWORD": 1,
	}
)

func (x AuthType) Enum() *AuthType {
	p := new(AuthType)
	*p = x
	return p
}

func (x AuthType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AuthType) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_proxy_socks_proto_enumTypes[0].Descriptor()
}

func (AuthType) Type() protoreflect.EnumType {
	return &file_protos_proxy_socks_proto_enumTypes[0]
}

func (x AuthType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AuthType.Descriptor instead.
func (AuthType) EnumDescriptor() ([]byte, []int) {
	return file_protos_proxy_socks_proto_rawDescGZIP(), []int{0}
}

// ServerConfig is the protobuf config for Socks server.
type SocksServerConfig struct {
	state    protoimpl.MessageState `protogen:"open.v1"`
	AuthType AuthType               `protobuf:"varint,1,opt,name=auth_type,json=authType,proto3,enum=x.proxy.AuthType" json:"auth_type,omitempty"`
	Accounts []*configs.UserConfig  `protobuf:"bytes,2,rep,name=accounts,proto3" json:"accounts,omitempty"`
	// used as BND.ADDR of the reply to a socks request
	Address       string `protobuf:"bytes,3,opt,name=address,proto3" json:"address,omitempty"`
	UdpEnabled    bool   `protobuf:"varint,4,opt,name=udp_enabled,json=udpEnabled,proto3" json:"udp_enabled,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SocksServerConfig) Reset() {
	*x = SocksServerConfig{}
	mi := &file_protos_proxy_socks_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SocksServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SocksServerConfig) ProtoMessage() {}

func (x *SocksServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_socks_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SocksServerConfig.ProtoReflect.Descriptor instead.
func (*SocksServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_socks_proto_rawDescGZIP(), []int{0}
}

func (x *SocksServerConfig) GetAuthType() AuthType {
	if x != nil {
		return x.AuthType
	}
	return AuthType_NO_AUTH
}

func (x *SocksServerConfig) GetAccounts() []*configs.UserConfig {
	if x != nil {
		return x.Accounts
	}
	return nil
}

func (x *SocksServerConfig) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *SocksServerConfig) GetUdpEnabled() bool {
	if x != nil {
		return x.UdpEnabled
	}
	return false
}

// ClientConfig is the protobuf config for Socks client.
type SocksClientConfig struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	Name           string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Password       string                 `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	DelayAuthWrite bool                   `protobuf:"varint,3,opt,name=delay_auth_write,json=delayAuthWrite,proto3" json:"delay_auth_write,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *SocksClientConfig) Reset() {
	*x = SocksClientConfig{}
	mi := &file_protos_proxy_socks_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SocksClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SocksClientConfig) ProtoMessage() {}

func (x *SocksClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_socks_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SocksClientConfig.ProtoReflect.Descriptor instead.
func (*SocksClientConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_socks_proto_rawDescGZIP(), []int{1}
}

func (x *SocksClientConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SocksClientConfig) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *SocksClientConfig) GetDelayAuthWrite() bool {
	if x != nil {
		return x.DelayAuthWrite
	}
	return false
}

var File_protos_proxy_socks_proto protoreflect.FileDescriptor

const file_protos_proxy_socks_proto_rawDesc = "" +
	"\n" +
	"\x18protos/proxy/socks.proto\x12\ax.proxy\x1a\x11protos/user.proto\"\xa9\x01\n" +
	"\x11SocksServerConfig\x12.\n" +
	"\tauth_type\x18\x01 \x01(\x0e2\x11.x.proxy.AuthTypeR\bauthType\x12)\n" +
	"\baccounts\x18\x02 \x03(\v2\r.x.UserConfigR\baccounts\x12\x18\n" +
	"\aaddress\x18\x03 \x01(\tR\aaddress\x12\x1f\n" +
	"\vudp_enabled\x18\x04 \x01(\bR\n" +
	"udpEnabled\"m\n" +
	"\x11SocksClientConfig\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x1a\n" +
	"\bpassword\x18\x02 \x01(\tR\bpassword\x12(\n" +
	"\x10delay_auth_write\x18\x03 \x01(\bR\x0edelayAuthWrite*%\n" +
	"\bAuthType\x12\v\n" +
	"\aNO_AUTH\x10\x00\x12\f\n" +
	"\bPASSWORD\x10\x01B0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_socks_proto_rawDescOnce sync.Once
	file_protos_proxy_socks_proto_rawDescData []byte
)

func file_protos_proxy_socks_proto_rawDescGZIP() []byte {
	file_protos_proxy_socks_proto_rawDescOnce.Do(func() {
		file_protos_proxy_socks_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_socks_proto_rawDesc), len(file_protos_proxy_socks_proto_rawDesc)))
	})
	return file_protos_proxy_socks_proto_rawDescData
}

var file_protos_proxy_socks_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protos_proxy_socks_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protos_proxy_socks_proto_goTypes = []any{
	(AuthType)(0),              // 0: x.proxy.AuthType
	(*SocksServerConfig)(nil),  // 1: x.proxy.SocksServerConfig
	(*SocksClientConfig)(nil),  // 2: x.proxy.SocksClientConfig
	(*configs.UserConfig)(nil), // 3: x.UserConfig
}
var file_protos_proxy_socks_proto_depIdxs = []int32{
	0, // 0: x.proxy.SocksServerConfig.auth_type:type_name -> x.proxy.AuthType
	3, // 1: x.proxy.SocksServerConfig.accounts:type_name -> x.UserConfig
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_protos_proxy_socks_proto_init() }
func file_protos_proxy_socks_proto_init() {
	if File_protos_proxy_socks_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_socks_proto_rawDesc), len(file_protos_proxy_socks_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_socks_proto_goTypes,
		DependencyIndexes: file_protos_proxy_socks_proto_depIdxs,
		EnumInfos:         file_protos_proxy_socks_proto_enumTypes,
		MessageInfos:      file_protos_proxy_socks_proto_msgTypes,
	}.Build()
	File_protos_proxy_socks_proto = out.File
	file_protos_proxy_socks_proto_goTypes = nil
	file_protos_proxy_socks_proto_depIdxs = nil
}
