package splithttp

import (
	reality "github.com/5vnetwork/vx-core/transport/security/reality"
	tls "github.com/5vnetwork/vx-core/transport/security/tls"
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

type RangeConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	From          int32                  `protobuf:"varint,1,opt,name=from,proto3" json:"from,omitempty"`
	To            int32                  `protobuf:"varint,2,opt,name=to,proto3" json:"to,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RangeConfig) Reset() {
	*x = RangeConfig{}
	mi := &file_transport_protocols_splithttp_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RangeConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RangeConfig) ProtoMessage() {}

func (x *RangeConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_splithttp_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RangeConfig.ProtoReflect.Descriptor instead.
func (*RangeConfig) Descriptor() ([]byte, []int) {
	return file_transport_protocols_splithttp_config_proto_rawDescGZIP(), []int{0}
}

func (x *RangeConfig) GetFrom() int32 {
	if x != nil {
		return x.From
	}
	return 0
}

func (x *RangeConfig) GetTo() int32 {
	if x != nil {
		return x.To
	}
	return 0
}

type XmuxConfig struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	MaxConcurrency   *RangeConfig           `protobuf:"bytes,1,opt,name=maxConcurrency,proto3" json:"maxConcurrency,omitempty"`
	MaxConnections   *RangeConfig           `protobuf:"bytes,2,opt,name=maxConnections,proto3" json:"maxConnections,omitempty"`
	CMaxReuseTimes   *RangeConfig           `protobuf:"bytes,3,opt,name=cMaxReuseTimes,proto3" json:"cMaxReuseTimes,omitempty"`
	HMaxRequestTimes *RangeConfig           `protobuf:"bytes,4,opt,name=hMaxRequestTimes,proto3" json:"hMaxRequestTimes,omitempty"`
	HMaxReusableSecs *RangeConfig           `protobuf:"bytes,5,opt,name=hMaxReusableSecs,proto3" json:"hMaxReusableSecs,omitempty"`
	HKeepAlivePeriod int64                  `protobuf:"varint,6,opt,name=hKeepAlivePeriod,proto3" json:"hKeepAlivePeriod,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *XmuxConfig) Reset() {
	*x = XmuxConfig{}
	mi := &file_transport_protocols_splithttp_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *XmuxConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*XmuxConfig) ProtoMessage() {}

