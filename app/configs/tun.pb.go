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

type Mode int32

const (
	Mode_MODE_SYSTEM Mode = 0
	Mode_MODE_GVISOR Mode = 1
)

// Enum value maps for Mode.
var (
	Mode_name = map[int32]string{
		0: "MODE_SYSTEM",
		1: "MODE_GVISOR",
	}
	Mode_value = map[string]int32{
		"MODE_SYSTEM": 0,
		"MODE_GVISOR": 1,
	}
)

func (x Mode) Enum() *Mode {
	p := new(Mode)
	*p = x
	return p
}

func (x Mode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Mode) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_tun_proto_enumTypes[0].Descriptor()
}

func (Mode) Type() protoreflect.EnumType {
	return &file_protos_tun_proto_enumTypes[0]
}

func (x Mode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Mode.Descriptor instead.
func (Mode) EnumDescriptor() ([]byte, []int) {
	return file_protos_tun_proto_rawDescGZIP(), []int{0}
}

type TunConfig_TUN46Setting int32

const (
	TunConfig_FOUR_ONLY TunConfig_TUN46Setting = 0
	TunConfig_BOTH      TunConfig_TUN46Setting = 1
	TunConfig_DYNAMIC   TunConfig_TUN46Setting = 2
)

// Enum value maps for TunConfig_TUN46Setting.
var (
	TunConfig_TUN46Setting_name = map[int32]string{
		0: "FOUR_ONLY",
		1: "BOTH",
		2: "DYNAMIC",
	}
	TunConfig_TUN46Setting_value = map[string]int32{
		"FOUR_ONLY": 0,
		"BOTH":      1,
		"DYNAMIC":   2,
	}
)

func (x TunConfig_TUN46Setting) Enum() *TunConfig_TUN46Setting {
	p := new(TunConfig_TUN46Setting)
	*p = x
	return p
}

func (x TunConfig_TUN46Setting) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TunConfig_TUN46Setting) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_tun_proto_enumTypes[1].Descriptor()
}

func (TunConfig_TUN46Setting) Type() protoreflect.EnumType {
	return &file_protos_tun_proto_enumTypes[1]
}

func (x TunConfig_TUN46Setting) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TunConfig_TUN46Setting.Descriptor instead.
func (TunConfig_TUN46Setting) EnumDescriptor() ([]byte, []int) {
	return file_protos_tun_proto_rawDescGZIP(), []int{0, 0}
}

type TunConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Tag   string                 `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	Mode  Mode                   `protobuf:"varint,3,opt,name=mode,proto3,enum=x.Mode" json:"mode,omitempty"`
	// whether to bind outbound traffic to the primary physical interface
	ShouldBindDevice bool                   `protobuf:"varint,4,opt,name=should_bind_device,json=shouldBindDevice,proto3" json:"should_bind_device,omitempty"`
	Device           *TunDeviceConfig       `protobuf:"bytes,5,opt,name=device,proto3" json:"device,omitempty"`
	Tun46Setting     TunConfig_TUN46Setting `protobuf:"varint,8,opt,name=tun46_setting,json=tun46Setting,proto3,enum=x.TunConfig_TUN46Setting" json:"tun46_setting,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *TunConfig) Reset() {
	*x = TunConfig{}
	mi := &file_protos_tun_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TunConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TunConfig) ProtoMessage() {}

