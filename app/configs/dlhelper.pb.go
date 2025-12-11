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

type SocketConfig_TCPFastOpenState int32

const (
	// AsIs is to leave the current TFO state as is, unmodified.
	SocketConfig_AsIs SocketConfig_TCPFastOpenState = 0
	// Enable is for enabling TFO explictly.
	SocketConfig_Enable SocketConfig_TCPFastOpenState = 1
	// Disable is for disabling TFO explictly.
	SocketConfig_Disable SocketConfig_TCPFastOpenState = 2
)

// Enum value maps for SocketConfig_TCPFastOpenState.
var (
	SocketConfig_TCPFastOpenState_name = map[int32]string{
		0: "AsIs",
		1: "Enable",
		2: "Disable",
	}
	SocketConfig_TCPFastOpenState_value = map[string]int32{
		"AsIs":    0,
		"Enable":  1,
		"Disable": 2,
	}
)

func (x SocketConfig_TCPFastOpenState) Enum() *SocketConfig_TCPFastOpenState {
	p := new(SocketConfig_TCPFastOpenState)
	*p = x
	return p
}

func (x SocketConfig_TCPFastOpenState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SocketConfig_TCPFastOpenState) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_dlhelper_proto_enumTypes[0].Descriptor()
}

func (SocketConfig_TCPFastOpenState) Type() protoreflect.EnumType {
	return &file_protos_dlhelper_proto_enumTypes[0]
}

func (x SocketConfig_TCPFastOpenState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SocketConfig_TCPFastOpenState.Descriptor instead.
func (SocketConfig_TCPFastOpenState) EnumDescriptor() ([]byte, []int) {
	return file_protos_dlhelper_proto_rawDescGZIP(), []int{0, 0}
}

type SocketConfig_TProxyMode int32

const (
	// TProxy is off.
	SocketConfig_Off SocketConfig_TProxyMode = 0
	// TProxy mode.
	SocketConfig_TProxy SocketConfig_TProxyMode = 1
	// Redirect mode.
	SocketConfig_Redirect SocketConfig_TProxyMode = 2
)

// Enum value maps for SocketConfig_TProxyMode.
var (
	SocketConfig_TProxyMode_name = map[int32]string{
		0: "Off",
		1: "TProxy",
		2: "Redirect",
	}
	SocketConfig_TProxyMode_value = map[string]int32{
		"Off":      0,
		"TProxy":   1,
		"Redirect": 2,
	}
)

func (x SocketConfig_TProxyMode) Enum() *SocketConfig_TProxyMode {
	p := new(SocketConfig_TProxyMode)
	*p = x
	return p
}

func (x SocketConfig_TProxyMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SocketConfig_TProxyMode) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_dlhelper_proto_enumTypes[1].Descriptor()
}

func (SocketConfig_TProxyMode) Type() protoreflect.EnumType {
	return &file_protos_dlhelper_proto_enumTypes[1]
}