func (x *XmuxConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_splithttp_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use XmuxConfig.ProtoReflect.Descriptor instead.
func (*XmuxConfig) Descriptor() ([]byte, []int) {
	return file_transport_protocols_splithttp_config_proto_rawDescGZIP(), []int{1}
}

func (x *XmuxConfig) GetMaxConcurrency() *RangeConfig {
	if x != nil {
		return x.MaxConcurrency
	}
	return nil
}

func (x *XmuxConfig) GetMaxConnections() *RangeConfig {
	if x != nil {
		return x.MaxConnections
	}
	return nil
}

func (x *XmuxConfig) GetCMaxReuseTimes() *RangeConfig {
	if x != nil {
		return x.CMaxReuseTimes
	}
	return nil
}

func (x *XmuxConfig) GetHMaxRequestTimes() *RangeConfig {
	if x != nil {
		return x.HMaxRequestTimes
	}
	return nil
}

func (x *XmuxConfig) GetHMaxReusableSecs() *RangeConfig {
	if x != nil {
		return x.HMaxReusableSecs
	}
	return nil
}

func (x *XmuxConfig) GetHKeepAlivePeriod() int64 {
	if x != nil {
		return x.HKeepAlivePeriod
	}
	return 0
}

type SplitHttpConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Host          string                 `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Path          string                 `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Mode          string                 `protobuf:"bytes,3,opt,name=mode,proto3" json:"mode,omitempty"`
	Headers       map[string]string      `protobuf:"bytes,4,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	XPaddingBytes *RangeConfig           `protobuf:"bytes,5,opt,name=xPaddingBytes,proto3" json:"xPaddingBytes,omitempty"`
	NoGRPCHeader  bool                   `protobuf:"varint,6,opt,name=noGRPCHeader,proto3" json:"noGRPCHeader,omitempty"`
	// server only
	NoSSEHeader          bool         `protobuf:"varint,7,opt,name=noSSEHeader,proto3" json:"noSSEHeader,omitempty"`
	ScMaxEachPostBytes   *RangeConfig `protobuf:"bytes,8,opt,name=scMaxEachPostBytes,proto3" json:"scMaxEachPostBytes,omitempty"`
	ScMinPostsIntervalMs *RangeConfig `protobuf:"bytes,9,opt,name=scMinPostsIntervalMs,proto3" json:"scMinPostsIntervalMs,omitempty"`
	// server only
	ScMaxBufferedPosts int64 `protobuf:"varint,10,opt,name=scMaxBufferedPosts,proto3" json:"scMaxBufferedPosts,omitempty"`
	// server only
	ScStreamUpServerSecs *RangeConfig `protobuf:"bytes,11,opt,name=scStreamUpServerSecs,proto3" json:"scStreamUpServerSecs,omitempty"`
	Xmux                 *XmuxConfig  `protobuf:"bytes,12,opt,name=xmux,proto3" json:"xmux,omitempty"`
	DownloadSettings     *DownConfig  `protobuf:"bytes,13,opt,name=downloadSettings,proto3" json:"downloadSettings,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *SplitHttpConfig) Reset() {
	*x = SplitHttpConfig{}
	mi := &file_transport_protocols_splithttp_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SplitHttpConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SplitHttpConfig) ProtoMessage() {}

func (x *SplitHttpConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_splithttp_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SplitHttpConfig.ProtoReflect.Descriptor instead.
func (*SplitHttpConfig) Descriptor() ([]byte, []int) {
	return file_transport_protocols_splithttp_config_proto_rawDescGZIP(), []int{2}
}

func (x *SplitHttpConfig) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *SplitHttpConfig) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *SplitHttpConfig) GetMode() string {
	if x != nil {
		return x.Mode
	}
	return ""
}

func (x *SplitHttpConfig) GetHeaders() map[string]string {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *SplitHttpConfig) GetXPaddingBytes() *RangeConfig {
	if x != nil {
		return x.XPaddingBytes
	}
	return nil
}

func (x *SplitHttpConfig) GetNoGRPCHeader() bool {
	if x != nil {
		return x.NoGRPCHeader
	}
	return false
}

func (x *SplitHttpConfig) GetNoSSEHeader() bool {
	if x != nil {
		return x.NoSSEHeader
	}
	return false
}

func (x *SplitHttpConfig) GetScMaxEachPostBytes() *RangeConfig {
	if x != nil {
		return x.ScMaxEachPostBytes
	}
	return nil
}

func (x *SplitHttpConfig) GetScMinPostsIntervalMs() *RangeConfig {
	if x != nil {
		return x.ScMinPostsIntervalMs
	}
	return nil
}

func (x *SplitHttpConfig) GetScMaxBufferedPosts() int64 {
	if x != nil {
		return x.ScMaxBufferedPosts
	}
	return 0
}

func (x *SplitHttpConfig) GetScStreamUpServerSecs() *RangeConfig {
	if x != nil {
		return x.ScStreamUpServerSecs
	}
	return nil
}

func (x *SplitHttpConfig) GetXmux() *XmuxConfig {
	if x != nil {
		return x.Xmux
	}
	return nil
}

func (x *SplitHttpConfig) GetDownloadSettings() *DownConfig {
	if x != nil {
		return x.DownloadSettings
	}
	return nil
}