func (x *TunConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_tun_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TunConfig.ProtoReflect.Descriptor instead.
func (*TunConfig) Descriptor() ([]byte, []int) {
	return file_protos_tun_proto_rawDescGZIP(), []int{0}
}

func (x *TunConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *TunConfig) GetMode() Mode {
	if x != nil {
		return x.Mode
	}
	return Mode_MODE_SYSTEM
}

func (x *TunConfig) GetShouldBindDevice() bool {
	if x != nil {
		return x.ShouldBindDevice
	}
	return false
}

func (x *TunConfig) GetDevice() *TunDeviceConfig {
	if x != nil {
		return x.Device
	}
	return nil
}

func (x *TunConfig) GetTun46Setting() TunConfig_TUN46Setting {
	if x != nil {
		return x.Tun46Setting
	}
	return TunConfig_FOUR_ONLY
}

type TunDeviceConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Cidr4 string                 `protobuf:"bytes,1,opt,name=cidr4,proto3" json:"cidr4,omitempty"`
	Cidr6 string                 `protobuf:"bytes,2,opt,name=cidr6,proto3" json:"cidr6,omitempty"`
	Name  string                 `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Mtu   uint32                 `protobuf:"varint,4,opt,name=mtu,proto3" json:"mtu,omitempty"`
	// path to a dir containing four folders: amd64, x86, arm, arm64.
	// Each folder contains the wintun.dll for the corresponding platform.
	// Absolute or relative to cwd
	// Windows only
	Path    string   `protobuf:"bytes,5,opt,name=path,proto3" json:"path,omitempty"`
	Dns4    []string `protobuf:"bytes,6,rep,name=dns4,proto3" json:"dns4,omitempty"`
	Dns6    []string `protobuf:"bytes,7,rep,name=dns6,proto3" json:"dns6,omitempty"`
	Routes4 []string `protobuf:"bytes,9,rep,name=routes4,proto3" json:"routes4,omitempty"`
	Routes6 []string `protobuf:"bytes,10,rep,name=routes6,proto3" json:"routes6,omitempty"`
	// !Windows
	Fd uint32 `protobuf:"varint,11,opt,name=fd,proto3" json:"fd,omitempty"`
	// apps that does not use tun. Android only
	BlackListApps []string `protobuf:"bytes,12,rep,name=black_list_apps,json=blackListApps,proto3" json:"black_list_apps,omitempty"`
	// apps that use tun. Android only
	WhiteListApps []string `protobuf:"bytes,13,rep,name=white_list_apps,json=whiteListApps,proto3" json:"white_list_apps,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TunDeviceConfig) Reset() {
	*x = TunDeviceConfig{}
	mi := &file_protos_tun_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TunDeviceConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TunDeviceConfig) ProtoMessage() {}

func (x *TunDeviceConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_tun_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TunDeviceConfig.ProtoReflect.Descriptor instead.
func (*TunDeviceConfig) Descriptor() ([]byte, []int) {
	return file_protos_tun_proto_rawDescGZIP(), []int{1}
}

func (x *TunDeviceConfig) GetCidr4() string {
	if x != nil {
		return x.Cidr4
	}
	return ""
}

func (x *TunDeviceConfig) GetCidr6() string {
	if x != nil {
		return x.Cidr6
	}
	return ""
}

func (x *TunDeviceConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *TunDeviceConfig) GetMtu() uint32 {
	if x != nil {
		return x.Mtu
	}
	return 0
}

func (x *TunDeviceConfig) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *TunDeviceConfig) GetDns4() []string {
	if x != nil {
		return x.Dns4
	}
	return nil
}

func (x *TunDeviceConfig) GetDns6() []string {
	if x != nil {
		return x.Dns6
	}
	return nil
}

func (x *TunDeviceConfig) GetRoutes4() []string {
	if x != nil {
		return x.Routes4
	}
	return nil
}

func (x *TunDeviceConfig) GetRoutes6() []string {
	if x != nil {
		return x.Routes6
	}
	return nil
}

func (x *TunDeviceConfig) GetFd() uint32 {
	if x != nil {
		return x.Fd
	}
	return 0
}

func (x *TunDeviceConfig) GetBlackListApps() []string {
	if x != nil {
		return x.BlackListApps
	}
	return nil
}

func (x *TunDeviceConfig) GetWhiteListApps() []string {
	if x != nil {
		return x.WhiteListApps
	}
	return nil
}

var File_protos_tun_proto protoreflect.FileDescriptor

