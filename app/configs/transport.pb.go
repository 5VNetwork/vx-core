package configs

import (
	grpc "github.com/5vnetwork/vx-core/transport/protocols/grpc"
	http "github.com/5vnetwork/vx-core/transport/protocols/http"
	httpupgrade "github.com/5vnetwork/vx-core/transport/protocols/httpupgrade"
	kcp "github.com/5vnetwork/vx-core/transport/protocols/kcp"
	quic "github.com/5vnetwork/vx-core/transport/protocols/quic"
	splithttp "github.com/5vnetwork/vx-core/transport/protocols/splithttp"
	tcp "github.com/5vnetwork/vx-core/transport/protocols/tcp"
	websocket "github.com/5vnetwork/vx-core/transport/protocols/websocket"
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

type TransportConfig struct {
	state  protoimpl.MessageState `protogen:"open.v1"`
	Socket *SocketConfig          `protobuf:"bytes,3,opt,name=socket,proto3" json:"socket,omitempty"`
	// Types that are valid to be assigned to Protocol:
	//
	//	*TransportConfig_Tcp
	//	*TransportConfig_Kcp
	//	*TransportConfig_Websocket
	//	*TransportConfig_Http
	//	*TransportConfig_Quic
	//	*TransportConfig_Grpc
	//	*TransportConfig_Httpupgrade
	//	*TransportConfig_Splithttp
	Protocol isTransportConfig_Protocol `protobuf_oneof:"protocol"`
	// Types that are valid to be assigned to Security:
	//
	//	*TransportConfig_Tls
	//	*TransportConfig_Reality
	Security      isTransportConfig_Security `protobuf_oneof:"security"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TransportConfig) Reset() {
	*x = TransportConfig{}
	mi := &file_protos_transport_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransportConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransportConfig) ProtoMessage() {}

func (x *TransportConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_transport_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransportConfig.ProtoReflect.Descriptor instead.
func (*TransportConfig) Descriptor() ([]byte, []int) {
	return file_protos_transport_proto_rawDescGZIP(), []int{0}
}

func (x *TransportConfig) GetSocket() *SocketConfig {
	if x != nil {
		return x.Socket
	}
	return nil
}

func (x *TransportConfig) GetProtocol() isTransportConfig_Protocol {
	if x != nil {
		return x.Protocol
	}
	return nil
}

func (x *TransportConfig) GetTcp() *tcp.TcpConfig {
	if x != nil {
		if x, ok := x.Protocol.(*TransportConfig_Tcp); ok {
			return x.Tcp
		}
	}
	return nil
}

func (x *TransportConfig) GetKcp() *kcp.KcpConfig {
	if x != nil {
		if x, ok := x.Protocol.(*TransportConfig_Kcp); ok {
			return x.Kcp
		}
	}
	return nil
}

func (x *TransportConfig) GetWebsocket() *websocket.WebsocketConfig {
	if x != nil {
		if x, ok := x.Protocol.(*TransportConfig_Websocket); ok {
			return x.Websocket
		}
	}
	return nil
}

func (x *TransportConfig) GetHttp() *http.HttpConfig {
	if x != nil {
		if x, ok := x.Protocol.(*TransportConfig_Http); ok {
			return x.Http
		}
	}
	return nil
}

func (x *TransportConfig) GetQuic() *quic.QuicConfig {
	if x != nil {
		if x, ok := x.Protocol.(*TransportConfig_Quic); ok {
			return x.Quic
		}
	}
	return nil
}

func (x *TransportConfig) GetGrpc() *grpc.GrpcConfig {
	if x != nil {
		if x, ok := x.Protocol.(*TransportConfig_Grpc); ok {
			return x.Grpc
		}
	}
	return nil
}

func (x *TransportConfig) GetHttpupgrade() *httpupgrade.HttpUpgradeConfig {
	if x != nil {
		if x, ok := x.Protocol.(*TransportConfig_Httpupgrade); ok {
			return x.Httpupgrade
		}
	}
	return nil
}

func (x *TransportConfig) GetSplithttp() *splithttp.SplitHttpConfig {
	if x != nil {
		if x, ok := x.Protocol.(*TransportConfig_Splithttp); ok {
			return x.Splithttp
		}
	}
	return nil
}

func (x *TransportConfig) GetSecurity() isTransportConfig_Security {
	if x != nil {
		return x.Security
	}
	return nil
}

func (x *TransportConfig) GetTls() *tls.TlsConfig {
	if x != nil {
		if x, ok := x.Security.(*TransportConfig_Tls); ok {
			return x.Tls
		}
	}
	return nil
}

func (x *TransportConfig) GetReality() *reality.RealityConfig {
	if x != nil {
		if x, ok := x.Security.(*TransportConfig_Reality); ok {
			return x.Reality
		}
	}
	return nil
}

type isTransportConfig_Protocol interface {
	isTransportConfig_Protocol()
}

type TransportConfig_Tcp struct {
	Tcp *tcp.TcpConfig `protobuf:"bytes,5,opt,name=tcp,proto3,oneof"`
}

type TransportConfig_Kcp struct {
	Kcp *kcp.KcpConfig `protobuf:"bytes,6,opt,name=kcp,proto3,oneof"`
}

type TransportConfig_Websocket struct {
	Websocket *websocket.WebsocketConfig `protobuf:"bytes,7,opt,name=websocket,proto3,oneof"`
}

type TransportConfig_Http struct {
	Http *http.HttpConfig `protobuf:"bytes,8,opt,name=http,proto3,oneof"`
}

type TransportConfig_Quic struct {
	Quic *quic.QuicConfig `protobuf:"bytes,9,opt,name=quic,proto3,oneof"`
}

type TransportConfig_Grpc struct {
	Grpc *grpc.GrpcConfig `protobuf:"bytes,10,opt,name=grpc,proto3,oneof"`
}

type TransportConfig_Httpupgrade struct {
	Httpupgrade *httpupgrade.HttpUpgradeConfig `protobuf:"bytes,11,opt,name=httpupgrade,proto3,oneof"`
}

type TransportConfig_Splithttp struct {
	Splithttp *splithttp.SplitHttpConfig `protobuf:"bytes,12,opt,name=splithttp,proto3,oneof"`
}

func (*TransportConfig_Tcp) isTransportConfig_Protocol() {}

func (*TransportConfig_Kcp) isTransportConfig_Protocol() {}

func (*TransportConfig_Websocket) isTransportConfig_Protocol() {}

func (*TransportConfig_Http) isTransportConfig_Protocol() {}

func (*TransportConfig_Quic) isTransportConfig_Protocol() {}

func (*TransportConfig_Grpc) isTransportConfig_Protocol() {}

func (*TransportConfig_Httpupgrade) isTransportConfig_Protocol() {}

func (*TransportConfig_Splithttp) isTransportConfig_Protocol() {}

type isTransportConfig_Security interface {
	isTransportConfig_Security()
}

type TransportConfig_Tls struct {
	Tls *tls.TlsConfig `protobuf:"bytes,20,opt,name=tls,proto3,oneof"`
}

type TransportConfig_Reality struct {
	Reality *reality.RealityConfig `protobuf:"bytes,21,opt,name=reality,proto3,oneof"`
}

func (*TransportConfig_Tls) isTransportConfig_Security() {}

func (*TransportConfig_Reality) isTransportConfig_Security() {}

var File_protos_transport_proto protoreflect.FileDescriptor

const file_protos_transport_proto_rawDesc = "" +
	"\n" +
	"\x16protos/transport.proto\x12\x01x\x1a\x14protos/tls/tls.proto\x1a'transport/security/reality/config.proto\x1a$transport/protocols/tcp/config.proto\x1a$transport/protocols/kcp/config.proto\x1a*transport/protocols/websocket/config.proto\x1a%transport/protocols/http/config.proto\x1a%transport/protocols/quic/config.proto\x1a%transport/protocols/grpc/config.proto\x1a,transport/protocols/httpupgrade/config.proto\x1a*transport/protocols/splithttp/config.proto\x1a\x15protos/dlhelper.proto\"\xed\x05\n" +
	"\x0fTransportConfig\x12'\n" +
	"\x06socket\x18\x03 \x01(\v2\x0f.x.SocketConfigR\x06socket\x128\n" +
	"\x03tcp\x18\x05 \x01(\v2$.x.transport.protocols.tcp.TcpConfigH\x00R\x03tcp\x128\n" +
	"\x03kcp\x18\x06 \x01(\v2$.x.transport.protocols.kcp.KcpConfigH\x00R\x03kcp\x12P\n" +
	"\twebsocket\x18\a \x01(\v20.x.transport.protocols.websocket.WebsocketConfigH\x00R\twebsocket\x12<\n" +
	"\x04http\x18\b \x01(\v2&.x.transport.protocols.http.HttpConfigH\x00R\x04http\x12<\n" +
	"\x04quic\x18\t \x01(\v2&.x.transport.protocols.quic.QuicConfigH\x00R\x04quic\x12<\n" +
	"\x04grpc\x18\n" +
	" \x01(\v2&.x.transport.protocols.grpc.GrpcConfigH\x00R\x04grpc\x12X\n" +
	"\vhttpupgrade\x18\v \x01(\v24.x.transport.protocols.httpupgrade.HttpUpgradeConfigH\x00R\vhttpupgrade\x12P\n" +
	"\tsplithttp\x18\f \x01(\v20.x.transport.protocols.splithttp.SplitHttpConfigH\x00R\tsplithttp\x12$\n" +
	"\x03tls\x18\x14 \x01(\v2\x10.x.tls.TlsConfigH\x01R\x03tls\x12G\n" +
	"\areality\x18\x15 \x01(\v2+.x.transport.security.reality.RealityConfigH\x01R\arealityB\n" +
	"\n" +
	"\bprotocolB\n" +
	"\n" +
	"\bsecurityB*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_transport_proto_rawDescOnce sync.Once
	file_protos_transport_proto_rawDescData []byte
)

func file_protos_transport_proto_rawDescGZIP() []byte {
	file_protos_transport_proto_rawDescOnce.Do(func() {
		file_protos_transport_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_transport_proto_rawDesc), len(file_protos_transport_proto_rawDesc)))
	})
	return file_protos_transport_proto_rawDescData
}

var file_protos_transport_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_transport_proto_goTypes = []any{
	(*TransportConfig)(nil),               // 0: x.TransportConfig
	(*SocketConfig)(nil),                  // 1: x.SocketConfig
	(*tcp.TcpConfig)(nil),                 // 2: x.transport.protocols.tcp.TcpConfig
	(*kcp.KcpConfig)(nil),                 // 3: x.transport.protocols.kcp.KcpConfig
	(*websocket.WebsocketConfig)(nil),     // 4: x.transport.protocols.websocket.WebsocketConfig
	(*http.HttpConfig)(nil),               // 5: x.transport.protocols.http.HttpConfig
	(*quic.QuicConfig)(nil),               // 6: x.transport.protocols.quic.QuicConfig
	(*grpc.GrpcConfig)(nil),               // 7: x.transport.protocols.grpc.GrpcConfig
	(*httpupgrade.HttpUpgradeConfig)(nil), // 8: x.transport.protocols.httpupgrade.HttpUpgradeConfig
	(*splithttp.SplitHttpConfig)(nil),     // 9: x.transport.protocols.splithttp.SplitHttpConfig
	(*tls.TlsConfig)(nil),                 // 10: x.tls.TlsConfig
	(*reality.RealityConfig)(nil),         // 11: x.transport.security.reality.RealityConfig
}
var file_protos_transport_proto_depIdxs = []int32{
	1,  // 0: x.TransportConfig.socket:type_name -> x.SocketConfig
	2,  // 1: x.TransportConfig.tcp:type_name -> x.transport.protocols.tcp.TcpConfig
	3,  // 2: x.TransportConfig.kcp:type_name -> x.transport.protocols.kcp.KcpConfig
	4,  // 3: x.TransportConfig.websocket:type_name -> x.transport.protocols.websocket.WebsocketConfig
	5,  // 4: x.TransportConfig.http:type_name -> x.transport.protocols.http.HttpConfig
	6,  // 5: x.TransportConfig.quic:type_name -> x.transport.protocols.quic.QuicConfig
	7,  // 6: x.TransportConfig.grpc:type_name -> x.transport.protocols.grpc.GrpcConfig
	8,  // 7: x.TransportConfig.httpupgrade:type_name -> x.transport.protocols.httpupgrade.HttpUpgradeConfig
	9,  // 8: x.TransportConfig.splithttp:type_name -> x.transport.protocols.splithttp.SplitHttpConfig
	10, // 9: x.TransportConfig.tls:type_name -> x.tls.TlsConfig
	11, // 10: x.TransportConfig.reality:type_name -> x.transport.security.reality.RealityConfig
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_protos_transport_proto_init() }
func file_protos_transport_proto_init() {
	if File_protos_transport_proto != nil {
		return
	}
	file_protos_dlhelper_proto_init()
	file_protos_transport_proto_msgTypes[0].OneofWrappers = []any{
		(*TransportConfig_Tcp)(nil),
		(*TransportConfig_Kcp)(nil),
		(*TransportConfig_Websocket)(nil),
		(*TransportConfig_Http)(nil),
		(*TransportConfig_Quic)(nil),
		(*TransportConfig_Grpc)(nil),
		(*TransportConfig_Httpupgrade)(nil),
		(*TransportConfig_Splithttp)(nil),
		(*TransportConfig_Tls)(nil),
		(*TransportConfig_Reality)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_transport_proto_rawDesc), len(file_protos_transport_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_transport_proto_goTypes,
		DependencyIndexes: file_protos_transport_proto_depIdxs,
		MessageInfos:      file_protos_transport_proto_msgTypes,
	}.Build()
	File_protos_transport_proto = out.File
	file_protos_transport_proto_goTypes = nil
	file_protos_transport_proto_depIdxs = nil
}
