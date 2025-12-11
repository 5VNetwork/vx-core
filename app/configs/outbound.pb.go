package configs

import (
	net "github.com/5vnetwork/vx-core/common/net"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
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

type DomainStrategy int32

const (
	DomainStrategy_PreferIPv4 DomainStrategy = 0
	DomainStrategy_PreferIPv6 DomainStrategy = 1
	DomainStrategy_IPv4Only   DomainStrategy = 2
	DomainStrategy_IPv6Only   DomainStrategy = 43
)

// Enum value maps for DomainStrategy.
var (
	DomainStrategy_name = map[int32]string{
		0:  "PreferIPv4",
		1:  "PreferIPv6",
		2:  "IPv4Only",
		43: "IPv6Only",
	}
	DomainStrategy_value = map[string]int32{
		"PreferIPv4": 0,
		"PreferIPv6": 1,
		"IPv4Only":   2,
		"IPv6Only":   43,
	}
)

func (x DomainStrategy) Enum() *DomainStrategy {
	p := new(DomainStrategy)
	*p = x
	return p
}

func (x DomainStrategy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DomainStrategy) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_outbound_proto_enumTypes[0].Descriptor()
}

func (DomainStrategy) Type() protoreflect.EnumType {
	return &file_protos_outbound_proto_enumTypes[0]
}

func (x DomainStrategy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DomainStrategy.Descriptor instead.
func (DomainStrategy) EnumDescriptor() ([]byte, []int) {
	return file_protos_outbound_proto_rawDescGZIP(), []int{0}
}

type OutboundHandlerConfig struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	Tag            string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Transport      *TransportConfig       `protobuf:"bytes,3,opt,name=transport,proto3" json:"transport,omitempty"`
	EnableMux      bool                   `protobuf:"varint,4,opt,name=enable_mux,json=enableMux,proto3" json:"enable_mux,omitempty"`
	MuxConfig      *MuxConfig             `protobuf:"bytes,12,opt,name=mux_config,json=muxConfig,proto3" json:"mux_config,omitempty"`
	Address        string                 `protobuf:"bytes,5,opt,name=address,proto3" json:"address,omitempty"`
	Port           uint32                 `protobuf:"varint,6,opt,name=port,proto3" json:"port,omitempty"`
	Ports          []*net.PortRange       `protobuf:"bytes,7,rep,name=ports,proto3" json:"ports,omitempty"`
	Protocol       *anypb.Any             `protobuf:"bytes,8,opt,name=protocol,proto3" json:"protocol,omitempty"`
	Uot            bool                   `protobuf:"varint,9,opt,name=uot,proto3" json:"uot,omitempty"`
	DomainStrategy DomainStrategy         `protobuf:"varint,10,opt,name=domain_strategy,json=domainStrategy,proto3,enum=x.DomainStrategy" json:"domain_strategy,omitempty"` // bool pre_connect = 11;
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *OutboundHandlerConfig) Reset() {
	*x = OutboundHandlerConfig{}
	mi := &file_protos_outbound_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OutboundHandlerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OutboundHandlerConfig) ProtoMessage() {}

func (x *OutboundHandlerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_outbound_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OutboundHandlerConfig.ProtoReflect.Descriptor instead.
func (*OutboundHandlerConfig) Descriptor() ([]byte, []int) {
	return file_protos_outbound_proto_rawDescGZIP(), []int{0}
}

func (x *OutboundHandlerConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *OutboundHandlerConfig) GetTransport() *TransportConfig {
	if x != nil {
		return x.Transport
	}
	return nil
}

func (x *OutboundHandlerConfig) GetEnableMux() bool {
	if x != nil {
		return x.EnableMux
	}
	return false
}

func (x *OutboundHandlerConfig) GetMuxConfig() *MuxConfig {
	if x != nil {
		return x.MuxConfig
	}
	return nil
}

func (x *OutboundHandlerConfig) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *OutboundHandlerConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *OutboundHandlerConfig) GetPorts() []*net.PortRange {
	if x != nil {
		return x.Ports
	}
	return nil
}

func (x *OutboundHandlerConfig) GetProtocol() *anypb.Any {
	if x != nil {
		return x.Protocol
	}
	return nil
}

func (x *OutboundHandlerConfig) GetUot() bool {
	if x != nil {
		return x.Uot
	}
	return false
}

func (x *OutboundHandlerConfig) GetDomainStrategy() DomainStrategy {
	if x != nil {
		return x.DomainStrategy
	}
	return DomainStrategy_PreferIPv4
}

