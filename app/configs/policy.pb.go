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

type PolicyConfig struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	UserPolicyMap    map[uint32]*UserPolicy `protobuf:"bytes,11,rep,name=user_policy_map,json=userPolicyMap,proto3" json:"user_policy_map,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	HandshakeTimeout uint32                 `protobuf:"varint,1,opt,name=handshake_timeout,json=handshakeTimeout,proto3" json:"handshake_timeout,omitempty"`
	// tcp
	ConnectionIdleTimeout uint32 `protobuf:"varint,2,opt,name=connection_idle_timeout,json=connectionIdleTimeout,proto3" json:"connection_idle_timeout,omitempty"`
	UpLinkOnlyTimeout     uint32 `protobuf:"varint,4,opt,name=upLink_only_timeout,json=upLinkOnlyTimeout,proto3" json:"upLink_only_timeout,omitempty"`
	DownLinkOnlyTimeout   uint32 `protobuf:"varint,5,opt,name=downLink_only_timeout,json=downLinkOnlyTimeout,proto3" json:"downLink_only_timeout,omitempty"`
	// udp
	UdpIdleTimeout uint32 `protobuf:"varint,3,opt,name=udp_idle_timeout,json=udpIdleTimeout,proto3" json:"udp_idle_timeout,omitempty"`
	// inbound link stats
	LinkStats bool `protobuf:"varint,7,opt,name=link_stats,json=linkStats,proto3" json:"link_stats,omitempty"`
	// inbound stats
	InboundStats  bool `protobuf:"varint,8,opt,name=inbound_stats,json=inboundStats,proto3" json:"inbound_stats,omitempty"`
	UserStats     bool `protobuf:"varint,9,opt,name=user_stats,json=userStats,proto3" json:"user_stats,omitempty"`
	OutboundStats bool `protobuf:"varint,10,opt,name=outbound_stats,json=outboundStats,proto3" json:"outbound_stats,omitempty"`
	// for debug purpose
	SessionStats      bool  `protobuf:"varint,13,opt,name=session_stats,json=sessionStats,proto3" json:"session_stats,omitempty"`
	DefaultBufferSize int32 `protobuf:"varint,12,opt,name=default_buffer_size,json=defaultBufferSize,proto3" json:"default_buffer_size,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *PolicyConfig) Reset() {
	*x = PolicyConfig{}
	mi := &file_protos_policy_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PolicyConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PolicyConfig) ProtoMessage() {}

