package proxy

import (
	configs "github.com/5vnetwork/vx-core/app/configs"
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

type Obfs int32

const (
	Obfs_Salamander Obfs = 0
)

// Enum value maps for Obfs.
var (
	Obfs_name = map[int32]string{
		0: "Salamander",
	}
	Obfs_value = map[string]int32{
		"Salamander": 0,
	}
)

func (x Obfs) Enum() *Obfs {
	p := new(Obfs)
	*p = x
	return p
}

func (x Obfs) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Obfs) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_proxy_hysteria_proto_enumTypes[0].Descriptor()
}

func (Obfs) Type() protoreflect.EnumType {
	return &file_protos_proxy_hysteria_proto_enumTypes[0]
}

func (x Obfs) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Obfs.Descriptor instead.
func (Obfs) EnumDescriptor() ([]byte, []int) {
	return file_protos_proxy_hysteria_proto_rawDescGZIP(), []int{0}
}

type QuicConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// MB
	InitialStreamReceiveWindow     uint32 `protobuf:"varint,5,opt,name=initial_stream_receive_window,json=initialStreamReceiveWindow,proto3" json:"initial_stream_receive_window,omitempty"`
	MaxStreamReceiveWindow         uint32 `protobuf:"varint,6,opt,name=max_stream_receive_window,json=maxStreamReceiveWindow,proto3" json:"max_stream_receive_window,omitempty"`
	InitialConnectionReceiveWindow uint32 `protobuf:"varint,7,opt,name=initial_connection_receive_window,json=initialConnectionReceiveWindow,proto3" json:"initial_connection_receive_window,omitempty"`
	MaxConnectionReceiveWindow     uint32 `protobuf:"varint,8,opt,name=max_connection_receive_window,json=maxConnectionReceiveWindow,proto3" json:"max_connection_receive_window,omitempty"`
	// bytes
	InitialStreamReceiveWindowBytes     uint64 `protobuf:"varint,13,opt,name=initial_stream_receive_window_bytes,json=initialStreamReceiveWindowBytes,proto3" json:"initial_stream_receive_window_bytes,omitempty"`
	MaxStreamReceiveWindowBytes         uint64 `protobuf:"varint,14,opt,name=max_stream_receive_window_bytes,json=maxStreamReceiveWindowBytes,proto3" json:"max_stream_receive_window_bytes,omitempty"`
	InitialConnectionReceiveWindowBytes uint64 `protobuf:"varint,15,opt,name=initial_connection_receive_window_bytes,json=initialConnectionReceiveWindowBytes,proto3" json:"initial_connection_receive_window_bytes,omitempty"`
	MaxConnectionReceiveWindowBytes     uint64 `protobuf:"varint,16,opt,name=max_connection_receive_window_bytes,json=maxConnectionReceiveWindowBytes,proto3" json:"max_connection_receive_window_bytes,omitempty"`
	MaxIdleTimeout                      uint32 `protobuf:"varint,9,opt,name=max_idle_timeout,json=maxIdleTimeout,proto3" json:"max_idle_timeout,omitempty"`
	KeepAlivePeriod                     uint32 `protobuf:"varint,10,opt,name=keep_alive_period,json=keepAlivePeriod,proto3" json:"keep_alive_period,omitempty"`
	DisablePathMtuDiscovery             bool   `protobuf:"varint,11,opt,name=disable_path_mtu_discovery,json=disablePathMtuDiscovery,proto3" json:"disable_path_mtu_discovery,omitempty"`
	// server only
	MaxIncomingStreams uint32 `protobuf:"varint,12,opt,name=max_incoming_streams,json=maxIncomingStreams,proto3" json:"max_incoming_streams,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *QuicConfig) Reset() {
	*x = QuicConfig{}
	mi := &file_protos_proxy_hysteria_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QuicConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QuicConfig) ProtoMessage() {}

func (x *QuicConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_hysteria_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QuicConfig.ProtoReflect.Descriptor instead.
func (*QuicConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_hysteria_proto_rawDescGZIP(), []int{0}
}

func (x *QuicConfig) GetInitialStreamReceiveWindow() uint32 {
	if x != nil {
		return x.InitialStreamReceiveWindow
	}
	return 0
}

func (x *QuicConfig) GetMaxStreamReceiveWindow() uint32 {
	if x != nil {
		return x.MaxStreamReceiveWindow
	}
	return 0
}

func (x *QuicConfig) GetInitialConnectionReceiveWindow() uint32 {
	if x != nil {
		return x.InitialConnectionReceiveWindow
	}
	return 0
}

func (x *QuicConfig) GetMaxConnectionReceiveWindow() uint32 {
	if x != nil {
		return x.MaxConnectionReceiveWindow
	}
	return 0
}

func (x *QuicConfig) GetInitialStreamReceiveWindowBytes() uint64 {
	if x != nil {
		return x.InitialStreamReceiveWindowBytes
	}
	return 0
}

func (x *QuicConfig) GetMaxStreamReceiveWindowBytes() uint64 {
	if x != nil {
		return x.MaxStreamReceiveWindowBytes
	}
	return 0
}

func (x *QuicConfig) GetInitialConnectionReceiveWindowBytes() uint64 {
	if x != nil {
		return x.InitialConnectionReceiveWindowBytes
	}
	return 0
}

func (x *QuicConfig) GetMaxConnectionReceiveWindowBytes() uint64 {
	if x != nil {
		return x.MaxConnectionReceiveWindowBytes
	}
	return 0
}

func (x *QuicConfig) GetMaxIdleTimeout() uint32 {
	if x != nil {
		return x.MaxIdleTimeout
	}
	return 0
}

func (x *QuicConfig) GetKeepAlivePeriod() uint32 {
	if x != nil {
		return x.KeepAlivePeriod
	}
	return 0
}

func (x *QuicConfig) GetDisablePathMtuDiscovery() bool {
	if x != nil {
		return x.DisablePathMtuDiscovery
	}
	return false
}

func (x *QuicConfig) GetMaxIncomingStreams() uint32 {
	if x != nil {
		return x.MaxIncomingStreams
	}
	return 0
}

type BandwidthConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// for client, this is upload
	MaxTx uint32 `protobuf:"varint,1,opt,name=max_tx,json=maxTx,proto3" json:"max_tx,omitempty"`
	// for client, this is download
	MaxRx         uint32 `protobuf:"varint,2,opt,name=max_rx,json=maxRx,proto3" json:"max_rx,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BandwidthConfig) Reset() {
	*x = BandwidthConfig{}
	mi := &file_protos_proxy_hysteria_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BandwidthConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BandwidthConfig) ProtoMessage() {}