type MuxConfig struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	MaxConnection  uint32                 `protobuf:"varint,1,opt,name=max_connection,json=maxConnection,proto3" json:"max_connection,omitempty"`
	MaxConcurrency uint32                 `protobuf:"varint,2,opt,name=max_concurrency,json=maxConcurrency,proto3" json:"max_concurrency,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *MuxConfig) Reset() {
	*x = MuxConfig{}
	mi := &file_protos_outbound_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MuxConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MuxConfig) ProtoMessage() {}

func (x *MuxConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_outbound_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MuxConfig.ProtoReflect.Descriptor instead.
func (*MuxConfig) Descriptor() ([]byte, []int) {
	return file_protos_outbound_proto_rawDescGZIP(), []int{1}
}

func (x *MuxConfig) GetMaxConnection() uint32 {
	if x != nil {
		return x.MaxConnection
	}
	return 0
}

func (x *MuxConfig) GetMaxConcurrency() uint32 {
	if x != nil {
		return x.MaxConcurrency
	}
	return 0
}

type ChainHandlerConfig struct {
	state         protoimpl.MessageState   `protogen:"open.v1"`
	Handlers      []*OutboundHandlerConfig `protobuf:"bytes,1,rep,name=handlers,proto3" json:"handlers,omitempty"`
	Tag           string                   `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChainHandlerConfig) Reset() {
	*x = ChainHandlerConfig{}
	mi := &file_protos_outbound_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChainHandlerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChainHandlerConfig) ProtoMessage() {}

