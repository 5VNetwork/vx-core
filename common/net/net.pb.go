package net

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

type Network int32

const (
	Network_Unknown Network = 0
	Network_TCP     Network = 2
	Network_UDP     Network = 3
	Network_UNIX    Network = 4
)

// Enum value maps for Network.
var (
	Network_name = map[int32]string{
		0: "Unknown",
		2: "TCP",
		3: "UDP",
		4: "UNIX",
	}
	Network_value = map[string]int32{
		"Unknown": 0,
		"TCP":     2,
		"UDP":     3,
		"UNIX":    4,
	}
)

func (x Network) Enum() *Network {
	p := new(Network)
	*p = x
	return p
}

func (x Network) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Network) Descriptor() protoreflect.EnumDescriptor {
	return file_common_net_net_proto_enumTypes[0].Descriptor()
}

func (Network) Type() protoreflect.EnumType {
	return &file_common_net_net_proto_enumTypes[0]
}

func (x Network) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Network.Descriptor instead.
func (Network) EnumDescriptor() ([]byte, []int) {
	return file_common_net_net_proto_rawDescGZIP(), []int{0}
}

type PacketAddrType int32

const (
	PacketAddrType_None   PacketAddrType = 0
	PacketAddrType_Packet PacketAddrType = 1
)

// Enum value maps for PacketAddrType.
var (
	PacketAddrType_name = map[int32]string{
		0: "None",
		1: "Packet",
	}
	PacketAddrType_value = map[string]int32{
		"None":   0,
		"Packet": 1,
	}
)

func (x PacketAddrType) Enum() *PacketAddrType {
	p := new(PacketAddrType)
	*p = x
	return p
}

func (x PacketAddrType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PacketAddrType) Descriptor() protoreflect.EnumDescriptor {
	return file_common_net_net_proto_enumTypes[1].Descriptor()
}

func (PacketAddrType) Type() protoreflect.EnumType {
	return &file_common_net_net_proto_enumTypes[1]
}

func (x PacketAddrType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PacketAddrType.Descriptor instead.
func (PacketAddrType) EnumDescriptor() ([]byte, []int) {
	return file_common_net_net_proto_rawDescGZIP(), []int{1}
}

type NetworkList struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Network       []Network              `protobuf:"varint,1,rep,packed,name=network,proto3,enum=x.common.net.Network" json:"network,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *NetworkList) Reset() {
	*x = NetworkList{}
	mi := &file_common_net_net_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NetworkList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkList) ProtoMessage() {}

func (x *NetworkList) ProtoReflect() protoreflect.Message {
	mi := &file_common_net_net_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkList.ProtoReflect.Descriptor instead.
func (*NetworkList) Descriptor() ([]byte, []int) {
	return file_common_net_net_proto_rawDescGZIP(), []int{0}
}

func (x *NetworkList) GetNetwork() []Network {
	if x != nil {
		return x.Network
	}
	return nil
}

type Endpoint struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Network       Network                `protobuf:"varint,1,opt,name=network,proto3,enum=x.common.net.Network" json:"network,omitempty"`
	Address       string                 `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	Port          uint32                 `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Endpoint) Reset() {
	*x = Endpoint{}
	mi := &file_common_net_net_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Endpoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Endpoint) ProtoMessage() {}

func (x *Endpoint) ProtoReflect() protoreflect.Message {
	mi := &file_common_net_net_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Endpoint.ProtoReflect.Descriptor instead.
func (*Endpoint) Descriptor() ([]byte, []int) {
	return file_common_net_net_proto_rawDescGZIP(), []int{1}
}

func (x *Endpoint) GetNetwork() Network {
	if x != nil {
		return x.Network
	}
	return Network_Unknown
}

func (x *Endpoint) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *Endpoint) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

type IPPort struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Ip            string                 `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Port          uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *IPPort) Reset() {
	*x = IPPort{}
	mi := &file_common_net_net_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IPPort) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPPort) ProtoMessage() {}