type DownConfig struct {
	state       protoimpl.MessageState `protogen:"open.v1"`
	Address     string                 `protobuf:"bytes,8,opt,name=address,proto3" json:"address,omitempty"`
	Port        uint32                 `protobuf:"varint,9,opt,name=port,proto3" json:"port,omitempty"`
	XhttpConfig *SplitHttpConfig       `protobuf:"bytes,2,opt,name=xhttp_config,json=xhttpConfig,proto3" json:"xhttp_config,omitempty"`
	// Types that are valid to be assigned to Security:
	//
	//	*DownConfig_Tls
	//	*DownConfig_Reality
	Security      isDownConfig_Security `protobuf_oneof:"security"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DownConfig) Reset() {
	*x = DownConfig{}
	mi := &file_transport_protocols_splithttp_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DownConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownConfig) ProtoMessage() {}

func (x *DownConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_protocols_splithttp_config_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownConfig.ProtoReflect.Descriptor instead.
func (*DownConfig) Descriptor() ([]byte, []int) {
	return file_transport_protocols_splithttp_config_proto_rawDescGZIP(), []int{3}
}

func (x *DownConfig) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *DownConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *DownConfig) GetXhttpConfig() *SplitHttpConfig {
	if x != nil {
		return x.XhttpConfig
	}
	return nil
}

func (x *DownConfig) GetSecurity() isDownConfig_Security {
	if x != nil {
		return x.Security
	}
	return nil
}

func (x *DownConfig) GetTls() *tls.TlsConfig {
	if x != nil {
		if x, ok := x.Security.(*DownConfig_Tls); ok {
			return x.Tls
		}
	}
	return nil
}

func (x *DownConfig) GetReality() *reality.RealityConfig {
	if x != nil {
		if x, ok := x.Security.(*DownConfig_Reality); ok {
			return x.Reality
		}
	}
	return nil
}

type isDownConfig_Security interface {
	isDownConfig_Security()
}

type DownConfig_Tls struct {
	Tls *tls.TlsConfig `protobuf:"bytes,20,opt,name=tls,proto3,oneof"`
}

type DownConfig_Reality struct {
	Reality *reality.RealityConfig `protobuf:"bytes,21,opt,name=reality,proto3,oneof"`
}

func (*DownConfig_Tls) isDownConfig_Security() {}

func (*DownConfig_Reality) isDownConfig_Security() {}

var File_transport_protocols_splithttp_config_proto protoreflect.FileDescriptor

const file_transport_protocols_splithttp_config_proto_rawDesc = "" +
	"\n" +
	"*transport/protocols/splithttp/config.proto\x12\x1fx.transport.protocols.splithttp\x1a\x14protos/tls/tls.proto\x1a'transport/security/reality/config.proto\"1\n" +
	"\vRangeConfig\x12\x12\n" +
	"\x04from\x18\x01 \x01(\x05R\x04from\x12\x0e\n" +
	"\x02to\x18\x02 \x01(\x05R\x02to\"\xee\x03\n" +
	"\n" +
	"XmuxConfig\x12T\n" +
	"\x0emaxConcurrency\x18\x01 \x01(\v2,.x.transport.protocols.splithttp.RangeConfigR\x0emaxConcurrency\x12T\n" +
	"\x0emaxConnections\x18\x02 \x01(\v2,.x.transport.protocols.splithttp.RangeConfigR\x0emaxConnections\x12T\n" +
	"\x0ecMaxReuseTimes\x18\x03 \x01(\v2,.x.transport.protocols.splithttp.RangeConfigR\x0ecMaxReuseTimes\x12X\n" +
	"\x10hMaxRequestTimes\x18\x04 \x01(\v2,.x.transport.protocols.splithttp.RangeConfigR\x10hMaxRequestTimes\x12X\n" +
	"\x10hMaxReusableSecs\x18\x05 \x01(\v2,.x.transport.protocols.splithttp.RangeConfigR\x10hMaxReusableSecs\x12*\n" +
	"\x10hKeepAlivePeriod\x18\x06 \x01(\x03R\x10hKeepAlivePeriod\"\xe8\x06\n" +
	"\x0fSplitHttpConfig\x12\x12\n" +
	"\x04host\x18\x01 \x01(\tR\x04host\x12\x12\n" +
	"\x04path\x18\x02 \x01(\tR\x04path\x12\x12\n" +
	"\x04mode\x18\x03 \x01(\tR\x04mode\x12W\n" +
	"\aheaders\x18\x04 \x03(\v2=.x.transport.protocols.splithttp.SplitHttpConfig.HeadersEntryR\aheaders\x12R\n" +
	"\rxPaddingBytes\x18\x05 \x01(\v2,.x.transport.protocols.splithttp.RangeConfigR\rxPaddingBytes\x12\"\n" +
	"\fnoGRPCHeader\x18\x06 \x01(\bR\fnoGRPCHeader\x12 \n" +
	"\vnoSSEHeader\x18\a \x01(\bR\vnoSSEHeader\x12\\\n" +
	"\x12scMaxEachPostBytes\x18\b \x01(\v2,.x.transport.protocols.splithttp.RangeConfigR\x12scMaxEachPostBytes\x12`\n" +
	"\x14scMinPostsIntervalMs\x18\t \x01(\v2,.x.transport.protocols.splithttp.RangeConfigR\x14scMinPostsIntervalMs\x12.\n" +
	"\x12scMaxBufferedPosts\x18\n" +
	" \x01(\x03R\x12scMaxBufferedPosts\x12`\n" +
	"\x14scStreamUpServerSecs\x18\v \x01(\v2,.x.transport.protocols.splithttp.RangeConfigR\x14scStreamUpServerSecs\x12?\n" +
	"\x04xmux\x18\f \x01(\v2+.x.transport.protocols.splithttp.XmuxConfigR\x04xmux\x12W\n" +
	"\x10downloadSettings\x18\r \x01(\v2+.x.transport.protocols.splithttp.DownConfigR\x10downloadSettings\x1a:\n" +
	"\fHeadersEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\x8a\x02\n" +
	"\n" +
	"DownConfig\x12\x18\n" +
	"\aaddress\x18\b \x01(\tR\aaddress\x12\x12\n" +
	"\x04port\x18\t \x01(\rR\x04port\x12S\n" +
	"\fxhttp_config\x18\x02 \x01(\v20.x.transport.protocols.splithttp.SplitHttpConfigR\vxhttpConfig\x12$\n" +
	"\x03tls\x18\x14 \x01(\v2\x10.x.tls.TlsConfigH\x00R\x03tls\x12G\n" +
	"\areality\x18\x15 \x01(\v2+.x.transport.security.reality.RealityConfigH\x00R\arealityB\n" +
	"\n" +
	"\bsecurityB<Z:github.com/5vnetwork/vx-core/transport/protocols/splithttpb\x06proto3"

var (
	file_transport_protocols_splithttp_config_proto_rawDescOnce sync.Once
	file_transport_protocols_splithttp_config_proto_rawDescData []byte
)

func file_transport_protocols_splithttp_config_proto_rawDescGZIP() []byte {
	file_transport_protocols_splithttp_config_proto_rawDescOnce.Do(func() {
		file_transport_protocols_splithttp_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_protocols_splithttp_config_proto_rawDesc), len(file_transport_protocols_splithttp_config_proto_rawDesc)))
	})
	return file_transport_protocols_splithttp_config_proto_rawDescData
}