func (x SocketConfig_TProxyMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SocketConfig_TProxyMode.Descriptor instead.
func (SocketConfig_TProxyMode) EnumDescriptor() ([]byte, []int) {
	return file_protos_dlhelper_proto_rawDescGZIP(), []int{0, 1}
}

// SocketConfig is options to be applied on network sockets.
type SocketConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Mark of the connection. If non-zero, the value will be set to SO_MARK.
	Mark uint32 `protobuf:"varint,1,opt,name=mark,proto3" json:"mark,omitempty"`
	// TFO is the state of TFO settings.
	Tfo SocketConfig_TCPFastOpenState `protobuf:"varint,2,opt,name=tfo,proto3,enum=x.SocketConfig_TCPFastOpenState" json:"tfo,omitempty"`
	// TProxy is for enabling TProxy socket option.
	Tproxy SocketConfig_TProxyMode `protobuf:"varint,3,opt,name=tproxy,proto3,enum=x.SocketConfig_TProxyMode" json:"tproxy,omitempty"`
	// ReceiveOriginalDestAddress is for enabling IP_RECVORIGDSTADDR socket
	// option. This option is for UDP only.
	ReceiveOriginalDestAddress bool `protobuf:"varint,4,opt,name=receive_original_dest_address,json=receiveOriginalDestAddress,proto3" json:"receive_original_dest_address,omitempty"`
	// BindAddress is the address to bind to. Determines local address of the
	// socket. Linux only.
	BindAddress          []byte `protobuf:"bytes,5,opt,name=bind_address,json=bindAddress,proto3" json:"bind_address,omitempty"`
	BindPort             uint32 `protobuf:"varint,6,opt,name=bind_port,json=bindPort,proto3" json:"bind_port,omitempty"`
	AcceptProxyProtocol  bool   `protobuf:"varint,7,opt,name=accept_proxy_protocol,json=acceptProxyProtocol,proto3" json:"accept_proxy_protocol,omitempty"`
	TcpKeepAliveInterval int32  `protobuf:"varint,8,opt,name=tcp_keep_alive_interval,json=tcpKeepAliveInterval,proto3" json:"tcp_keep_alive_interval,omitempty"`
	TfoQueueLength       uint32 `protobuf:"varint,9,opt,name=tfo_queue_length,json=tfoQueueLength,proto3" json:"tfo_queue_length,omitempty"`
	TcpKeepAliveIdle     int32  `protobuf:"varint,10,opt,name=tcp_keep_alive_idle,json=tcpKeepAliveIdle,proto3" json:"tcp_keep_alive_idle,omitempty"`
	// Determines the nic device to bind to.
	BindToDevice uint32 `protobuf:"varint,11,opt,name=bind_to_device,json=bindToDevice,proto3" json:"bind_to_device,omitempty"`
	RxBufSize    int64  `protobuf:"varint,12,opt,name=rx_buf_size,json=rxBufSize,proto3" json:"rx_buf_size,omitempty"`
	TxBufSize    int64  `protobuf:"varint,13,opt,name=tx_buf_size,json=txBufSize,proto3" json:"tx_buf_size,omitempty"`
	ForceBufSize bool   `protobuf:"varint,14,opt,name=force_buf_size,json=forceBufSize,proto3" json:"force_buf_size,omitempty"`
	// For dial, local addr is the LocalAddr of the net.Dialer
	// For udp packetConn, local addr is listening address
	// In V2ray, this is the Via property of a outbound handler
	LocalAddr4    string `protobuf:"bytes,16,opt,name=local_addr4,json=localAddr4,proto3" json:"local_addr4,omitempty"`
	LocalAddr6    string `protobuf:"bytes,17,opt,name=local_addr6,json=localAddr6,proto3" json:"local_addr6,omitempty"`
	DialTimeout   uint32 `protobuf:"varint,18,opt,name=dial_timeout,json=dialTimeout,proto3" json:"dial_timeout,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SocketConfig) Reset() {
	*x = SocketConfig{}
	mi := &file_protos_dlhelper_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SocketConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SocketConfig) ProtoMessage() {}

func (x *SocketConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dlhelper_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SocketConfig.ProtoReflect.Descriptor instead.
func (*SocketConfig) Descriptor() ([]byte, []int) {
	return file_protos_dlhelper_proto_rawDescGZIP(), []int{0}
}

func (x *SocketConfig) GetMark() uint32 {
	if x != nil {
		return x.Mark
	}
	return 0
}

func (x *SocketConfig) GetTfo() SocketConfig_TCPFastOpenState {
	if x != nil {
		return x.Tfo
	}
	return SocketConfig_AsIs
}

func (x *SocketConfig) GetTproxy() SocketConfig_TProxyMode {
	if x != nil {
		return x.Tproxy
	}
	return SocketConfig_Off
}

func (x *SocketConfig) GetReceiveOriginalDestAddress() bool {
	if x != nil {
		return x.ReceiveOriginalDestAddress
	}
	return false
}

func (x *SocketConfig) GetBindAddress() []byte {
	if x != nil {
		return x.BindAddress
	}
	return nil
}

func (x *SocketConfig) GetBindPort() uint32 {
	if x != nil {
		return x.BindPort
	}
	return 0
}

func (x *SocketConfig) GetAcceptProxyProtocol() bool {
	if x != nil {
		return x.AcceptProxyProtocol
	}
	return false
}

func (x *SocketConfig) GetTcpKeepAliveInterval() int32 {
	if x != nil {
		return x.TcpKeepAliveInterval
	}
	return 0
}

func (x *SocketConfig) GetTfoQueueLength() uint32 {
	if x != nil {
		return x.TfoQueueLength
	}
	return 0
}

func (x *SocketConfig) GetTcpKeepAliveIdle() int32 {
	if x != nil {
		return x.TcpKeepAliveIdle
	}
	return 0
}

func (x *SocketConfig) GetBindToDevice() uint32 {
	if x != nil {
		return x.BindToDevice
	}
	return 0
}

func (x *SocketConfig) GetRxBufSize() int64 {
	if x != nil {
		return x.RxBufSize
	}
	return 0
}

func (x *SocketConfig) GetTxBufSize() int64 {
	if x != nil {
		return x.TxBufSize
	}
	return 0
}

func (x *SocketConfig) GetForceBufSize() bool {
	if x != nil {
		return x.ForceBufSize
	}
	return false
}

func (x *SocketConfig) GetLocalAddr4() string {
	if x != nil {
		return x.LocalAddr4
	}
	return ""
}

func (x *SocketConfig) GetLocalAddr6() string {
	if x != nil {
		return x.LocalAddr6
	}
	return ""
}

func (x *SocketConfig) GetDialTimeout() uint32 {
	if x != nil {
		return x.DialTimeout
	}
	return 0
}

var File_protos_dlhelper_proto protoreflect.FileDescriptor

const file_protos_dlhelper_proto_rawDesc = "" +
	"\n" +
	"\x15protos/dlhelper.proto\x12\x01x\"\xaa\x06\n" +
	"\fSocketConfig\x12\x12\n" +
	"\x04mark\x18\x01 \x01(\rR\x04mark\x122\n" +
	"\x03tfo\x18\x02 \x01(\x0e2 .x.SocketConfig.TCPFastOpenStateR\x03tfo\x122\n" +
	"\x06tproxy\x18\x03 \x01(\x0e2\x1a.x.SocketConfig.TProxyModeR\x06tproxy\x12A\n" +
	"\x1dreceive_original_dest_address\x18\x04 \x01(\bR\x1areceiveOriginalDestAddress\x12!\n" +
	"\fbind_address\x18\x05 \x01(\fR\vbindAddress\x12\x1b\n" +
	"\tbind_port\x18\x06 \x01(\rR\bbindPort\x122\n" +
	"\x15accept_proxy_protocol\x18\a \x01(\bR\x13acceptProxyProtocol\x125\n" +
	"\x17tcp_keep_alive_interval\x18\b \x01(\x05R\x14tcpKeepAliveInterval\x12(\n" +
	"\x10tfo_queue_length\x18\t \x01(\rR\x0etfoQueueLength\x12-\n" +
	"\x13tcp_keep_alive_idle\x18\n" +
	" \x01(\x05R\x10tcpKeepAliveIdle\x12$\n" +
	"\x0ebind_to_device\x18\v \x01(\rR\fbindToDevice\x12\x1e\n" +
	"\vrx_buf_size\x18\f \x01(\x03R\trxBufSize\x12\x1e\n" +
	"\vtx_buf_size\x18\r \x01(\x03R\ttxBufSize\x12$\n" +
	"\x0eforce_buf_size\x18\x0e \x01(\bR\fforceBufSize\x12\x1f\n" +
	"\vlocal_addr4\x18\x10 \x01(\tR\n" +
	"localAddr4\x12\x1f\n" +
	"\vlocal_addr6\x18\x11 \x01(\tR\n" +
	"localAddr6\x12!\n" +
	"\fdial_timeout\x18\x12 \x01(\rR\vdialTimeout\"5\n" +
	"\x10TCPFastOpenState\x12\b\n" +
	"\x04AsIs\x10\x00\x12\n" +
	"\n" +
	"\x06Enable\x10\x01\x12\v\n" +
	"\aDisable\x10\x02\"/\n" +
	"\n" +
	"TProxyMode\x12\a\n" +
	"\x03Off\x10\x00\x12\n" +
	"\n" +
	"\x06TProxy\x10\x01\x12\f\n" +
	"\bRedirect\x10\x02B*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_dlhelper_proto_rawDescOnce sync.Once
	file_protos_dlhelper_proto_rawDescData []byte
)

func file_protos_dlhelper_proto_rawDescGZIP() []byte {
	file_protos_dlhelper_proto_rawDescOnce.Do(func() {
		file_protos_dlhelper_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_dlhelper_proto_rawDesc), len(file_protos_dlhelper_proto_rawDesc)))
	})
	return file_protos_dlhelper_proto_rawDescData
}

var file_protos_dlhelper_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_protos_dlhelper_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_dlhelper_proto_goTypes = []any{
	(SocketConfig_TCPFastOpenState)(0), // 0: x.SocketConfig.TCPFastOpenState
	(SocketConfig_TProxyMode)(0),       // 1: x.SocketConfig.TProxyMode
	(*SocketConfig)(nil),               // 2: x.SocketConfig
}
var file_protos_dlhelper_proto_depIdxs = []int32{
	0, // 0: x.SocketConfig.tfo:type_name -> x.SocketConfig.TCPFastOpenState
	1, // 1: x.SocketConfig.tproxy:type_name -> x.SocketConfig.TProxyMode
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_protos_dlhelper_proto_init() }
func file_protos_dlhelper_proto_init() {
	if File_protos_dlhelper_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_dlhelper_proto_rawDesc), len(file_protos_dlhelper_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_dlhelper_proto_goTypes,
		DependencyIndexes: file_protos_dlhelper_proto_depIdxs,
		EnumInfos:         file_protos_dlhelper_proto_enumTypes,
		MessageInfos:      file_protos_dlhelper_proto_msgTypes,
	}.Build()
	File_protos_dlhelper_proto = out.File
	file_protos_dlhelper_proto_goTypes = nil
	file_protos_dlhelper_proto_depIdxs = nil
}
