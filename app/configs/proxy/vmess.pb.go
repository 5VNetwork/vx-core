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

type SecurityType int32

const (
	SecurityType_SecurityType_UNKNOWN           SecurityType = 0
	SecurityType_SecurityType_LEGACY            SecurityType = 1
	SecurityType_SecurityType_AUTO              SecurityType = 2
	SecurityType_SecurityType_AES128_GCM        SecurityType = 3
	SecurityType_SecurityType_CHACHA20_POLY1305 SecurityType = 4
	SecurityType_SecurityType_NONE              SecurityType = 5
	SecurityType_SecurityType_ZERO              SecurityType = 6
)

// Enum value maps for SecurityType.
var (
	SecurityType_name = map[int32]string{
		0: "SecurityType_UNKNOWN",
		1: "SecurityType_LEGACY",
		2: "SecurityType_AUTO",
		3: "SecurityType_AES128_GCM",
		4: "SecurityType_CHACHA20_POLY1305",
		5: "SecurityType_NONE",
		6: "SecurityType_ZERO",
	}
	SecurityType_value = map[string]int32{
		"SecurityType_UNKNOWN":           0,
		"SecurityType_LEGACY":            1,
		"SecurityType_AUTO":              2,
		"SecurityType_AES128_GCM":        3,
		"SecurityType_CHACHA20_POLY1305": 4,
		"SecurityType_NONE":              5,
		"SecurityType_ZERO":              6,
	}
)

func (x SecurityType) Enum() *SecurityType {
	p := new(SecurityType)
	*p = x
	return p
}

func (x SecurityType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SecurityType) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_proxy_vmess_proto_enumTypes[0].Descriptor()
}

func (SecurityType) Type() protoreflect.EnumType {
	return &file_protos_proxy_vmess_proto_enumTypes[0]
}

func (x SecurityType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SecurityType.Descriptor instead.
func (SecurityType) EnumDescriptor() ([]byte, []int) {
	return file_protos_proxy_vmess_proto_rawDescGZIP(), []int{0}
}

type VmessClientConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
	Security      SecurityType           `protobuf:"varint,4,opt,name=security,proto3,enum=x.proxy.SecurityType" json:"security,omitempty"`
	Special       bool                   `protobuf:"varint,6,opt,name=special,proto3" json:"special,omitempty"`
	AlterId       uint32                 `protobuf:"varint,7,opt,name=alter_id,json=alterId,proto3" json:"alter_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VmessClientConfig) Reset() {
	*x = VmessClientConfig{}
	mi := &file_protos_proxy_vmess_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VmessClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VmessClientConfig) ProtoMessage() {}

func (x *VmessClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_vmess_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VmessClientConfig.ProtoReflect.Descriptor instead.
func (*VmessClientConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_vmess_proto_rawDescGZIP(), []int{0}
}

func (x *VmessClientConfig) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *VmessClientConfig) GetSecurity() SecurityType {
	if x != nil {
		return x.Security
	}
	return SecurityType_SecurityType_UNKNOWN
}

func (x *VmessClientConfig) GetSpecial() bool {
	if x != nil {
		return x.Special
	}
	return false
}

func (x *VmessClientConfig) GetAlterId() uint32 {
	if x != nil {
		return x.AlterId
	}
	return 0
}

type VmessServerConfig struct {
	state                protoimpl.MessageState `protogen:"open.v1"`
	Accounts             []*configs.UserConfig  `protobuf:"bytes,1,rep,name=accounts,proto3" json:"accounts,omitempty"`
	SecureEncryptionOnly bool                   `protobuf:"varint,4,opt,name=secure_encryption_only,json=secureEncryptionOnly,proto3" json:"secure_encryption_only,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *VmessServerConfig) Reset() {
	*x = VmessServerConfig{}
	mi := &file_protos_proxy_vmess_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VmessServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VmessServerConfig) ProtoMessage() {}