func (x *IPPort) ProtoReflect() protoreflect.Message {
	mi := &file_common_net_net_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPPort.ProtoReflect.Descriptor instead.
func (*IPPort) Descriptor() ([]byte, []int) {
	return file_common_net_net_proto_rawDescGZIP(), []int{2}
}

func (x *IPPort) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *IPPort) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

// Address of a network host. It may be either an IP address or a domain
// address.
type IPOrDomain struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Address:
	//
	//	*IPOrDomain_Ip
	//	*IPOrDomain_Domain
	Address       isIPOrDomain_Address `protobuf_oneof:"address"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *IPOrDomain) Reset() {
	*x = IPOrDomain{}
	mi := &file_common_net_net_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IPOrDomain) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPOrDomain) ProtoMessage() {}

func (x *IPOrDomain) ProtoReflect() protoreflect.Message {
	mi := &file_common_net_net_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPOrDomain.ProtoReflect.Descriptor instead.
func (*IPOrDomain) Descriptor() ([]byte, []int) {
	return file_common_net_net_proto_rawDescGZIP(), []int{3}
}

func (x *IPOrDomain) GetAddress() isIPOrDomain_Address {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *IPOrDomain) GetIp() []byte {
	if x != nil {
		if x, ok := x.Address.(*IPOrDomain_Ip); ok {
			return x.Ip
		}
	}
	return nil
}

func (x *IPOrDomain) GetDomain() string {
	if x != nil {
		if x, ok := x.Address.(*IPOrDomain_Domain); ok {
			return x.Domain
		}
	}
	return ""
}

type isIPOrDomain_Address interface {
	isIPOrDomain_Address()
}

type IPOrDomain_Ip struct {
	// IP address. Must by either 4 or 16 bytes.
	Ip []byte `protobuf:"bytes,1,opt,name=ip,proto3,oneof"`
}

type IPOrDomain_Domain struct {
	// Domain address.
	Domain string `protobuf:"bytes,2,opt,name=domain,proto3,oneof"`
}

func (*IPOrDomain_Ip) isIPOrDomain_Address() {}

func (*IPOrDomain_Domain) isIPOrDomain_Address() {}

// PortRange represents a range of ports.
type PortRange struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The port that this range starts from.
	From uint32 `protobuf:"varint,1,opt,name=from,proto3" json:"from,omitempty"`
	// The port that this range ends with (inclusive).
	To            uint32 `protobuf:"varint,2,opt,name=to,proto3" json:"to,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PortRange) Reset() {
	*x = PortRange{}
	mi := &file_common_net_net_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PortRange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortRange) ProtoMessage() {}