func (x *ChainHandlerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_outbound_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChainHandlerConfig.ProtoReflect.Descriptor instead.
func (*ChainHandlerConfig) Descriptor() ([]byte, []int) {
	return file_protos_outbound_proto_rawDescGZIP(), []int{2}
}

func (x *ChainHandlerConfig) GetHandlers() []*OutboundHandlerConfig {
	if x != nil {
		return x.Handlers
	}
	return nil
}

func (x *ChainHandlerConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

type OutboundConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// @deprecated
	OutboundHandlers []*OutboundHandlerConfig `protobuf:"bytes,1,rep,name=outbound_handlers,json=outboundHandlers,proto3" json:"outbound_handlers,omitempty"`
	// @deprecated
	ChainHandlers []*ChainHandlerConfig `protobuf:"bytes,2,rep,name=chain_handlers,json=chainHandlers,proto3" json:"chain_handlers,omitempty"`
	Handlers      []*HandlerConfig      `protobuf:"bytes,3,rep,name=handlers,proto3" json:"handlers,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OutboundConfig) Reset() {
	*x = OutboundConfig{}
	mi := &file_protos_outbound_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OutboundConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OutboundConfig) ProtoMessage() {}

func (x *OutboundConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_outbound_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OutboundConfig.ProtoReflect.Descriptor instead.
func (*OutboundConfig) Descriptor() ([]byte, []int) {
	return file_protos_outbound_proto_rawDescGZIP(), []int{3}
}

func (x *OutboundConfig) GetOutboundHandlers() []*OutboundHandlerConfig {
	if x != nil {
		return x.OutboundHandlers
	}
	return nil
}

func (x *OutboundConfig) GetChainHandlers() []*ChainHandlerConfig {
	if x != nil {
		return x.ChainHandlers
	}
	return nil
}

func (x *OutboundConfig) GetHandlers() []*HandlerConfig {
	if x != nil {
		return x.Handlers
	}
	return nil
}

type HandlerConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Type:
	//
	//	*HandlerConfig_Outbound
	//	*HandlerConfig_Chain
	Type          isHandlerConfig_Type `protobuf_oneof:"type"`
	SupportIpv6   *bool                `protobuf:"varint,3,opt,name=support_ipv6,json=supportIpv6,proto3,oneof" json:"support_ipv6,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerConfig) Reset() {
	*x = HandlerConfig{}
	mi := &file_protos_outbound_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerConfig) ProtoMessage() {}

func (x *HandlerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_outbound_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerConfig.ProtoReflect.Descriptor instead.
func (*HandlerConfig) Descriptor() ([]byte, []int) {
	return file_protos_outbound_proto_rawDescGZIP(), []int{4}
}

func (x *HandlerConfig) GetType() isHandlerConfig_Type {
	if x != nil {
		return x.Type
	}
	return nil
}

func (x *HandlerConfig) GetOutbound() *OutboundHandlerConfig {
	if x != nil {
		if x, ok := x.Type.(*HandlerConfig_Outbound); ok {
			return x.Outbound
		}
	}
	return nil
}

func (x *HandlerConfig) GetChain() *ChainHandlerConfig {
	if x != nil {
		if x, ok := x.Type.(*HandlerConfig_Chain); ok {
			return x.Chain
		}
	}
	return nil
}

func (x *HandlerConfig) GetSupportIpv6() bool {
	if x != nil && x.SupportIpv6 != nil {
		return *x.SupportIpv6
	}
	return false
}

type isHandlerConfig_Type interface {
	isHandlerConfig_Type()
}

type HandlerConfig_Outbound struct {
	Outbound *OutboundHandlerConfig `protobuf:"bytes,1,opt,name=outbound,proto3,oneof"`
}

type HandlerConfig_Chain struct {
	Chain *ChainHandlerConfig `protobuf:"bytes,2,opt,name=chain,proto3,oneof"`
}

func (*HandlerConfig_Outbound) isHandlerConfig_Type() {}

func (*HandlerConfig_Chain) isHandlerConfig_Type() {}

type HandlerConfigs struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Configs       []*HandlerConfig       `protobuf:"bytes,1,rep,name=configs,proto3" json:"configs,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandlerConfigs) Reset() {
	*x = HandlerConfigs{}
	mi := &file_protos_outbound_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandlerConfigs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandlerConfigs) ProtoMessage() {}

func (x *HandlerConfigs) ProtoReflect() protoreflect.Message {
	mi := &file_protos_outbound_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandlerConfigs.ProtoReflect.Descriptor instead.
func (*HandlerConfigs) Descriptor() ([]byte, []int) {
	return file_protos_outbound_proto_rawDescGZIP(), []int{5}
}

func (x *HandlerConfigs) GetConfigs() []*HandlerConfig {
	if x != nil {
		return x.Configs
	}
	return nil
}

var File_protos_outbound_proto protoreflect.FileDescriptor

const file_protos_outbound_proto_rawDesc = "" +
	"\n" +
	"\x15protos/outbound.proto\x12\x01x\x1a\x16protos/transport.proto\x1a\x14common/net/net.proto\x1a\x19google/protobuf/any.proto\"\x84\x03\n" +
	"\x15OutboundHandlerConfig\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x120\n" +
	"\ttransport\x18\x03 \x01(\v2\x12.x.TransportConfigR\ttransport\x12\x1d\n" +
	"\n" +
	"enable_mux\x18\x04 \x01(\bR\tenableMux\x12+\n" +
	"\n" +
	"mux_config\x18\f \x01(\v2\f.x.MuxConfigR\tmuxConfig\x12\x18\n" +
	"\aaddress\x18\x05 \x01(\tR\aaddress\x12\x12\n" +
	"\x04port\x18\x06 \x01(\rR\x04port\x12-\n" +
	"\x05ports\x18\a \x03(\v2\x17.x.common.net.PortRangeR\x05ports\x120\n" +
	"\bprotocol\x18\b \x01(\v2\x14.google.protobuf.AnyR\bprotocol\x12\x10\n" +
	"\x03uot\x18\t \x01(\bR\x03uot\x12:\n" +
	"\x0fdomain_strategy\x18\n" +
	" \x01(\x0e2\x11.x.DomainStrategyR\x0edomainStrategy\"[\n" +
	"\tMuxConfig\x12%\n" +
	"\x0emax_connection\x18\x01 \x01(\rR\rmaxConnection\x12'\n" +
	"\x0fmax_concurrency\x18\x02 \x01(\rR\x0emaxConcurrency\"\\\n" +
	"\x12ChainHandlerConfig\x124\n" +
	"\bhandlers\x18\x01 \x03(\v2\x18.x.OutboundHandlerConfigR\bhandlers\x12\x10\n" +
	"\x03tag\x18\x02 \x01(\tR\x03tag\"\xc3\x01\n" +
	"\x0eOutboundConfig\x12E\n" +
	"\x11outbound_handlers\x18\x01 \x03(\v2\x18.x.OutboundHandlerConfigR\x10outboundHandlers\x12<\n" +
	"\x0echain_handlers\x18\x02 \x03(\v2\x15.x.ChainHandlerConfigR\rchainHandlers\x12,\n" +
	"\bhandlers\x18\x03 \x03(\v2\x10.x.HandlerConfigR\bhandlers\"\xb7\x01\n" +
	"\rHandlerConfig\x126\n" +
	"\boutbound\x18\x01 \x01(\v2\x18.x.OutboundHandlerConfigH\x00R\boutbound\x12-\n" +
	"\x05chain\x18\x02 \x01(\v2\x15.x.ChainHandlerConfigH\x00R\x05chain\x12&\n" +
	"\fsupport_ipv6\x18\x03 \x01(\bH\x01R\vsupportIpv6\x88\x01\x01B\x06\n" +
	"\x04typeB\x0f\n" +
	"\r_support_ipv6\"<\n" +
	"\x0eHandlerConfigs\x12*\n" +
	"\aconfigs\x18\x01 \x03(\v2\x10.x.HandlerConfigR\aconfigs*L\n" +
	"\x0eDomainStrategy\x12\x0e\n" +
	"\n" +
	"PreferIPv4\x10\x00\x12\x0e\n" +
	"\n" +
	"PreferIPv6\x10\x01\x12\f\n" +
	"\bIPv4Only\x10\x02\x12\f\n" +
	"\bIPv6Only\x10+B*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_outbound_proto_rawDescOnce sync.Once
	file_protos_outbound_proto_rawDescData []byte
)

func file_protos_outbound_proto_rawDescGZIP() []byte {
	file_protos_outbound_proto_rawDescOnce.Do(func() {
		file_protos_outbound_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_outbound_proto_rawDesc), len(file_protos_outbound_proto_rawDesc)))
	})
	return file_protos_outbound_proto_rawDescData
}

var file_protos_outbound_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protos_outbound_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_protos_outbound_proto_goTypes = []any{
	(DomainStrategy)(0),           // 0: x.DomainStrategy
	(*OutboundHandlerConfig)(nil), // 1: x.OutboundHandlerConfig
	(*MuxConfig)(nil),             // 2: x.MuxConfig
	(*ChainHandlerConfig)(nil),    // 3: x.ChainHandlerConfig
	(*OutboundConfig)(nil),        // 4: x.OutboundConfig
	(*HandlerConfig)(nil),         // 5: x.HandlerConfig
	(*HandlerConfigs)(nil),        // 6: x.HandlerConfigs
	(*TransportConfig)(nil),       // 7: x.TransportConfig
	(*net.PortRange)(nil),         // 8: x.common.net.PortRange
	(*anypb.Any)(nil),             // 9: google.protobuf.Any
}
var file_protos_outbound_proto_depIdxs = []int32{
	7,  // 0: x.OutboundHandlerConfig.transport:type_name -> x.TransportConfig
	2,  // 1: x.OutboundHandlerConfig.mux_config:type_name -> x.MuxConfig
	8,  // 2: x.OutboundHandlerConfig.ports:type_name -> x.common.net.PortRange
	9,  // 3: x.OutboundHandlerConfig.protocol:type_name -> google.protobuf.Any
	0,  // 4: x.OutboundHandlerConfig.domain_strategy:type_name -> x.DomainStrategy
	1,  // 5: x.ChainHandlerConfig.handlers:type_name -> x.OutboundHandlerConfig
	1,  // 6: x.OutboundConfig.outbound_handlers:type_name -> x.OutboundHandlerConfig
	3,  // 7: x.OutboundConfig.chain_handlers:type_name -> x.ChainHandlerConfig
	5,  // 8: x.OutboundConfig.handlers:type_name -> x.HandlerConfig
	1,  // 9: x.HandlerConfig.outbound:type_name -> x.OutboundHandlerConfig
	3,  // 10: x.HandlerConfig.chain:type_name -> x.ChainHandlerConfig
	5,  // 11: x.HandlerConfigs.configs:type_name -> x.HandlerConfig
	12, // [12:12] is the sub-list for method output_type
	12, // [12:12] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_protos_outbound_proto_init() }
func file_protos_outbound_proto_init() {
	if File_protos_outbound_proto != nil {
		return
	}
	file_protos_transport_proto_init()
	file_protos_outbound_proto_msgTypes[4].OneofWrappers = []any{
		(*HandlerConfig_Outbound)(nil),
		(*HandlerConfig_Chain)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_outbound_proto_rawDesc), len(file_protos_outbound_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_outbound_proto_goTypes,
		DependencyIndexes: file_protos_outbound_proto_depIdxs,
		EnumInfos:         file_protos_outbound_proto_enumTypes,
		MessageInfos:      file_protos_outbound_proto_msgTypes,
	}.Build()
	File_protos_outbound_proto = out.File
	file_protos_outbound_proto_goTypes = nil
	file_protos_outbound_proto_depIdxs = nil
}