func (x *VmessServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_vmess_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VmessServerConfig.ProtoReflect.Descriptor instead.
func (*VmessServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_vmess_proto_rawDescGZIP(), []int{1}
}

func (x *VmessServerConfig) GetAccounts() []*configs.UserConfig {
	if x != nil {
		return x.Accounts
	}
	return nil
}

func (x *VmessServerConfig) GetSecureEncryptionOnly() bool {
	if x != nil {
		return x.SecureEncryptionOnly
	}
	return false
}

var File_protos_proxy_vmess_proto protoreflect.FileDescriptor

const file_protos_proxy_vmess_proto_rawDesc = "" +
	"\n" +
	"\x18protos/proxy/vmess.proto\x12\ax.proxy\x1a\x11protos/user.proto\"\x8b\x01\n" +
	"\x11VmessClientConfig\x12\x0e\n" +
	"\x02id\x18\x03 \x01(\tR\x02id\x121\n" +
	"\bsecurity\x18\x04 \x01(\x0e2\x15.x.proxy.SecurityTypeR\bsecurity\x12\x18\n" +
	"\aspecial\x18\x06 \x01(\bR\aspecial\x12\x19\n" +
	"\balter_id\x18\a \x01(\rR\aalterId\"t\n" +
	"\x11VmessServerConfig\x12)\n" +
	"\baccounts\x18\x01 \x03(\v2\r.x.UserConfigR\baccounts\x124\n" +
	"\x16secure_encryption_only\x18\x04 \x01(\bR\x14secureEncryptionOnly*\xc7\x01\n" +
	"\fSecurityType\x12\x18\n" +
	"\x14SecurityType_UNKNOWN\x10\x00\x12\x17\n" +
	"\x13SecurityType_LEGACY\x10\x01\x12\x15\n" +
	"\x11SecurityType_AUTO\x10\x02\x12\x1b\n" +
	"\x17SecurityType_AES128_GCM\x10\x03\x12\"\n" +
	"\x1eSecurityType_CHACHA20_POLY1305\x10\x04\x12\x15\n" +
	"\x11SecurityType_NONE\x10\x05\x12\x15\n" +
	"\x11SecurityType_ZERO\x10\x06B0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_vmess_proto_rawDescOnce sync.Once
	file_protos_proxy_vmess_proto_rawDescData []byte
)

func file_protos_proxy_vmess_proto_rawDescGZIP() []byte {
	file_protos_proxy_vmess_proto_rawDescOnce.Do(func() {
		file_protos_proxy_vmess_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_vmess_proto_rawDesc), len(file_protos_proxy_vmess_proto_rawDesc)))
	})
	return file_protos_proxy_vmess_proto_rawDescData
}

var file_protos_proxy_vmess_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protos_proxy_vmess_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protos_proxy_vmess_proto_goTypes = []any{
	(SecurityType)(0),          // 0: x.proxy.SecurityType
	(*VmessClientConfig)(nil),  // 1: x.proxy.VmessClientConfig
	(*VmessServerConfig)(nil),  // 2: x.proxy.VmessServerConfig
	(*configs.UserConfig)(nil), // 3: x.UserConfig
}
var file_protos_proxy_vmess_proto_depIdxs = []int32{
	0, // 0: x.proxy.VmessClientConfig.security:type_name -> x.proxy.SecurityType
	3, // 1: x.proxy.VmessServerConfig.accounts:type_name -> x.UserConfig
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_protos_proxy_vmess_proto_init() }
func file_protos_proxy_vmess_proto_init() {
	if File_protos_proxy_vmess_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_vmess_proto_rawDesc), len(file_protos_proxy_vmess_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_vmess_proto_goTypes,
		DependencyIndexes: file_protos_proxy_vmess_proto_depIdxs,
		EnumInfos:         file_protos_proxy_vmess_proto_enumTypes,
		MessageInfos:      file_protos_proxy_vmess_proto_msgTypes,
	}.Build()
	File_protos_proxy_vmess_proto = out.File
	file_protos_proxy_vmess_proto_goTypes = nil
	file_protos_proxy_vmess_proto_depIdxs = nil
}
