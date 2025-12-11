package configs

import (
	grpc "github.com/5vnetwork/vx-core/transport/protocols/grpc"
	http "github.com/5vnetwork/vx-core/transport/protocols/http"
	httpupgrade "github.com/5vnetwork/vx-core/transport/protocols/httpupgrade"
	splithttp "github.com/5vnetwork/vx-core/transport/protocols/splithttp"
	tcp "github.com/5vnetwork/vx-core/transport/protocols/tcp"
	websocket "github.com/5vnetwork/vx-core/transport/protocols/websocket"
	reality "github.com/5vnetwork/vx-core/transport/security/reality"
	tls "github.com/5vnetwork/vx-core/transport/security/tls"
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

type ProxyInboundConfig struct {
	state   protoimpl.MessageState `protogen:"open.v1"`
	Address string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Tag     string                 `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	// if both port and ports are null, 5 random ports will be used
	Port          uint32           `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	Ports         []uint32         `protobuf:"varint,4,rep,packed,name=ports,proto3" json:"ports,omitempty"`
	Transport     *TransportConfig `protobuf:"bytes,6,opt,name=transport,proto3" json:"transport,omitempty"`
	Protocol      *anypb.Any       `protobuf:"bytes,7,opt,name=protocol,proto3" json:"protocol,omitempty"`
	Protocols     []*anypb.Any     `protobuf:"bytes,8,rep,name=protocols,proto3" json:"protocols,omitempty"`
	Users         []*UserConfig    `protobuf:"bytes,9,rep,name=users,proto3" json:"users,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProxyInboundConfig) Reset() {
	*x = ProxyInboundConfig{}
	mi := &file_protos_inbound_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProxyInboundConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProxyInboundConfig) ProtoMessage() {}

func (x *ProxyInboundConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_inbound_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProxyInboundConfig.ProtoReflect.Descriptor instead.
func (*ProxyInboundConfig) Descriptor() ([]byte, []int) {
	return file_protos_inbound_proto_rawDescGZIP(), []int{0}
}

func (x *ProxyInboundConfig) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ProxyInboundConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *ProxyInboundConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ProxyInboundConfig) GetPorts() []uint32 {
	if x != nil {
		return x.Ports
	}
	return nil
}

func (x *ProxyInboundConfig) GetTransport() *TransportConfig {
	if x != nil {
		return x.Transport
	}
	return nil
}

func (x *ProxyInboundConfig) GetProtocol() *anypb.Any {
	if x != nil {
		return x.Protocol
	}
	return nil
}

func (x *ProxyInboundConfig) GetProtocols() []*anypb.Any {
	if x != nil {
		return x.Protocols
	}
	return nil
}

func (x *ProxyInboundConfig) GetUsers() []*UserConfig {
	if x != nil {
		return x.Users
	}
	return nil
}

type InboundManagerConfig struct {
	state         protoimpl.MessageState     `protogen:"open.v1"`
	Handlers      []*ProxyInboundConfig      `protobuf:"bytes,1,rep,name=handlers,proto3" json:"handlers,omitempty"`
	MultiInbounds []*MultiProxyInboundConfig `protobuf:"bytes,2,rep,name=multi_inbounds,json=multiInbounds,proto3" json:"multi_inbounds,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InboundManagerConfig) Reset() {
	*x = InboundManagerConfig{}
	mi := &file_protos_inbound_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InboundManagerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InboundManagerConfig) ProtoMessage() {}

func (x *InboundManagerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_inbound_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InboundManagerConfig.ProtoReflect.Descriptor instead.
func (*InboundManagerConfig) Descriptor() ([]byte, []int) {
	return file_protos_inbound_proto_rawDescGZIP(), []int{1}
}

func (x *InboundManagerConfig) GetHandlers() []*ProxyInboundConfig {
	if x != nil {
		return x.Handlers
	}
	return nil
}

func (x *InboundManagerConfig) GetMultiInbounds() []*MultiProxyInboundConfig {
	if x != nil {
		return x.MultiInbounds
	}
	return nil
}

type MultiProxyInboundConfig struct {
	state              protoimpl.MessageState              `protogen:"open.v1"`
	Address            string                              `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Tag                string                              `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	Ports              []uint32                            `protobuf:"varint,3,rep,packed,name=ports,proto3" json:"ports,omitempty"`
	Protocols          []*anypb.Any                        `protobuf:"bytes,4,rep,name=protocols,proto3" json:"protocols,omitempty"`
	SecurityConfigs    []*MultiProxyInboundConfig_Security `protobuf:"bytes,5,rep,name=security_configs,json=securityConfigs,proto3" json:"security_configs,omitempty"`
	TransportProtocols []*MultiProxyInboundConfig_Protocol `protobuf:"bytes,6,rep,name=transport_protocols,json=transportProtocols,proto3" json:"transport_protocols,omitempty"`
	Socket             *SocketConfig                       `protobuf:"bytes,8,opt,name=socket,proto3" json:"socket,omitempty"`
	Users              []*UserConfig                       `protobuf:"bytes,9,rep,name=users,proto3" json:"users,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *MultiProxyInboundConfig) Reset() {
	*x = MultiProxyInboundConfig{}
	mi := &file_protos_inbound_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MultiProxyInboundConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MultiProxyInboundConfig) ProtoMessage() {}

func (x *MultiProxyInboundConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_inbound_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MultiProxyInboundConfig.ProtoReflect.Descriptor instead.
func (*MultiProxyInboundConfig) Descriptor() ([]byte, []int) {
	return file_protos_inbound_proto_rawDescGZIP(), []int{2}
}

func (x *MultiProxyInboundConfig) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *MultiProxyInboundConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *MultiProxyInboundConfig) GetPorts() []uint32 {
	if x != nil {
		return x.Ports
	}
	return nil
}

func (x *MultiProxyInboundConfig) GetProtocols() []*anypb.Any {
	if x != nil {
		return x.Protocols
	}
	return nil
}

func (x *MultiProxyInboundConfig) GetSecurityConfigs() []*MultiProxyInboundConfig_Security {
	if x != nil {
		return x.SecurityConfigs
	}
	return nil
}

func (x *MultiProxyInboundConfig) GetTransportProtocols() []*MultiProxyInboundConfig_Protocol {
	if x != nil {
		return x.TransportProtocols
	}
	return nil
}

func (x *MultiProxyInboundConfig) GetSocket() *SocketConfig {
	if x != nil {
		return x.Socket
	}
	return nil
}

func (x *MultiProxyInboundConfig) GetUsers() []*UserConfig {
	if x != nil {
		return x.Users
	}
	return nil
}

type WfpConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TcpPort       uint32                 `protobuf:"varint,1,opt,name=tcp_port,json=tcpPort,proto3" json:"tcp_port,omitempty"`
	UdpPort       uint32                 `protobuf:"varint,2,opt,name=udp_port,json=udpPort,proto3" json:"udp_port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *WfpConfig) Reset() {
	*x = WfpConfig{}
	mi := &file_protos_inbound_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WfpConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WfpConfig) ProtoMessage() {}

func (x *WfpConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_inbound_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WfpConfig.ProtoReflect.Descriptor instead.
func (*WfpConfig) Descriptor() ([]byte, []int) {
	return file_protos_inbound_proto_rawDescGZIP(), []int{3}
}

func (x *WfpConfig) GetTcpPort() uint32 {
	if x != nil {
		return x.TcpPort
	}
	return 0
}

func (x *WfpConfig) GetUdpPort() uint32 {
	if x != nil {
		return x.UdpPort
	}
	return 0
}

type MultiProxyInboundConfig_Security struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Security:
	//
	//	*MultiProxyInboundConfig_Security_Tls
	//	*MultiProxyInboundConfig_Security_Reality
	Security          isMultiProxyInboundConfig_Security_Security `protobuf_oneof:"security"`
	Domains           []string                                    `protobuf:"bytes,1,rep,name=domains,proto3" json:"domains,omitempty"`
	RegularExpression string                                      `protobuf:"bytes,2,opt,name=regular_expression,json=regularExpression,proto3" json:"regular_expression,omitempty"`
	Always            bool                                        `protobuf:"varint,3,opt,name=always,proto3" json:"always,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *MultiProxyInboundConfig_Security) Reset() {
	*x = MultiProxyInboundConfig_Security{}
	mi := &file_protos_inbound_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MultiProxyInboundConfig_Security) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MultiProxyInboundConfig_Security) ProtoMessage() {}

func (x *MultiProxyInboundConfig_Security) ProtoReflect() protoreflect.Message {
	mi := &file_protos_inbound_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MultiProxyInboundConfig_Security.ProtoReflect.Descriptor instead.
func (*MultiProxyInboundConfig_Security) Descriptor() ([]byte, []int) {
	return file_protos_inbound_proto_rawDescGZIP(), []int{2, 0}
}

func (x *MultiProxyInboundConfig_Security) GetSecurity() isMultiProxyInboundConfig_Security_Security {
	if x != nil {
		return x.Security
	}
	return nil
}

func (x *MultiProxyInboundConfig_Security) GetTls() *tls.TlsConfig {
	if x != nil {
		if x, ok := x.Security.(*MultiProxyInboundConfig_Security_Tls); ok {
			return x.Tls
		}
	}
	return nil
}

func (x *MultiProxyInboundConfig_Security) GetReality() *reality.RealityConfig {
	if x != nil {
		if x, ok := x.Security.(*MultiProxyInboundConfig_Security_Reality); ok {
			return x.Reality
		}
	}
	return nil
}

func (x *MultiProxyInboundConfig_Security) GetDomains() []string {
	if x != nil {
		return x.Domains
	}
	return nil
}

func (x *MultiProxyInboundConfig_Security) GetRegularExpression() string {
	if x != nil {
		return x.RegularExpression
	}
	return ""
}

func (x *MultiProxyInboundConfig_Security) GetAlways() bool {
	if x != nil {
		return x.Always
	}
	return false
}

type isMultiProxyInboundConfig_Security_Security interface {
	isMultiProxyInboundConfig_Security_Security()
}

type MultiProxyInboundConfig_Security_Tls struct {
	Tls *tls.TlsConfig `protobuf:"bytes,20,opt,name=tls,proto3,oneof"`
}

type MultiProxyInboundConfig_Security_Reality struct {
	Reality *reality.RealityConfig `protobuf:"bytes,21,opt,name=reality,proto3,oneof"`
}

func (*MultiProxyInboundConfig_Security_Tls) isMultiProxyInboundConfig_Security_Security() {}

func (*MultiProxyInboundConfig_Security_Reality) isMultiProxyInboundConfig_Security_Security() {}

type MultiProxyInboundConfig_Protocol struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Protocol:
	//
	//	*MultiProxyInboundConfig_Protocol_Websocket
	//	*MultiProxyInboundConfig_Protocol_Http
	//	*MultiProxyInboundConfig_Protocol_Grpc
	//	*MultiProxyInboundConfig_Protocol_Httpupgrade
	//	*MultiProxyInboundConfig_Protocol_Splithttp
	//	*MultiProxyInboundConfig_Protocol_Tcp
	Protocol      isMultiProxyInboundConfig_Protocol_Protocol `protobuf_oneof:"protocol"`
	Alpn          string                                      `protobuf:"bytes,1,opt,name=alpn,proto3" json:"alpn,omitempty"`
	Sni           string                                      `protobuf:"bytes,2,opt,name=sni,proto3" json:"sni,omitempty"`
	Path          string                                      `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	H2            bool                                        `protobuf:"varint,4,opt,name=h2,proto3" json:"h2,omitempty"`
	Always        bool                                        `protobuf:"varint,5,opt,name=always,proto3" json:"always,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MultiProxyInboundConfig_Protocol) Reset() {
	*x = MultiProxyInboundConfig_Protocol{}
	mi := &file_protos_inbound_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MultiProxyInboundConfig_Protocol) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MultiProxyInboundConfig_Protocol) ProtoMessage() {}

func (x *MultiProxyInboundConfig_Protocol) ProtoReflect() protoreflect.Message {
	mi := &file_protos_inbound_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MultiProxyInboundConfig_Protocol.ProtoReflect.Descriptor instead.
func (*MultiProxyInboundConfig_Protocol) Descriptor() ([]byte, []int) {
	return file_protos_inbound_proto_rawDescGZIP(), []int{2, 1}
}

func (x *MultiProxyInboundConfig_Protocol) GetProtocol() isMultiProxyInboundConfig_Protocol_Protocol {
	if x != nil {
		return x.Protocol
	}
	return nil
}

func (x *MultiProxyInboundConfig_Protocol) GetWebsocket() *websocket.WebsocketConfig {
	if x != nil {
		if x, ok := x.Protocol.(*MultiProxyInboundConfig_Protocol_Websocket); ok {
			return x.Websocket
		}
	}
	return nil
}

func (x *MultiProxyInboundConfig_Protocol) GetHttp() *http.HttpConfig {
	if x != nil {
		if x, ok := x.Protocol.(*MultiProxyInboundConfig_Protocol_Http); ok {
			return x.Http
		}
	}
	return nil
}

func (x *MultiProxyInboundConfig_Protocol) GetGrpc() *grpc.GrpcConfig {
	if x != nil {
		if x, ok := x.Protocol.(*MultiProxyInboundConfig_Protocol_Grpc); ok {
			return x.Grpc
		}
	}
	return nil
}

func (x *MultiProxyInboundConfig_Protocol) GetHttpupgrade() *httpupgrade.HttpUpgradeConfig {
	if x != nil {
		if x, ok := x.Protocol.(*MultiProxyInboundConfig_Protocol_Httpupgrade); ok {
			return x.Httpupgrade
		}
	}
	return nil
}

func (x *MultiProxyInboundConfig_Protocol) GetSplithttp() *splithttp.SplitHttpConfig {
	if x != nil {
		if x, ok := x.Protocol.(*MultiProxyInboundConfig_Protocol_Splithttp); ok {
			return x.Splithttp
		}
	}
	return nil
}

func (x *MultiProxyInboundConfig_Protocol) GetTcp() *tcp.TcpConfig {
	if x != nil {
		if x, ok := x.Protocol.(*MultiProxyInboundConfig_Protocol_Tcp); ok {
			return x.Tcp
		}
	}
	return nil
}

func (x *MultiProxyInboundConfig_Protocol) GetAlpn() string {
	if x != nil {
		return x.Alpn
	}
	return ""
}

func (x *MultiProxyInboundConfig_Protocol) GetSni() string {
	if x != nil {
		return x.Sni
	}
	return ""
}

func (x *MultiProxyInboundConfig_Protocol) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *MultiProxyInboundConfig_Protocol) GetH2() bool {
	if x != nil {
		return x.H2
	}
	return false
}

func (x *MultiProxyInboundConfig_Protocol) GetAlways() bool {
	if x != nil {
		return x.Always
	}
	return false
}

type isMultiProxyInboundConfig_Protocol_Protocol interface {
	isMultiProxyInboundConfig_Protocol_Protocol()
}

type MultiProxyInboundConfig_Protocol_Websocket struct {
	Websocket *websocket.WebsocketConfig `protobuf:"bytes,7,opt,name=websocket,proto3,oneof"`
}

type MultiProxyInboundConfig_Protocol_Http struct {
	Http *http.HttpConfig `protobuf:"bytes,8,opt,name=http,proto3,oneof"`
}

type MultiProxyInboundConfig_Protocol_Grpc struct {
	Grpc *grpc.GrpcConfig `protobuf:"bytes,10,opt,name=grpc,proto3,oneof"`
}

type MultiProxyInboundConfig_Protocol_Httpupgrade struct {
	Httpupgrade *httpupgrade.HttpUpgradeConfig `protobuf:"bytes,11,opt,name=httpupgrade,proto3,oneof"`
}

type MultiProxyInboundConfig_Protocol_Splithttp struct {
	Splithttp *splithttp.SplitHttpConfig `protobuf:"bytes,12,opt,name=splithttp,proto3,oneof"`
}

type MultiProxyInboundConfig_Protocol_Tcp struct {
	Tcp *tcp.TcpConfig `protobuf:"bytes,13,opt,name=tcp,proto3,oneof"`
}

func (*MultiProxyInboundConfig_Protocol_Websocket) isMultiProxyInboundConfig_Protocol_Protocol() {}

func (*MultiProxyInboundConfig_Protocol_Http) isMultiProxyInboundConfig_Protocol_Protocol() {}

func (*MultiProxyInboundConfig_Protocol_Grpc) isMultiProxyInboundConfig_Protocol_Protocol() {}

func (*MultiProxyInboundConfig_Protocol_Httpupgrade) isMultiProxyInboundConfig_Protocol_Protocol() {}

func (*MultiProxyInboundConfig_Protocol_Splithttp) isMultiProxyInboundConfig_Protocol_Protocol() {}

func (*MultiProxyInboundConfig_Protocol_Tcp) isMultiProxyInboundConfig_Protocol_Protocol() {}

var File_protos_inbound_proto protoreflect.FileDescriptor

const file_protos_inbound_proto_rawDesc = "" +
	"\n" +
	"\x14protos/inbound.proto\x12\x01x\x1a\x16protos/transport.proto\x1a\x19google/protobuf/any.proto\x1a\x14protos/tls/tls.proto\x1a'transport/security/reality/config.proto\x1a*transport/protocols/websocket/config.proto\x1a%transport/protocols/http/config.proto\x1a%transport/protocols/grpc/config.proto\x1a,transport/protocols/httpupgrade/config.proto\x1a*transport/protocols/splithttp/config.proto\x1a$transport/protocols/tcp/config.proto\x1a\x15protos/dlhelper.proto\x1a\x11protos/user.proto\"\xa7\x02\n" +
	"\x12ProxyInboundConfig\x12\x18\n" +
	"\aaddress\x18\x01 \x01(\tR\aaddress\x12\x10\n" +
	"\x03tag\x18\x02 \x01(\tR\x03tag\x12\x12\n" +
	"\x04port\x18\x03 \x01(\rR\x04port\x12\x14\n" +
	"\x05ports\x18\x04 \x03(\rR\x05ports\x120\n" +
	"\ttransport\x18\x06 \x01(\v2\x12.x.TransportConfigR\ttransport\x120\n" +
	"\bprotocol\x18\a \x01(\v2\x14.google.protobuf.AnyR\bprotocol\x122\n" +
	"\tprotocols\x18\b \x03(\v2\x14.google.protobuf.AnyR\tprotocols\x12#\n" +
	"\x05users\x18\t \x03(\v2\r.x.UserConfigR\x05users\"\x8c\x01\n" +
	"\x14InboundManagerConfig\x121\n" +
	"\bhandlers\x18\x01 \x03(\v2\x15.x.ProxyInboundConfigR\bhandlers\x12A\n" +
	"\x0emulti_inbounds\x18\x02 \x03(\v2\x1a.x.MultiProxyInboundConfigR\rmultiInbounds\"\x9b\t\n" +
	"\x17MultiProxyInboundConfig\x12\x18\n" +
	"\aaddress\x18\x01 \x01(\tR\aaddress\x12\x10\n" +
	"\x03tag\x18\x02 \x01(\tR\x03tag\x12\x14\n" +
	"\x05ports\x18\x03 \x03(\rR\x05ports\x122\n" +
	"\tprotocols\x18\x04 \x03(\v2\x14.google.protobuf.AnyR\tprotocols\x12N\n" +
	"\x10security_configs\x18\x05 \x03(\v2#.x.MultiProxyInboundConfig.SecurityR\x0fsecurityConfigs\x12T\n" +
	"\x13transport_protocols\x18\x06 \x03(\v2#.x.MultiProxyInboundConfig.ProtocolR\x12transportProtocols\x12'\n" +
	"\x06socket\x18\b \x01(\v2\x0f.x.SocketConfigR\x06socket\x12#\n" +
	"\x05users\x18\t \x03(\v2\r.x.UserConfigR\x05users\x1a\xe6\x01\n" +
	"\bSecurity\x12$\n" +
	"\x03tls\x18\x14 \x01(\v2\x10.x.tls.TlsConfigH\x00R\x03tls\x12G\n" +
	"\areality\x18\x15 \x01(\v2+.x.transport.security.reality.RealityConfigH\x00R\areality\x12\x18\n" +
	"\adomains\x18\x01 \x03(\tR\adomains\x12-\n" +
	"\x12regular_expression\x18\x02 \x01(\tR\x11regularExpression\x12\x16\n" +
	"\x06always\x18\x03 \x01(\bR\x06alwaysB\n" +
	"\n" +
	"\bsecurity\x1a\xac\x04\n" +
	"\bProtocol\x12P\n" +
	"\twebsocket\x18\a \x01(\v20.x.transport.protocols.websocket.WebsocketConfigH\x00R\twebsocket\x12<\n" +
	"\x04http\x18\b \x01(\v2&.x.transport.protocols.http.HttpConfigH\x00R\x04http\x12<\n" +
	"\x04grpc\x18\n" +
	" \x01(\v2&.x.transport.protocols.grpc.GrpcConfigH\x00R\x04grpc\x12X\n" +
	"\vhttpupgrade\x18\v \x01(\v24.x.transport.protocols.httpupgrade.HttpUpgradeConfigH\x00R\vhttpupgrade\x12P\n" +
	"\tsplithttp\x18\f \x01(\v20.x.transport.protocols.splithttp.SplitHttpConfigH\x00R\tsplithttp\x128\n" +
	"\x03tcp\x18\r \x01(\v2$.x.transport.protocols.tcp.TcpConfigH\x00R\x03tcp\x12\x12\n" +
	"\x04alpn\x18\x01 \x01(\tR\x04alpn\x12\x10\n" +
	"\x03sni\x18\x02 \x01(\tR\x03sni\x12\x12\n" +
	"\x04path\x18\x03 \x01(\tR\x04path\x12\x0e\n" +
	"\x02h2\x18\x04 \x01(\bR\x02h2\x12\x16\n" +
	"\x06always\x18\x05 \x01(\bR\x06alwaysB\n" +
	"\n" +
	"\bprotocol\"A\n" +
	"\tWfpConfig\x12\x19\n" +
	"\btcp_port\x18\x01 \x01(\rR\atcpPort\x12\x19\n" +
	"\budp_port\x18\x02 \x01(\rR\audpPortB*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_inbound_proto_rawDescOnce sync.Once
	file_protos_inbound_proto_rawDescData []byte
)

func file_protos_inbound_proto_rawDescGZIP() []byte {
	file_protos_inbound_proto_rawDescOnce.Do(func() {
		file_protos_inbound_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_inbound_proto_rawDesc), len(file_protos_inbound_proto_rawDesc)))
	})
	return file_protos_inbound_proto_rawDescData
}

var file_protos_inbound_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_protos_inbound_proto_goTypes = []any{
	(*ProxyInboundConfig)(nil),               // 0: x.ProxyInboundConfig
	(*InboundManagerConfig)(nil),             // 1: x.InboundManagerConfig
	(*MultiProxyInboundConfig)(nil),          // 2: x.MultiProxyInboundConfig
	(*WfpConfig)(nil),                        // 3: x.WfpConfig
	(*MultiProxyInboundConfig_Security)(nil), // 4: x.MultiProxyInboundConfig.Security
	(*MultiProxyInboundConfig_Protocol)(nil), // 5: x.MultiProxyInboundConfig.Protocol
	(*TransportConfig)(nil),                  // 6: x.TransportConfig
	(*anypb.Any)(nil),                        // 7: google.protobuf.Any
	(*UserConfig)(nil),                       // 8: x.UserConfig
	(*SocketConfig)(nil),                     // 9: x.SocketConfig
	(*tls.TlsConfig)(nil),                    // 10: x.tls.TlsConfig
	(*reality.RealityConfig)(nil),            // 11: x.transport.security.reality.RealityConfig
	(*websocket.WebsocketConfig)(nil),        // 12: x.transport.protocols.websocket.WebsocketConfig
	(*http.HttpConfig)(nil),                  // 13: x.transport.protocols.http.HttpConfig
	(*grpc.GrpcConfig)(nil),                  // 14: x.transport.protocols.grpc.GrpcConfig
	(*httpupgrade.HttpUpgradeConfig)(nil),    // 15: x.transport.protocols.httpupgrade.HttpUpgradeConfig
	(*splithttp.SplitHttpConfig)(nil),        // 16: x.transport.protocols.splithttp.SplitHttpConfig
	(*tcp.TcpConfig)(nil),                    // 17: x.transport.protocols.tcp.TcpConfig
}
var file_protos_inbound_proto_depIdxs = []int32{
	6,  // 0: x.ProxyInboundConfig.transport:type_name -> x.TransportConfig
	7,  // 1: x.ProxyInboundConfig.protocol:type_name -> google.protobuf.Any
	7,  // 2: x.ProxyInboundConfig.protocols:type_name -> google.protobuf.Any
	8,  // 3: x.ProxyInboundConfig.users:type_name -> x.UserConfig
	0,  // 4: x.InboundManagerConfig.handlers:type_name -> x.ProxyInboundConfig
	2,  // 5: x.InboundManagerConfig.multi_inbounds:type_name -> x.MultiProxyInboundConfig
	7,  // 6: x.MultiProxyInboundConfig.protocols:type_name -> google.protobuf.Any
	4,  // 7: x.MultiProxyInboundConfig.security_configs:type_name -> x.MultiProxyInboundConfig.Security
	5,  // 8: x.MultiProxyInboundConfig.transport_protocols:type_name -> x.MultiProxyInboundConfig.Protocol
	9,  // 9: x.MultiProxyInboundConfig.socket:type_name -> x.SocketConfig
	8,  // 10: x.MultiProxyInboundConfig.users:type_name -> x.UserConfig
	10, // 11: x.MultiProxyInboundConfig.Security.tls:type_name -> x.tls.TlsConfig
	11, // 12: x.MultiProxyInboundConfig.Security.reality:type_name -> x.transport.security.reality.RealityConfig
	12, // 13: x.MultiProxyInboundConfig.Protocol.websocket:type_name -> x.transport.protocols.websocket.WebsocketConfig
	13, // 14: x.MultiProxyInboundConfig.Protocol.http:type_name -> x.transport.protocols.http.HttpConfig
	14, // 15: x.MultiProxyInboundConfig.Protocol.grpc:type_name -> x.transport.protocols.grpc.GrpcConfig
	15, // 16: x.MultiProxyInboundConfig.Protocol.httpupgrade:type_name -> x.transport.protocols.httpupgrade.HttpUpgradeConfig
	16, // 17: x.MultiProxyInboundConfig.Protocol.splithttp:type_name -> x.transport.protocols.splithttp.SplitHttpConfig
	17, // 18: x.MultiProxyInboundConfig.Protocol.tcp:type_name -> x.transport.protocols.tcp.TcpConfig
	19, // [19:19] is the sub-list for method output_type
	19, // [19:19] is the sub-list for method input_type
	19, // [19:19] is the sub-list for extension type_name
	19, // [19:19] is the sub-list for extension extendee
	0,  // [0:19] is the sub-list for field type_name
}

func init() { file_protos_inbound_proto_init() }
func file_protos_inbound_proto_init() {
	if File_protos_inbound_proto != nil {
		return
	}
	file_protos_transport_proto_init()
	file_protos_dlhelper_proto_init()
	file_protos_user_proto_init()
	file_protos_inbound_proto_msgTypes[4].OneofWrappers = []any{
		(*MultiProxyInboundConfig_Security_Tls)(nil),
		(*MultiProxyInboundConfig_Security_Reality)(nil),
	}
	file_protos_inbound_proto_msgTypes[5].OneofWrappers = []any{
		(*MultiProxyInboundConfig_Protocol_Websocket)(nil),
		(*MultiProxyInboundConfig_Protocol_Http)(nil),
		(*MultiProxyInboundConfig_Protocol_Grpc)(nil),
		(*MultiProxyInboundConfig_Protocol_Httpupgrade)(nil),
		(*MultiProxyInboundConfig_Protocol_Splithttp)(nil),
		(*MultiProxyInboundConfig_Protocol_Tcp)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_inbound_proto_rawDesc), len(file_protos_inbound_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_inbound_proto_goTypes,
		DependencyIndexes: file_protos_inbound_proto_depIdxs,
		MessageInfos:      file_protos_inbound_proto_msgTypes,
	}.Build()
	File_protos_inbound_proto = out.File
	file_protos_inbound_proto_goTypes = nil
	file_protos_inbound_proto_depIdxs = nil
}