func (x *BandwidthConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_hysteria_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BandwidthConfig.ProtoReflect.Descriptor instead.
func (*BandwidthConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_hysteria_proto_rawDescGZIP(), []int{1}
}

func (x *BandwidthConfig) GetMaxTx() uint32 {
	if x != nil {
		return x.MaxTx
	}
	return 0
}

func (x *BandwidthConfig) GetMaxRx() uint32 {
	if x != nil {
		return x.MaxRx
	}
	return 0
}

type Hysteria2ClientConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Auth          string                 `protobuf:"bytes,3,opt,name=auth,proto3" json:"auth,omitempty"`
	TlsConfig     *tls.TlsConfig         `protobuf:"bytes,4,opt,name=tls_config,json=tlsConfig,proto3" json:"tls_config,omitempty"`
	Quic          *QuicConfig            `protobuf:"bytes,12,opt,name=quic,proto3" json:"quic,omitempty"`
	FastOpen      bool                   `protobuf:"varint,13,opt,name=fast_open,json=fastOpen,proto3" json:"fast_open,omitempty"`
	Bandwidth     *BandwidthConfig       `protobuf:"bytes,14,opt,name=bandwidth,proto3" json:"bandwidth,omitempty"`
	Obfs          *ObfsConfig            `protobuf:"bytes,15,opt,name=obfs,proto3" json:"obfs,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Hysteria2ClientConfig) Reset() {
	*x = Hysteria2ClientConfig{}
	mi := &file_protos_proxy_hysteria_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Hysteria2ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Hysteria2ClientConfig) ProtoMessage() {}

func (x *Hysteria2ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_hysteria_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Hysteria2ClientConfig.ProtoReflect.Descriptor instead.
func (*Hysteria2ClientConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_hysteria_proto_rawDescGZIP(), []int{2}
}

func (x *Hysteria2ClientConfig) GetAuth() string {
	if x != nil {
		return x.Auth
	}
	return ""
}

func (x *Hysteria2ClientConfig) GetTlsConfig() *tls.TlsConfig {
	if x != nil {
		return x.TlsConfig
	}
	return nil
}

func (x *Hysteria2ClientConfig) GetQuic() *QuicConfig {
	if x != nil {
		return x.Quic
	}
	return nil
}

func (x *Hysteria2ClientConfig) GetFastOpen() bool {
	if x != nil {
		return x.FastOpen
	}
	return false
}

func (x *Hysteria2ClientConfig) GetBandwidth() *BandwidthConfig {
	if x != nil {
		return x.Bandwidth
	}
	return nil
}

func (x *Hysteria2ClientConfig) GetObfs() *ObfsConfig {
	if x != nil {
		return x.Obfs
	}
	return nil
}

type Hysteria2ServerConfig struct {
	state                 protoimpl.MessageState `protogen:"open.v1"`
	Users                 []*configs.UserConfig  `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
	Obfs                  *ObfsConfig            `protobuf:"bytes,2,opt,name=obfs,proto3" json:"obfs,omitempty"`
	Bandwidth             *BandwidthConfig       `protobuf:"bytes,3,opt,name=bandwidth,proto3" json:"bandwidth,omitempty"`
	Quic                  *QuicConfig            `protobuf:"bytes,4,opt,name=quic,proto3" json:"quic,omitempty"`
	IgnoreClientBandwidth bool                   `protobuf:"varint,7,opt,name=ignore_client_bandwidth,json=ignoreClientBandwidth,proto3" json:"ignore_client_bandwidth,omitempty"`
	TlsConfig             *tls.TlsConfig         `protobuf:"bytes,8,opt,name=tls_config,json=tlsConfig,proto3" json:"tls_config,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *Hysteria2ServerConfig) Reset() {
	*x = Hysteria2ServerConfig{}
	mi := &file_protos_proxy_hysteria_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Hysteria2ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Hysteria2ServerConfig) ProtoMessage() {}

func (x *Hysteria2ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_hysteria_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Hysteria2ServerConfig.ProtoReflect.Descriptor instead.
func (*Hysteria2ServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_hysteria_proto_rawDescGZIP(), []int{3}
}

func (x *Hysteria2ServerConfig) GetUsers() []*configs.UserConfig {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *Hysteria2ServerConfig) GetObfs() *ObfsConfig {
	if x != nil {
		return x.Obfs
	}
	return nil
}

func (x *Hysteria2ServerConfig) GetBandwidth() *BandwidthConfig {
	if x != nil {
		return x.Bandwidth
	}
	return nil
}

func (x *Hysteria2ServerConfig) GetQuic() *QuicConfig {
	if x != nil {
		return x.Quic
	}
	return nil
}

func (x *Hysteria2ServerConfig) GetIgnoreClientBandwidth() bool {
	if x != nil {
		return x.IgnoreClientBandwidth
	}
	return false
}

func (x *Hysteria2ServerConfig) GetTlsConfig() *tls.TlsConfig {
	if x != nil {
		return x.TlsConfig
	}
	return nil
}

type ObfsConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Obfs:
	//
	//	*ObfsConfig_Salamander
	Obfs          isObfsConfig_Obfs `protobuf_oneof:"obfs"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ObfsConfig) Reset() {
	*x = ObfsConfig{}
	mi := &file_protos_proxy_hysteria_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ObfsConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObfsConfig) ProtoMessage() {}

