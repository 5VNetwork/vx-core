package server

import (
	configs "github.com/5vnetwork/vx-core/app/configs"
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

type ServerConfig struct {
	state         protoimpl.MessageState             `protogen:"open.v1"`
	Id            uint32                             `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Inbounds      []*configs.ProxyInboundConfig      `protobuf:"bytes,3,rep,name=inbounds,proto3" json:"inbounds,omitempty"`
	MultiInbounds []*configs.MultiProxyInboundConfig `protobuf:"bytes,11,rep,name=multi_inbounds,json=multiInbounds,proto3" json:"multi_inbounds,omitempty"`
	Policy        *configs.PolicyConfig              `protobuf:"bytes,4,opt,name=policy,proto3" json:"policy,omitempty"`
	Router        *configs.RouterConfig              `protobuf:"bytes,5,opt,name=router,proto3" json:"router,omitempty"`
	Log           *configs.LoggerConfig              `protobuf:"bytes,6,opt,name=log,proto3" json:"log,omitempty"`
	Users         []*configs.UserConfig              `protobuf:"bytes,7,rep,name=users,proto3" json:"users,omitempty"`
	Outbounds     []*configs.OutboundHandlerConfig   `protobuf:"bytes,9,rep,name=outbounds,proto3" json:"outbounds,omitempty"`
	Geo           *configs.GeoConfig                 `protobuf:"bytes,10,opt,name=geo,proto3" json:"geo,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_protos_server_server_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_server_server_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerConfig.ProtoReflect.Descriptor instead.
func (*ServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_server_server_proto_rawDescGZIP(), []int{0}
}

func (x *ServerConfig) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ServerConfig) GetInbounds() []*configs.ProxyInboundConfig {
	if x != nil {
		return x.Inbounds
	}
	return nil
}

func (x *ServerConfig) GetMultiInbounds() []*configs.MultiProxyInboundConfig {
	if x != nil {
		return x.MultiInbounds
	}
	return nil
}

func (x *ServerConfig) GetPolicy() *configs.PolicyConfig {
	if x != nil {
		return x.Policy
	}
	return nil
}

func (x *ServerConfig) GetRouter() *configs.RouterConfig {
	if x != nil {
		return x.Router
	}
	return nil
}

func (x *ServerConfig) GetLog() *configs.LoggerConfig {
	if x != nil {
		return x.Log
	}
	return nil
}

func (x *ServerConfig) GetUsers() []*configs.UserConfig {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *ServerConfig) GetOutbounds() []*configs.OutboundHandlerConfig {
	if x != nil {
		return x.Outbounds
	}
	return nil
}

func (x *ServerConfig) GetGeo() *configs.GeoConfig {
	if x != nil {
		return x.Geo
	}
	return nil
}

var File_protos_server_server_proto protoreflect.FileDescriptor

const file_protos_server_server_proto_rawDesc = "" +
	"\n" +
	"\x1aprotos/server/server.proto\x12\x01x\x1a\x14protos/inbound.proto\x1a\x15protos/outbound.proto\x1a\x13protos/router.proto\x1a\x13protos/policy.proto\x1a\x13protos/logger.proto\x1a\x10protos/geo.proto\x1a\x11protos/user.proto\"\x86\x03\n" +
	"\fServerConfig\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\rR\x02id\x121\n" +
	"\binbounds\x18\x03 \x03(\v2\x15.x.ProxyInboundConfigR\binbounds\x12A\n" +
	"\x0emulti_inbounds\x18\v \x03(\v2\x1a.x.MultiProxyInboundConfigR\rmultiInbounds\x12'\n" +
	"\x06policy\x18\x04 \x01(\v2\x0f.x.PolicyConfigR\x06policy\x12'\n" +
	"\x06router\x18\x05 \x01(\v2\x0f.x.RouterConfigR\x06router\x12!\n" +
	"\x03log\x18\x06 \x01(\v2\x0f.x.LoggerConfigR\x03log\x12#\n" +
	"\x05users\x18\a \x03(\v2\r.x.UserConfigR\x05users\x126\n" +
	"\toutbounds\x18\t \x03(\v2\x18.x.OutboundHandlerConfigR\toutbounds\x12\x1e\n" +
	"\x03geo\x18\n" +
	" \x01(\v2\f.x.GeoConfigR\x03geoB1Z/github.com/5vnetwork/vx-core/app/configs/serverb\x06proto3"

var (
	file_protos_server_server_proto_rawDescOnce sync.Once
	file_protos_server_server_proto_rawDescData []byte
)

func file_protos_server_server_proto_rawDescGZIP() []byte {
	file_protos_server_server_proto_rawDescOnce.Do(func() {
		file_protos_server_server_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_server_server_proto_rawDesc), len(file_protos_server_server_proto_rawDesc)))
	})
	return file_protos_server_server_proto_rawDescData
}

var file_protos_server_server_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_server_server_proto_goTypes = []any{
	(*ServerConfig)(nil),                    // 0: x.ServerConfig
	(*configs.ProxyInboundConfig)(nil),      // 1: x.ProxyInboundConfig
	(*configs.MultiProxyInboundConfig)(nil), // 2: x.MultiProxyInboundConfig
	(*configs.PolicyConfig)(nil),            // 3: x.PolicyConfig
	(*configs.RouterConfig)(nil),            // 4: x.RouterConfig
	(*configs.LoggerConfig)(nil),            // 5: x.LoggerConfig
	(*configs.UserConfig)(nil),              // 6: x.UserConfig
	(*configs.OutboundHandlerConfig)(nil),   // 7: x.OutboundHandlerConfig
	(*configs.GeoConfig)(nil),               // 8: x.GeoConfig
}
var file_protos_server_server_proto_depIdxs = []int32{
	1, // 0: x.ServerConfig.inbounds:type_name -> x.ProxyInboundConfig
	2, // 1: x.ServerConfig.multi_inbounds:type_name -> x.MultiProxyInboundConfig
	3, // 2: x.ServerConfig.policy:type_name -> x.PolicyConfig
	4, // 3: x.ServerConfig.router:type_name -> x.RouterConfig
	5, // 4: x.ServerConfig.log:type_name -> x.LoggerConfig
	6, // 5: x.ServerConfig.users:type_name -> x.UserConfig
	7, // 6: x.ServerConfig.outbounds:type_name -> x.OutboundHandlerConfig
	8, // 7: x.ServerConfig.geo:type_name -> x.GeoConfig
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_protos_server_server_proto_init() }
func file_protos_server_server_proto_init() {
	if File_protos_server_server_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_server_server_proto_rawDesc), len(file_protos_server_server_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_server_server_proto_goTypes,
		DependencyIndexes: file_protos_server_server_proto_depIdxs,
		MessageInfos:      file_protos_server_server_proto_msgTypes,
	}.Build()
	File_protos_server_server_proto = out.File
	file_protos_server_server_proto_goTypes = nil
	file_protos_server_server_proto_depIdxs = nil
}
