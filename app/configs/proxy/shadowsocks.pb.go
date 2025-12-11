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

type ShadowsocksCipherType int32

const (
	ShadowsocksCipherType_AES_128_GCM       ShadowsocksCipherType = 0
	ShadowsocksCipherType_AES_256_GCM       ShadowsocksCipherType = 1
	ShadowsocksCipherType_CHACHA20_POLY1305 ShadowsocksCipherType = 2
	ShadowsocksCipherType_NONE              ShadowsocksCipherType = 3
)

// Enum value maps for ShadowsocksCipherType.
var (
	ShadowsocksCipherType_name = map[int32]string{
		0: "AES_128_GCM",
		1: "AES_256_GCM",
		2: "CHACHA20_POLY1305",
		3: "NONE",
	}
	ShadowsocksCipherType_value = map[string]int32{
		"AES_128_GCM":       0,
		"AES_256_GCM":       1,
		"CHACHA20_POLY1305": 2,
		"NONE":              3,
	}
)

func (x ShadowsocksCipherType) Enum() *ShadowsocksCipherType {
	p := new(ShadowsocksCipherType)
	*p = x
	return p
}

func (x ShadowsocksCipherType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ShadowsocksCipherType) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_proxy_shadowsocks_proto_enumTypes[0].Descriptor()
}

func (ShadowsocksCipherType) Type() protoreflect.EnumType {
	return &file_protos_proxy_shadowsocks_proto_enumTypes[0]
}

