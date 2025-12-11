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

type SysProxyConfig struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	HttpProxyAddress  string                 `protobuf:"bytes,1,opt,name=http_proxy_address,json=httpProxyAddress,proto3" json:"http_proxy_address,omitempty"`
	HttpProxyPort     uint32                 `protobuf:"varint,2,opt,name=http_proxy_port,json=httpProxyPort,proto3" json:"http_proxy_port,omitempty"`
	HttpsProxyAddress string                 `protobuf:"bytes,3,opt,name=https_proxy_address,json=httpsProxyAddress,proto3" json:"https_proxy_address,omitempty"`
	HttpsProxyPort    uint32                 `protobuf:"varint,4,opt,name=https_proxy_port,json=httpsProxyPort,proto3" json:"https_proxy_port,omitempty"`
	SocksProxyAddress string                 `protobuf:"bytes,5,opt,name=socks_proxy_address,json=socksProxyAddress,proto3" json:"socks_proxy_address,omitempty"`
	SocksProxyPort    uint32                 `protobuf:"varint,6,opt,name=socks_proxy_port,json=socksProxyPort,proto3" json:"socks_proxy_port,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *SysProxyConfig) Reset() {
	*x = SysProxyConfig{}
	mi := &file_protos_sysproxy_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SysProxyConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SysProxyConfig) ProtoMessage() {}

func (x *SysProxyConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_sysproxy_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SysProxyConfig.ProtoReflect.Descriptor instead.
func (*SysProxyConfig) Descriptor() ([]byte, []int) {
	return file_protos_sysproxy_proto_rawDescGZIP(), []int{0}
}

func (x *SysProxyConfig) GetHttpProxyAddress() string {
	if x != nil {
		return x.HttpProxyAddress
	}
	return ""
}

func (x *SysProxyConfig) GetHttpProxyPort() uint32 {
	if x != nil {
		return x.HttpProxyPort
	}
	return 0
}

func (x *SysProxyConfig) GetHttpsProxyAddress() string {
	if x != nil {
		return x.HttpsProxyAddress
	}
	return ""
}

func (x *SysProxyConfig) GetHttpsProxyPort() uint32 {
	if x != nil {
		return x.HttpsProxyPort
	}
	return 0
}

func (x *SysProxyConfig) GetSocksProxyAddress() string {
	if x != nil {
		return x.SocksProxyAddress
	}
	return ""
}

func (x *SysProxyConfig) GetSocksProxyPort() uint32 {
	if x != nil {
		return x.SocksProxyPort
	}
	return 0
}

var File_protos_sysproxy_proto protoreflect.FileDescriptor

const file_protos_sysproxy_proto_rawDesc = "" +
	"\n" +
	"\x15protos/sysproxy.proto\x12\x01x\"\x9a\x02\n" +
	"\x0eSysProxyConfig\x12,\n" +
	"\x12http_proxy_address\x18\x01 \x01(\tR\x10httpProxyAddress\x12&\n" +
	"\x0fhttp_proxy_port\x18\x02 \x01(\rR\rhttpProxyPort\x12.\n" +
	"\x13https_proxy_address\x18\x03 \x01(\tR\x11httpsProxyAddress\x12(\n" +
	"\x10https_proxy_port\x18\x04 \x01(\rR\x0ehttpsProxyPort\x12.\n" +
	"\x13socks_proxy_address\x18\x05 \x01(\tR\x11socksProxyAddress\x12(\n" +
	"\x10socks_proxy_port\x18\x06 \x01(\rR\x0esocksProxyPortB*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_sysproxy_proto_rawDescOnce sync.Once
	file_protos_sysproxy_proto_rawDescData []byte
)

func file_protos_sysproxy_proto_rawDescGZIP() []byte {
	file_protos_sysproxy_proto_rawDescOnce.Do(func() {
		file_protos_sysproxy_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_sysproxy_proto_rawDesc), len(file_protos_sysproxy_proto_rawDesc)))
	})
	return file_protos_sysproxy_proto_rawDescData
}

var file_protos_sysproxy_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_sysproxy_proto_goTypes = []any{
	(*SysProxyConfig)(nil), // 0: x.SysProxyConfig
}
var file_protos_sysproxy_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protos_sysproxy_proto_init() }
func file_protos_sysproxy_proto_init() {
	if File_protos_sysproxy_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_sysproxy_proto_rawDesc), len(file_protos_sysproxy_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_sysproxy_proto_goTypes,
		DependencyIndexes: file_protos_sysproxy_proto_depIdxs,
		MessageInfos:      file_protos_sysproxy_proto_msgTypes,
	}.Build()
	File_protos_sysproxy_proto = out.File
	file_protos_sysproxy_proto_goTypes = nil
	file_protos_sysproxy_proto_depIdxs = nil
}