const file_protos_tun_proto_rawDesc = "" +
	"\n" +
	"\x10protos/tun.proto\x12\x01x\"\x8a\x02\n" +
	"\tTunConfig\x12\x10\n" +
	"\x03tag\x18\x02 \x01(\tR\x03tag\x12\x1b\n" +
	"\x04mode\x18\x03 \x01(\x0e2\a.x.ModeR\x04mode\x12,\n" +
	"\x12should_bind_device\x18\x04 \x01(\bR\x10shouldBindDevice\x12*\n" +
	"\x06device\x18\x05 \x01(\v2\x12.x.TunDeviceConfigR\x06device\x12>\n" +
	"\rtun46_setting\x18\b \x01(\x0e2\x19.x.TunConfig.TUN46SettingR\ftun46Setting\"4\n" +
	"\fTUN46Setting\x12\r\n" +
	"\tFOUR_ONLY\x10\x00\x12\b\n" +
	"\x04BOTH\x10\x01\x12\v\n" +
	"\aDYNAMIC\x10\x02\"\xb3\x02\n" +
	"\x0fTunDeviceConfig\x12\x14\n" +
	"\x05cidr4\x18\x01 \x01(\tR\x05cidr4\x12\x14\n" +
	"\x05cidr6\x18\x02 \x01(\tR\x05cidr6\x12\x12\n" +
	"\x04name\x18\x03 \x01(\tR\x04name\x12\x10\n" +
	"\x03mtu\x18\x04 \x01(\rR\x03mtu\x12\x12\n" +
	"\x04path\x18\x05 \x01(\tR\x04path\x12\x12\n" +
	"\x04dns4\x18\x06 \x03(\tR\x04dns4\x12\x12\n" +
	"\x04dns6\x18\a \x03(\tR\x04dns6\x12\x18\n" +
	"\aroutes4\x18\t \x03(\tR\aroutes4\x12\x18\n" +
	"\aroutes6\x18\n" +
	" \x03(\tR\aroutes6\x12\x0e\n" +
	"\x02fd\x18\v \x01(\rR\x02fd\x12&\n" +
	"\x0fblack_list_apps\x18\f \x03(\tR\rblackListApps\x12&\n" +
	"\x0fwhite_list_apps\x18\r \x03(\tR\rwhiteListApps*(\n" +
	"\x04Mode\x12\x0f\n" +
	"\vMODE_SYSTEM\x10\x00\x12\x0f\n" +
	"\vMODE_GVISOR\x10\x01B*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_tun_proto_rawDescOnce sync.Once
	file_protos_tun_proto_rawDescData []byte
)

func file_protos_tun_proto_rawDescGZIP() []byte {
	file_protos_tun_proto_rawDescOnce.Do(func() {
		file_protos_tun_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_tun_proto_rawDesc), len(file_protos_tun_proto_rawDesc)))
	})
	return file_protos_tun_proto_rawDescData
}

var file_protos_tun_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_protos_tun_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protos_tun_proto_goTypes = []any{
	(Mode)(0),                   // 0: x.Mode
	(TunConfig_TUN46Setting)(0), // 1: x.TunConfig.TUN46Setting
	(*TunConfig)(nil),           // 2: x.TunConfig
	(*TunDeviceConfig)(nil),     // 3: x.TunDeviceConfig
}
var file_protos_tun_proto_depIdxs = []int32{
	0, // 0: x.TunConfig.mode:type_name -> x.Mode
	3, // 1: x.TunConfig.device:type_name -> x.TunDeviceConfig
	1, // 2: x.TunConfig.tun46_setting:type_name -> x.TunConfig.TUN46Setting
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_protos_tun_proto_init() }
func file_protos_tun_proto_init() {
	if File_protos_tun_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_tun_proto_rawDesc), len(file_protos_tun_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_tun_proto_goTypes,
		DependencyIndexes: file_protos_tun_proto_depIdxs,
		EnumInfos:         file_protos_tun_proto_enumTypes,
		MessageInfos:      file_protos_tun_proto_msgTypes,
	}.Build()
	File_protos_tun_proto = out.File
	file_protos_tun_proto_goTypes = nil
	file_protos_tun_proto_depIdxs = nil
}