func (x ShadowsocksCipherType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ShadowsocksCipherType.Descriptor instead.
func (ShadowsocksCipherType) EnumDescriptor() ([]byte, []int) {
	return file_protos_proxy_shadowsocks_proto_rawDescGZIP(), []int{0}
}

type ShadowsocksAccount struct {
	state                          protoimpl.MessageState `protogen:"open.v1"`
	CipherType                     ShadowsocksCipherType  `protobuf:"varint,2,opt,name=cipher_type,json=cipherType,proto3,enum=x.proxy.ShadowsocksCipherType" json:"cipher_type,omitempty"`
	IvCheck                        bool                   `protobuf:"varint,3,opt,name=iv_check,json=ivCheck,proto3" json:"iv_check,omitempty"`
	User                           *configs.UserConfig    `protobuf:"bytes,4,opt,name=user,proto3" json:"user,omitempty"`
	ExperimentReducedIvHeadEntropy bool                   `protobuf:"varint,90001,opt,name=experiment_reduced_iv_head_entropy,json=experimentReducedIvHeadEntropy,proto3" json:"experiment_reduced_iv_head_entropy,omitempty"`
	unknownFields                  protoimpl.UnknownFields
	sizeCache                      protoimpl.SizeCache
}

func (x *ShadowsocksAccount) Reset() {
	*x = ShadowsocksAccount{}
	mi := &file_protos_proxy_shadowsocks_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ShadowsocksAccount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShadowsocksAccount) ProtoMessage() {}

func (x *ShadowsocksAccount) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_shadowsocks_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShadowsocksAccount.ProtoReflect.Descriptor instead.
func (*ShadowsocksAccount) Descriptor() ([]byte, []int) {
	return file_protos_proxy_shadowsocks_proto_rawDescGZIP(), []int{0}
}

func (x *ShadowsocksAccount) GetCipherType() ShadowsocksCipherType {
	if x != nil {
		return x.CipherType
	}
	return ShadowsocksCipherType_AES_128_GCM
}

func (x *ShadowsocksAccount) GetIvCheck() bool {
	if x != nil {
		return x.IvCheck
	}
	return false
}

func (x *ShadowsocksAccount) GetUser() *configs.UserConfig {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *ShadowsocksAccount) GetExperimentReducedIvHeadEntropy() bool {
	if x != nil {
		return x.ExperimentReducedIvHeadEntropy
	}
	return false
}

type ShadowsocksServerConfig struct {
	state                          protoimpl.MessageState `protogen:"open.v1"`
	CipherType                     ShadowsocksCipherType  `protobuf:"varint,3,opt,name=cipher_type,json=cipherType,proto3,enum=x.proxy.ShadowsocksCipherType" json:"cipher_type,omitempty"`
	IvCheck                        bool                   `protobuf:"varint,4,opt,name=iv_check,json=ivCheck,proto3" json:"iv_check,omitempty"`
	User                           *configs.UserConfig    `protobuf:"bytes,5,opt,name=user,proto3" json:"user,omitempty"`
	ExperimentReducedIvHeadEntropy bool                   `protobuf:"varint,90001,opt,name=experiment_reduced_iv_head_entropy,json=experimentReducedIvHeadEntropy,proto3" json:"experiment_reduced_iv_head_entropy,omitempty"`
	unknownFields                  protoimpl.UnknownFields
	sizeCache                      protoimpl.SizeCache
}

func (x *ShadowsocksServerConfig) Reset() {
	*x = ShadowsocksServerConfig{}
	mi := &file_protos_proxy_shadowsocks_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ShadowsocksServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShadowsocksServerConfig) ProtoMessage() {}

func (x *ShadowsocksServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_shadowsocks_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShadowsocksServerConfig.ProtoReflect.Descriptor instead.
func (*ShadowsocksServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_shadowsocks_proto_rawDescGZIP(), []int{1}
}

func (x *ShadowsocksServerConfig) GetCipherType() ShadowsocksCipherType {
	if x != nil {
		return x.CipherType
	}
	return ShadowsocksCipherType_AES_128_GCM
}

func (x *ShadowsocksServerConfig) GetIvCheck() bool {
	if x != nil {
		return x.IvCheck
	}
	return false
}

func (x *ShadowsocksServerConfig) GetUser() *configs.UserConfig {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *ShadowsocksServerConfig) GetExperimentReducedIvHeadEntropy() bool {
	if x != nil {
		return x.ExperimentReducedIvHeadEntropy
	}
	return false
}

type ShadowsocksClientConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	CipherType    ShadowsocksCipherType  `protobuf:"varint,1,opt,name=cipher_type,json=cipherType,proto3,enum=x.proxy.ShadowsocksCipherType" json:"cipher_type,omitempty"`
	Password      string                 `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ShadowsocksClientConfig) Reset() {
	*x = ShadowsocksClientConfig{}
	mi := &file_protos_proxy_shadowsocks_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ShadowsocksClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShadowsocksClientConfig) ProtoMessage() {}

func (x *ShadowsocksClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_shadowsocks_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShadowsocksClientConfig.ProtoReflect.Descriptor instead.
func (*ShadowsocksClientConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_shadowsocks_proto_rawDescGZIP(), []int{2}
}

func (x *ShadowsocksClientConfig) GetCipherType() ShadowsocksCipherType {
	if x != nil {
		return x.CipherType
	}
	return ShadowsocksCipherType_AES_128_GCM
}

func (x *ShadowsocksClientConfig) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

type Shadowsocks2022ClientConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Method        string                 `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	Psk           []byte                 `protobuf:"bytes,2,opt,name=psk,proto3" json:"psk,omitempty"`
	Ipsk          [][]byte               `protobuf:"bytes,4,rep,name=ipsk,proto3" json:"ipsk,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Shadowsocks2022ClientConfig) Reset() {
	*x = Shadowsocks2022ClientConfig{}
	mi := &file_protos_proxy_shadowsocks_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Shadowsocks2022ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Shadowsocks2022ClientConfig) ProtoMessage() {}

func (x *Shadowsocks2022ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_shadowsocks_proto_msgTypes[3]
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
	return file_protos_proxy_shadowsocks_proto_rawDescGZIP(), []int{3}
}

func (x *Shadowsocks2022ClientConfig) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Shadowsocks2022ClientConfig) GetPsk() []byte {
	if x != nil {
		return x.Psk
	}
	return nil
}

func (x *Shadowsocks2022ClientConfig) GetIpsk() [][]byte {
	if x != nil {
		return x.Ipsk
	}
	return nil
}

var File_protos_proxy_shadowsocks_proto protoreflect.FileDescriptor

const file_protos_proxy_shadowsocks_proto_rawDesc = "" +
	"\n" +
	"\x1eprotos/proxy/shadowsocks.proto\x12\ax.proxy\x1a\x11protos/user.proto\"\xe1\x01\n" +
	"\x12ShadowsocksAccount\x12?\n" +
	"\vcipher_type\x18\x02 \x01(\x0e2\x1e.x.proxy.ShadowsocksCipherTypeR\n" +
	"cipherType\x12\x19\n" +
	"\biv_check\x18\x03 \x01(\bR\aivCheck\x12!\n" +
	"\x04user\x18\x04 \x01(\v2\r.x.UserConfigR\x04user\x12L\n" +
	"\"experiment_reduced_iv_head_entropy\x18\x91\xbf\x05 \x01(\bR\x1eexperimentReducedIvHeadEntropy\"\xec\x01\n" +
	"\x17ShadowsocksServerConfig\x12?\n" +
	"\vcipher_type\x18\x03 \x01(\x0e2\x1e.x.proxy.ShadowsocksCipherTypeR\n" +
	"cipherType\x12\x19\n" +
	"\biv_check\x18\x04 \x01(\bR\aivCheck\x12!\n" +
	"\x04user\x18\x05 \x01(\v2\r.x.UserConfigR\x04user\x12L\n" +
	"\"experiment_reduced_iv_head_entropy\x18\x91\xbf\x05 \x01(\bR\x1eexperimentReducedIvHeadEntropyJ\x04\b\x02\x10\x03\"v\n" +
	"\x17ShadowsocksClientConfig\x12?\n" +
	"\vcipher_type\x18\x01 \x01(\x0e2\x1e.x.proxy.ShadowsocksCipherTypeR\n" +
	"cipherType\x12\x1a\n" +
	"\bpassword\x18\x02 \x01(\tR\bpassword\"[\n" +
	"\x1bShadowsocks2022ClientConfig\x12\x16\n" +
	"\x06method\x18\x01 \x01(\tR\x06method\x12\x10\n" +
	"\x03psk\x18\x02 \x01(\fR\x03psk\x12\x12\n" +
	"\x04ipsk\x18\x04 \x03(\fR\x04ipsk*Z\n" +
	"\x15ShadowsocksCipherType\x12\x0f\n" +
	"\vAES_128_GCM\x10\x00\x12\x0f\n" +
	"\vAES_256_GCM\x10\x01\x12\x15\n" +
	"\x11CHACHA20_POLY1305\x10\x02\x12\b\n" +
	"\x04NONE\x10\x03B0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_shadowsocks_proto_rawDescOnce sync.Once
	file_protos_proxy_shadowsocks_proto_rawDescData []byte
)

func file_protos_proxy_shadowsocks_proto_rawDescGZIP() []byte {
	file_protos_proxy_shadowsocks_proto_rawDescOnce.Do(func() {
		file_protos_proxy_shadowsocks_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_shadowsocks_proto_rawDesc), len(file_protos_proxy_shadowsocks_proto_rawDesc)))
	})
	return file_protos_proxy_shadowsocks_proto_rawDescData
}

var file_protos_proxy_shadowsocks_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protos_proxy_shadowsocks_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_protos_proxy_shadowsocks_proto_goTypes = []any{
	(ShadowsocksCipherType)(0),          // 0: x.proxy.ShadowsocksCipherType
	(*ShadowsocksAccount)(nil),          // 1: x.proxy.ShadowsocksAccount
	(*ShadowsocksServerConfig)(nil),     // 2: x.proxy.ShadowsocksServerConfig
	(*ShadowsocksClientConfig)(nil),     // 3: x.proxy.ShadowsocksClientConfig
	(*Shadowsocks2022ClientConfig)(nil), // 4: x.proxy.Shadowsocks2022ClientConfig
	(*configs.UserConfig)(nil),          // 5: x.UserConfig
}
var file_protos_proxy_shadowsocks_proto_depIdxs = []int32{
	0, // 0: x.proxy.ShadowsocksAccount.cipher_type:type_name -> x.proxy.ShadowsocksCipherType
	5, // 1: x.proxy.ShadowsocksAccount.user:type_name -> x.UserConfig
	0, // 2: x.proxy.ShadowsocksServerConfig.cipher_type:type_name -> x.proxy.ShadowsocksCipherType
	5, // 3: x.proxy.ShadowsocksServerConfig.user:type_name -> x.UserConfig
	0, // 4: x.proxy.ShadowsocksClientConfig.cipher_type:type_name -> x.proxy.ShadowsocksCipherType
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_protos_proxy_shadowsocks_proto_init() }
func file_protos_proxy_shadowsocks_proto_init() {
	if File_protos_proxy_shadowsocks_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_shadowsocks_proto_rawDesc), len(file_protos_proxy_shadowsocks_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_shadowsocks_proto_goTypes,
		DependencyIndexes: file_protos_proxy_shadowsocks_proto_depIdxs,
		EnumInfos:         file_protos_proxy_shadowsocks_proto_enumTypes,
		MessageInfos:      file_protos_proxy_shadowsocks_proto_msgTypes,
	}.Build()
	File_protos_proxy_shadowsocks_proto = out.File
	file_protos_proxy_shadowsocks_proto_goTypes = nil
	file_protos_proxy_shadowsocks_proto_depIdxs = nil
}