func (x *PolicyConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_policy_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PolicyConfig.ProtoReflect.Descriptor instead.
func (*PolicyConfig) Descriptor() ([]byte, []int) {
	return file_protos_policy_proto_rawDescGZIP(), []int{0}
}

func (x *PolicyConfig) GetUserPolicyMap() map[uint32]*UserPolicy {
	if x != nil {
		return x.UserPolicyMap
	}
	return nil
}

func (x *PolicyConfig) GetHandshakeTimeout() uint32 {
	if x != nil {
		return x.HandshakeTimeout
	}
	return 0
}

func (x *PolicyConfig) GetConnectionIdleTimeout() uint32 {
	if x != nil {
		return x.ConnectionIdleTimeout
	}
	return 0
}

func (x *PolicyConfig) GetUpLinkOnlyTimeout() uint32 {
	if x != nil {
		return x.UpLinkOnlyTimeout
	}
	return 0
}

func (x *PolicyConfig) GetDownLinkOnlyTimeout() uint32 {
	if x != nil {
		return x.DownLinkOnlyTimeout
	}
	return 0
}

func (x *PolicyConfig) GetUdpIdleTimeout() uint32 {
	if x != nil {
		return x.UdpIdleTimeout
	}
	return 0
}

func (x *PolicyConfig) GetLinkStats() bool {
	if x != nil {
		return x.LinkStats
	}
	return false
}

func (x *PolicyConfig) GetInboundStats() bool {
	if x != nil {
		return x.InboundStats
	}
	return false
}

func (x *PolicyConfig) GetUserStats() bool {
	if x != nil {
		return x.UserStats
	}
	return false
}

func (x *PolicyConfig) GetOutboundStats() bool {
	if x != nil {
		return x.OutboundStats
	}
	return false
}

func (x *PolicyConfig) GetSessionStats() bool {
	if x != nil {
		return x.SessionStats
	}
	return false
}

func (x *PolicyConfig) GetDefaultBufferSize() int32 {
	if x != nil {
		return x.DefaultBufferSize
	}
	return 0
}

type UserPolicy struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	BufferSize    int32                  `protobuf:"varint,1,opt,name=buffer_size,json=bufferSize,proto3" json:"buffer_size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UserPolicy) Reset() {
	*x = UserPolicy{}
	mi := &file_protos_policy_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserPolicy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserPolicy) ProtoMessage() {}

func (x *UserPolicy) ProtoReflect() protoreflect.Message {
	mi := &file_protos_policy_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserPolicy.ProtoReflect.Descriptor instead.
func (*UserPolicy) Descriptor() ([]byte, []int) {
	return file_protos_policy_proto_rawDescGZIP(), []int{1}
}

func (x *UserPolicy) GetBufferSize() int32 {
	if x != nil {
		return x.BufferSize
	}
	return 0
}

var File_protos_policy_proto protoreflect.FileDescriptor

const file_protos_policy_proto_rawDesc = "" +
	"\n" +
	"\x13protos/policy.proto\x12\x01x\"\xfd\x04\n" +
	"\fPolicyConfig\x12J\n" +
	"\x0fuser_policy_map\x18\v \x03(\v2\".x.PolicyConfig.UserPolicyMapEntryR\ruserPolicyMap\x12+\n" +
	"\x11handshake_timeout\x18\x01 \x01(\rR\x10handshakeTimeout\x126\n" +
	"\x17connection_idle_timeout\x18\x02 \x01(\rR\x15connectionIdleTimeout\x12.\n" +
	"\x13upLink_only_timeout\x18\x04 \x01(\rR\x11upLinkOnlyTimeout\x122\n" +
	"\x15downLink_only_timeout\x18\x05 \x01(\rR\x13downLinkOnlyTimeout\x12(\n" +
	"\x10udp_idle_timeout\x18\x03 \x01(\rR\x0eudpIdleTimeout\x12\x1d\n" +
	"\n" +
	"link_stats\x18\a \x01(\bR\tlinkStats\x12#\n" +
	"\rinbound_stats\x18\b \x01(\bR\finboundStats\x12\x1d\n" +
	"\n" +
	"user_stats\x18\t \x01(\bR\tuserStats\x12%\n" +
	"\x0eoutbound_stats\x18\n" +
	" \x01(\bR\routboundStats\x12#\n" +
	"\rsession_stats\x18\r \x01(\bR\fsessionStats\x12.\n" +
	"\x13default_buffer_size\x18\f \x01(\x05R\x11defaultBufferSize\x1aO\n" +
	"\x12UserPolicyMapEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\rR\x03key\x12#\n" +
	"\x05value\x18\x02 \x01(\v2\r.x.UserPolicyR\x05value:\x028\x01\"-\n" +
	"\n" +
	"UserPolicy\x12\x1f\n" +
	"\vbuffer_size\x18\x01 \x01(\x05R\n" +
	"bufferSizeB*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_policy_proto_rawDescOnce sync.Once
	file_protos_policy_proto_rawDescData []byte
)

func file_protos_policy_proto_rawDescGZIP() []byte {
	file_protos_policy_proto_rawDescOnce.Do(func() {
		file_protos_policy_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_policy_proto_rawDesc), len(file_protos_policy_proto_rawDesc)))
	})
	return file_protos_policy_proto_rawDescData
}

var file_protos_policy_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_protos_policy_proto_goTypes = []any{
	(*PolicyConfig)(nil), // 0: x.PolicyConfig
	(*UserPolicy)(nil),   // 1: x.UserPolicy
	nil,                  // 2: x.PolicyConfig.UserPolicyMapEntry
}
var file_protos_policy_proto_depIdxs = []int32{
	2, // 0: x.PolicyConfig.user_policy_map:type_name -> x.PolicyConfig.UserPolicyMapEntry
	1, // 1: x.PolicyConfig.UserPolicyMapEntry.value:type_name -> x.UserPolicy
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_protos_policy_proto_init() }
func file_protos_policy_proto_init() {
	if File_protos_policy_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_policy_proto_rawDesc), len(file_protos_policy_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_policy_proto_goTypes,
		DependencyIndexes: file_protos_policy_proto_depIdxs,
		MessageInfos:      file_protos_policy_proto_msgTypes,
	}.Build()
	File_protos_policy_proto = out.File
	file_protos_policy_proto_goTypes = nil
	file_protos_policy_proto_depIdxs = nil
}