func (x *PortRange) ProtoReflect() protoreflect.Message {
	mi := &file_common_net_net_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortRange.ProtoReflect.Descriptor instead.
func (*PortRange) Descriptor() ([]byte, []int) {
	return file_common_net_net_proto_rawDescGZIP(), []int{4}
}

func (x *PortRange) GetFrom() uint32 {
	if x != nil {
		return x.From
	}
	return 0
}

func (x *PortRange) GetTo() uint32 {
	if x != nil {
		return x.To
	}
	return 0
}

// PortList is a list of ports.
type PortList struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Range         []*PortRange           `protobuf:"bytes,1,rep,name=range,proto3" json:"range,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PortList) Reset() {
	*x = PortList{}
	mi := &file_common_net_net_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PortList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortList) ProtoMessage() {}

func (x *PortList) ProtoReflect() protoreflect.Message {
	mi := &file_common_net_net_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortList.ProtoReflect.Descriptor instead.
func (*PortList) Descriptor() ([]byte, []int) {
	return file_common_net_net_proto_rawDescGZIP(), []int{5}
}

func (x *PortList) GetRange() []*PortRange {
	if x != nil {
		return x.Range
	}
	return nil
}

var File_common_net_net_proto protoreflect.FileDescriptor

const file_common_net_net_proto_rawDesc = "" +
	"\n" +
	"\x14common/net/net.proto\x12\fx.common.net\">\n" +
	"\vNetworkList\x12/\n" +
	"\anetwork\x18\x01 \x03(\x0e2\x15.x.common.net.NetworkR\anetwork\"i\n" +
	"\bEndpoint\x12/\n" +
	"\anetwork\x18\x01 \x01(\x0e2\x15.x.common.net.NetworkR\anetwork\x12\x18\n" +
	"\aaddress\x18\x02 \x01(\tR\aaddress\x12\x12\n" +
	"\x04port\x18\x03 \x01(\rR\x04port\",\n" +
	"\x06IPPort\x12\x0e\n" +
	"\x02ip\x18\x01 \x01(\tR\x02ip\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\"C\n" +
	"\n" +
	"IPOrDomain\x12\x10\n" +
	"\x02ip\x18\x01 \x01(\fH\x00R\x02ip\x12\x18\n" +
	"\x06domain\x18\x02 \x01(\tH\x00R\x06domainB\t\n" +
	"\aaddress\"/\n" +
	"\tPortRange\x12\x12\n" +
	"\x04from\x18\x01 \x01(\rR\x04from\x12\x0e\n" +
	"\x02to\x18\x02 \x01(\rR\x02to\"9\n" +
	"\bPortList\x12-\n" +
	"\x05range\x18\x01 \x03(\v2\x17.x.common.net.PortRangeR\x05range*2\n" +
	"\aNetwork\x12\v\n" +
	"\aUnknown\x10\x00\x12\a\n" +
	"\x03TCP\x10\x02\x12\a\n" +
	"\x03UDP\x10\x03\x12\b\n" +
	"\x04UNIX\x10\x04*&\n" +
	"\x0ePacketAddrType\x12\b\n" +
	"\x04None\x10\x00\x12\n" +
	"\n" +
	"\x06Packet\x10\x01B)Z'github.com/5vnetwork/vx-core/common/netb\x06proto3"

var (
	file_common_net_net_proto_rawDescOnce sync.Once
	file_common_net_net_proto_rawDescData []byte
)

func file_common_net_net_proto_rawDescGZIP() []byte {
	file_common_net_net_proto_rawDescOnce.Do(func() {
		file_common_net_net_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_net_net_proto_rawDesc), len(file_common_net_net_proto_rawDesc)))
	})
	return file_common_net_net_proto_rawDescData
}

var file_common_net_net_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_common_net_net_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_common_net_net_proto_goTypes = []any{
	(Network)(0),        // 0: x.common.net.Network
	(PacketAddrType)(0), // 1: x.common.net.PacketAddrType
	(*NetworkList)(nil), // 2: x.common.net.NetworkList
	(*Endpoint)(nil),    // 3: x.common.net.Endpoint
	(*IPPort)(nil),      // 4: x.common.net.IPPort
	(*IPOrDomain)(nil),  // 5: x.common.net.IPOrDomain
	(*PortRange)(nil),   // 6: x.common.net.PortRange
	(*PortList)(nil),    // 7: x.common.net.PortList
}
var file_common_net_net_proto_depIdxs = []int32{
	0, // 0: x.common.net.NetworkList.network:type_name -> x.common.net.Network
	0, // 1: x.common.net.Endpoint.network:type_name -> x.common.net.Network
	6, // 2: x.common.net.PortList.range:type_name -> x.common.net.PortRange
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_common_net_net_proto_init() }
func file_common_net_net_proto_init() {
	if File_common_net_net_proto != nil {
		return
	}
	file_common_net_net_proto_msgTypes[3].OneofWrappers = []any{
		(*IPOrDomain_Ip)(nil),
		(*IPOrDomain_Domain)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_net_net_proto_rawDesc), len(file_common_net_net_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_net_net_proto_goTypes,
		DependencyIndexes: file_common_net_net_proto_depIdxs,
		EnumInfos:         file_common_net_net_proto_enumTypes,
		MessageInfos:      file_common_net_net_proto_msgTypes,
	}.Build()
	File_common_net_net_proto = out.File
	file_common_net_net_proto_goTypes = nil
	file_common_net_net_proto_depIdxs = nil
}
