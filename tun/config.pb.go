package tun

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

type Config struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Defaults to "172.23.27.1/24". IP will be the ip address of the interface
	// Ending number determines OnLinkPrefixLength
	Ip4 string `protobuf:"bytes,1,opt,name=ip4,proto3" json:"ip4,omitempty"`
	// Defaults to "fd00::1/64". no zone.
	Ip6 string `protobuf:"bytes,2,opt,name=ip6,proto3" json:"ip6,omitempty"`
	// defaults to "vtun"
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// defaults to 8192
	Mtu uint32 `protobuf:"varint,4,opt,name=mtu,proto3" json:"mtu,omitempty"`
	// path to a dir containing four folders: amd64, x86, arm, arm64.
	// Each folder contains the wintun.dll for the corresponding platform.
	// Absolute or relative to cwd
	Path string `protobuf:"bytes,5,opt,name=path,proto3" json:"path,omitempty"`
	// If unset, dns nameserver address of the tun will be set to the next ip of
	// [ip4]. Same for dns6.
	Dns4 []string `protobuf:"bytes,6,rep,name=dns4,proto3" json:"dns4,omitempty"`
	Dns6 []string `protobuf:"bytes,7,rep,name=dns6,proto3" json:"dns6,omitempty"`
	// If set, the tun will only accept packets with the specified routes.
	StrictRoute bool `protobuf:"varint,8,opt,name=strict_route,json=strictRoute,proto3" json:"strict_route,omitempty"`
	// Only used when strict_route is set.
	Routes4       []string `protobuf:"bytes,9,rep,name=routes4,proto3" json:"routes4,omitempty"`
	Routes6       []string `protobuf:"bytes,10,rep,name=routes6,proto3" json:"routes6,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_tun_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_tun_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_tun_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetIp4() string {
	if x != nil {
		return x.Ip4
	}
	return ""
}

func (x *Config) GetIp6() string {
	if x != nil {
		return x.Ip6
	}
	return ""
}

func (x *Config) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Config) GetMtu() uint32 {
	if x != nil {
		return x.Mtu
	}
	return 0
}

func (x *Config) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Config) GetDns4() []string {
	if x != nil {
		return x.Dns4
	}
	return nil
}

func (x *Config) GetDns6() []string {
	if x != nil {
		return x.Dns6
	}
	return nil
}

func (x *Config) GetStrictRoute() bool {
	if x != nil {
		return x.StrictRoute
	}
	return false
}

func (x *Config) GetRoutes4() []string {
	if x != nil {
		return x.Routes4
	}
	return nil
}

func (x *Config) GetRoutes6() []string {
	if x != nil {
		return x.Routes6
	}
	return nil
}

type MonitorConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MonitorConfig) Reset() {
	*x = MonitorConfig{}
	mi := &file_tun_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MonitorConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MonitorConfig) ProtoMessage() {}

func (x *MonitorConfig) ProtoReflect() protoreflect.Message {
	mi := &file_tun_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MonitorConfig.ProtoReflect.Descriptor instead.
func (*MonitorConfig) Descriptor() ([]byte, []int) {
	return file_tun_config_proto_rawDescGZIP(), []int{1}
}

var File_tun_config_proto protoreflect.FileDescriptor

const file_tun_config_proto_rawDesc = "" +
	"\n" +
	"\x10tun/config.proto\x12\x05x.tun\"\xe5\x01\n" +
	"\x06Config\x12\x10\n" +
	"\x03ip4\x18\x01 \x01(\tR\x03ip4\x12\x10\n" +
	"\x03ip6\x18\x02 \x01(\tR\x03ip6\x12\x12\n" +
	"\x04name\x18\x03 \x01(\tR\x04name\x12\x10\n" +
	"\x03mtu\x18\x04 \x01(\rR\x03mtu\x12\x12\n" +
	"\x04path\x18\x05 \x01(\tR\x04path\x12\x12\n" +
	"\x04dns4\x18\x06 \x03(\tR\x04dns4\x12\x12\n" +
	"\x04dns6\x18\a \x03(\tR\x04dns6\x12!\n" +
	"\fstrict_route\x18\b \x01(\bR\vstrictRoute\x12\x18\n" +
	"\aroutes4\x18\t \x03(\tR\aroutes4\x12\x18\n" +
	"\aroutes6\x18\n" +
	" \x03(\tR\aroutes6\"\x0f\n" +
	"\rMonitorConfigB\"Z github.com/5vnetwork/vx-core/tunb\x06proto3"

var (
	file_tun_config_proto_rawDescOnce sync.Once
	file_tun_config_proto_rawDescData []byte
)

func file_tun_config_proto_rawDescGZIP() []byte {
	file_tun_config_proto_rawDescOnce.Do(func() {
		file_tun_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_tun_config_proto_rawDesc), len(file_tun_config_proto_rawDesc)))
	})
	return file_tun_config_proto_rawDescData
}

var file_tun_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_tun_config_proto_goTypes = []any{
	(*Config)(nil),        // 0: x.tun.Config
	(*MonitorConfig)(nil), // 1: x.tun.MonitorConfig
}
var file_tun_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_tun_config_proto_init() }
func file_tun_config_proto_init() {
	if File_tun_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_tun_config_proto_rawDesc), len(file_tun_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_tun_config_proto_goTypes,
		DependencyIndexes: file_tun_config_proto_depIdxs,
		MessageInfos:      file_tun_config_proto_msgTypes,
	}.Build()
	File_tun_config_proto = out.File
	file_tun_config_proto_goTypes = nil
	file_tun_config_proto_depIdxs = nil
}