var file_transport_protocols_splithttp_config_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_transport_protocols_splithttp_config_proto_goTypes = []any{
	(*RangeConfig)(nil),           // 0: x.transport.protocols.splithttp.RangeConfig
	(*XmuxConfig)(nil),            // 1: x.transport.protocols.splithttp.XmuxConfig
	(*SplitHttpConfig)(nil),       // 2: x.transport.protocols.splithttp.SplitHttpConfig
	(*DownConfig)(nil),            // 3: x.transport.protocols.splithttp.DownConfig
	nil,                           // 4: x.transport.protocols.splithttp.SplitHttpConfig.HeadersEntry
	(*tls.TlsConfig)(nil),         // 5: x.tls.TlsConfig
	(*reality.RealityConfig)(nil), // 6: x.transport.security.reality.RealityConfig
}
var file_transport_protocols_splithttp_config_proto_depIdxs = []int32{
	0,  // 0: x.transport.protocols.splithttp.XmuxConfig.maxConcurrency:type_name -> x.transport.protocols.splithttp.RangeConfig
	0,  // 1: x.transport.protocols.splithttp.XmuxConfig.maxConnections:type_name -> x.transport.protocols.splithttp.RangeConfig
	0,  // 2: x.transport.protocols.splithttp.XmuxConfig.cMaxReuseTimes:type_name -> x.transport.protocols.splithttp.RangeConfig
	0,  // 3: x.transport.protocols.splithttp.XmuxConfig.hMaxRequestTimes:type_name -> x.transport.protocols.splithttp.RangeConfig
	0,  // 4: x.transport.protocols.splithttp.XmuxConfig.hMaxReusableSecs:type_name -> x.transport.protocols.splithttp.RangeConfig
	4,  // 5: x.transport.protocols.splithttp.SplitHttpConfig.headers:type_name -> x.transport.protocols.splithttp.SplitHttpConfig.HeadersEntry
	0,  // 6: x.transport.protocols.splithttp.SplitHttpConfig.xPaddingBytes:type_name -> x.transport.protocols.splithttp.RangeConfig
	0,  // 7: x.transport.protocols.splithttp.SplitHttpConfig.scMaxEachPostBytes:type_name -> x.transport.protocols.splithttp.RangeConfig
	0,  // 8: x.transport.protocols.splithttp.SplitHttpConfig.scMinPostsIntervalMs:type_name -> x.transport.protocols.splithttp.RangeConfig
	0,  // 9: x.transport.protocols.splithttp.SplitHttpConfig.scStreamUpServerSecs:type_name -> x.transport.protocols.splithttp.RangeConfig
	1,  // 10: x.transport.protocols.splithttp.SplitHttpConfig.xmux:type_name -> x.transport.protocols.splithttp.XmuxConfig
	3,  // 11: x.transport.protocols.splithttp.SplitHttpConfig.downloadSettings:type_name -> x.transport.protocols.splithttp.DownConfig
	2,  // 12: x.transport.protocols.splithttp.DownConfig.xhttp_config:type_name -> x.transport.protocols.splithttp.SplitHttpConfig
	5,  // 13: x.transport.protocols.splithttp.DownConfig.tls:type_name -> x.tls.TlsConfig
	6,  // 14: x.transport.protocols.splithttp.DownConfig.reality:type_name -> x.transport.security.reality.RealityConfig
	15, // [15:15] is the sub-list for method output_type
	15, // [15:15] is the sub-list for method input_type
	15, // [15:15] is the sub-list for extension type_name
	15, // [15:15] is the sub-list for extension extendee
	0,  // [0:15] is the sub-list for field type_name
}

func init() { file_transport_protocols_splithttp_config_proto_init() }
func file_transport_protocols_splithttp_config_proto_init() {
	if File_transport_protocols_splithttp_config_proto != nil {
		return
	}
	file_transport_protocols_splithttp_config_proto_msgTypes[3].OneofWrappers = []any{
		(*DownConfig_Tls)(nil),
		(*DownConfig_Reality)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_protocols_splithttp_config_proto_rawDesc), len(file_transport_protocols_splithttp_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_protocols_splithttp_config_proto_goTypes,
		DependencyIndexes: file_transport_protocols_splithttp_config_proto_depIdxs,
		MessageInfos:      file_transport_protocols_splithttp_config_proto_msgTypes,
	}.Build()
	File_transport_protocols_splithttp_config_proto = out.File
	file_transport_protocols_splithttp_config_proto_goTypes = nil
	file_transport_protocols_splithttp_config_proto_depIdxs = nil
}