func (x *ObfsConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_hysteria_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObfsConfig.ProtoReflect.Descriptor instead.
func (*ObfsConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_hysteria_proto_rawDescGZIP(), []int{4}
}

func (x *ObfsConfig) GetObfs() isObfsConfig_Obfs {
	if x != nil {
		return x.Obfs
	}
	return nil
}

func (x *ObfsConfig) GetSalamander() *SalamanderConfig {
	if x != nil {
		if x, ok := x.Obfs.(*ObfsConfig_Salamander); ok {
			return x.Salamander
		}
	}
	return nil
}

type isObfsConfig_Obfs interface {
	isObfsConfig_Obfs()
}

type ObfsConfig_Salamander struct {
	Salamander *SalamanderConfig `protobuf:"bytes,1,opt,name=salamander,proto3,oneof"`
}

func (*ObfsConfig_Salamander) isObfsConfig_Obfs() {}

type SalamanderConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Password      string                 `protobuf:"bytes,1,opt,name=password,proto3" json:"password,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SalamanderConfig) Reset() {
	*x = SalamanderConfig{}
	mi := &file_protos_proxy_hysteria_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SalamanderConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SalamanderConfig) ProtoMessage() {}

func (x *SalamanderConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_hysteria_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SalamanderConfig.ProtoReflect.Descriptor instead.
func (*SalamanderConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_hysteria_proto_rawDescGZIP(), []int{5}
}

func (x *SalamanderConfig) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

var File_protos_proxy_hysteria_proto protoreflect.FileDescriptor

const file_protos_proxy_hysteria_proto_rawDesc = "" +
	"\n" +
	"\x1bprotos/proxy/hysteria.proto\x12\ax.proxy\x1a\x11protos/user.proto\x1a\x14protos/tls/tls.proto\"\x95\x06\n" +
	"\n" +
	"QuicConfig\x12A\n" +
	"\x1dinitial_stream_receive_window\x18\x05 \x01(\rR\x1ainitialStreamReceiveWindow\x129\n" +
	"\x19max_stream_receive_window\x18\x06 \x01(\rR\x16maxStreamReceiveWindow\x12I\n" +
	"!initial_connection_receive_window\x18\a \x01(\rR\x1einitialConnectionReceiveWindow\x12A\n" +
	"\x1dmax_connection_receive_window\x18\b \x01(\rR\x1amaxConnectionReceiveWindow\x12L\n" +
	"#initial_stream_receive_window_bytes\x18\r \x01(\x04R\x1finitialStreamReceiveWindowBytes\x12D\n" +
	"\x1fmax_stream_receive_window_bytes\x18\x0e \x01(\x04R\x1bmaxStreamReceiveWindowBytes\x12T\n" +
	"'initial_connection_receive_window_bytes\x18\x0f \x01(\x04R#initialConnectionReceiveWindowBytes\x12L\n" +
	"#max_connection_receive_window_bytes\x18\x10 \x01(\x04R\x1fmaxConnectionReceiveWindowBytes\x12(\n" +
	"\x10max_idle_timeout\x18\t \x01(\rR\x0emaxIdleTimeout\x12*\n" +
	"\x11keep_alive_period\x18\n" +
	" \x01(\rR\x0fkeepAlivePeriod\x12;\n" +
	"\x1adisable_path_mtu_discovery\x18\v \x01(\bR\x17disablePathMtuDiscovery\x120\n" +
	"\x14max_incoming_streams\x18\f \x01(\rR\x12maxIncomingStreams\"?\n" +
	"\x0fBandwidthConfig\x12\x15\n" +
	"\x06max_tx\x18\x01 \x01(\rR\x05maxTx\x12\x15\n" +
	"\x06max_rx\x18\x02 \x01(\rR\x05maxRx\"\x83\x02\n" +
	"\x15Hysteria2ClientConfig\x12\x12\n" +
	"\x04auth\x18\x03 \x01(\tR\x04auth\x12/\n" +
	"\n" +
	"tls_config\x18\x04 \x01(\v2\x10.x.tls.TlsConfigR\ttlsConfig\x12'\n" +
	"\x04quic\x18\f \x01(\v2\x13.x.proxy.QuicConfigR\x04quic\x12\x1b\n" +
	"\tfast_open\x18\r \x01(\bR\bfastOpen\x126\n" +
	"\tbandwidth\x18\x0e \x01(\v2\x18.x.proxy.BandwidthConfigR\tbandwidth\x12'\n" +
	"\x04obfs\x18\x0f \x01(\v2\x13.x.proxy.ObfsConfigR\x04obfs\"\xaf\x02\n" +
	"\x15Hysteria2ServerConfig\x12#\n" +
	"\x05users\x18\x01 \x03(\v2\r.x.UserConfigR\x05users\x12'\n" +
	"\x04obfs\x18\x02 \x01(\v2\x13.x.proxy.ObfsConfigR\x04obfs\x126\n" +
	"\tbandwidth\x18\x03 \x01(\v2\x18.x.proxy.BandwidthConfigR\tbandwidth\x12'\n" +
	"\x04quic\x18\x04 \x01(\v2\x13.x.proxy.QuicConfigR\x04quic\x126\n" +
	"\x17ignore_client_bandwidth\x18\a \x01(\bR\x15ignoreClientBandwidth\x12/\n" +
	"\n" +
	"tls_config\x18\b \x01(\v2\x10.x.tls.TlsConfigR\ttlsConfig\"Q\n" +
	"\n" +
	"ObfsConfig\x12;\n" +
	"\n" +
	"salamander\x18\x01 \x01(\v2\x19.x.proxy.SalamanderConfigH\x00R\n" +
	"salamanderB\x06\n" +
	"\x04obfs\".\n" +
	"\x10SalamanderConfig\x12\x1a\n" +
	"\bpassword\x18\x01 \x01(\tR\bpassword*\x16\n" +
	"\x04Obfs\x12\x0e\n" +
	"\n" +
	"Salamander\x10\x00B0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_hysteria_proto_rawDescOnce sync.Once
	file_protos_proxy_hysteria_proto_rawDescData []byte
)

func file_protos_proxy_hysteria_proto_rawDescGZIP() []byte {
	file_protos_proxy_hysteria_proto_rawDescOnce.Do(func() {
		file_protos_proxy_hysteria_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_hysteria_proto_rawDesc), len(file_protos_proxy_hysteria_proto_rawDesc)))
	})
	return file_protos_proxy_hysteria_proto_rawDescData
}

var file_protos_proxy_hysteria_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protos_proxy_hysteria_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_protos_proxy_hysteria_proto_goTypes = []any{
	(Obfs)(0),                     // 0: x.proxy.Obfs
	(*QuicConfig)(nil),            // 1: x.proxy.QuicConfig
	(*BandwidthConfig)(nil),       // 2: x.proxy.BandwidthConfig
	(*Hysteria2ClientConfig)(nil), // 3: x.proxy.Hysteria2ClientConfig
	(*Hysteria2ServerConfig)(nil), // 4: x.proxy.Hysteria2ServerConfig
	(*ObfsConfig)(nil),            // 5: x.proxy.ObfsConfig
	(*SalamanderConfig)(nil),      // 6: x.proxy.SalamanderConfig
	(*tls.TlsConfig)(nil),         // 7: x.tls.TlsConfig
	(*configs.UserConfig)(nil),    // 8: x.UserConfig
}
var file_protos_proxy_hysteria_proto_depIdxs = []int32{
	7,  // 0: x.proxy.Hysteria2ClientConfig.tls_config:type_name -> x.tls.TlsConfig
	1,  // 1: x.proxy.Hysteria2ClientConfig.quic:type_name -> x.proxy.QuicConfig
	2,  // 2: x.proxy.Hysteria2ClientConfig.bandwidth:type_name -> x.proxy.BandwidthConfig
	5,  // 3: x.proxy.Hysteria2ClientConfig.obfs:type_name -> x.proxy.ObfsConfig
	8,  // 4: x.proxy.Hysteria2ServerConfig.users:type_name -> x.UserConfig
	5,  // 5: x.proxy.Hysteria2ServerConfig.obfs:type_name -> x.proxy.ObfsConfig
	2,  // 6: x.proxy.Hysteria2ServerConfig.bandwidth:type_name -> x.proxy.BandwidthConfig
	1,  // 7: x.proxy.Hysteria2ServerConfig.quic:type_name -> x.proxy.QuicConfig
	7,  // 8: x.proxy.Hysteria2ServerConfig.tls_config:type_name -> x.tls.TlsConfig
	6,  // 9: x.proxy.ObfsConfig.salamander:type_name -> x.proxy.SalamanderConfig
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_protos_proxy_hysteria_proto_init() }
func file_protos_proxy_hysteria_proto_init() {
	if File_protos_proxy_hysteria_proto != nil {
		return
	}
	file_protos_proxy_hysteria_proto_msgTypes[4].OneofWrappers = []any{
		(*ObfsConfig_Salamander)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_hysteria_proto_rawDesc), len(file_protos_proxy_hysteria_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_hysteria_proto_goTypes,
		DependencyIndexes: file_protos_proxy_hysteria_proto_depIdxs,
		EnumInfos:         file_protos_proxy_hysteria_proto_enumTypes,
		MessageInfos:      file_protos_proxy_hysteria_proto_msgTypes,
	}.Build()
	File_protos_proxy_hysteria_proto = out.File
	file_protos_proxy_hysteria_proto_goTypes = nil
	file_protos_proxy_hysteria_proto_depIdxs = nil
}
