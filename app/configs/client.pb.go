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

type TmConfig struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	InboundManager *InboundManagerConfig  `protobuf:"bytes,1,opt,name=inbound_manager,json=inboundManager,proto3" json:"inbound_manager,omitempty"`
	Dns            *DnsConfig             `protobuf:"bytes,3,opt,name=dns,proto3" json:"dns,omitempty"`
	Policy         *PolicyConfig          `protobuf:"bytes,4,opt,name=policy,proto3" json:"policy,omitempty"`
	Selectors      *SelectorsConfig       `protobuf:"bytes,5,opt,name=selectors,proto3" json:"selectors,omitempty"`
	Router         *RouterConfig          `protobuf:"bytes,6,opt,name=router,proto3" json:"router,omitempty"`
	Log            *LoggerConfig          `protobuf:"bytes,7,opt,name=log,proto3" json:"log,omitempty"`
	Dispatcher     *DispatcherConfig      `protobuf:"bytes,8,opt,name=dispatcher,proto3" json:"dispatcher,omitempty"`
	Geo            *GeoConfig             `protobuf:"bytes,13,opt,name=geo,proto3" json:"geo,omitempty"`
	Grpc           *GrpcConfig            `protobuf:"bytes,15,opt,name=grpc,proto3" json:"grpc,omitempty"`
	Tun            *TunConfig             `protobuf:"bytes,17,opt,name=tun,proto3" json:"tun,omitempty"`
	SysProxy       *SysProxyConfig        `protobuf:"bytes,18,opt,name=sys_proxy,json=sysProxy,proto3" json:"sys_proxy,omitempty"`
	DbPath         string                 `protobuf:"bytes,19,opt,name=db_path,json=dbPath,proto3" json:"db_path,omitempty"`
	// used by system extension to communicate with its containing app
	ServiceSecret string `protobuf:"bytes,22,opt,name=service_secret,json=serviceSecret,proto3" json:"service_secret,omitempty"`
	ServicePort   uint32 `protobuf:"varint,23,opt,name=service_port,json=servicePort,proto3" json:"service_port,omitempty"`
	// if true, a component for monitoring default nic will be added
	DefaultNicMonitor bool `protobuf:"varint,20,opt,name=default_nic_monitor,json=defaultNicMonitor,proto3" json:"default_nic_monitor,omitempty"`
	// subscription
	Subscription        *SubscriptionConfig `protobuf:"bytes,21,opt,name=subscription,proto3" json:"subscription,omitempty"`
	Hysteria2RejectQuic bool                `protobuf:"varint,24,opt,name=hysteria2_reject_quic,json=hysteria2RejectQuic,proto3" json:"hysteria2_reject_quic,omitempty"`
	// outbound
	Outbound       *OutboundConfig `protobuf:"bytes,30,opt,name=outbound,proto3" json:"outbound,omitempty"`
	RedirectStdErr string          `protobuf:"bytes,31,opt,name=redirect_std_err,json=redirectStdErr,proto3" json:"redirect_std_err,omitempty"`
	// based on dst, src, network, if there is no handler matched or the
	// matched handler does not support ipv6, reject the request
	RejectIpv6     bool       `protobuf:"varint,32,opt,name=reject_ipv6,json=rejectIpv6,proto3" json:"reject_ipv6,omitempty"`
	Wfp            *WfpConfig `protobuf:"bytes,33,opt,name=wfp,proto3" json:"wfp,omitempty"`
	UseRealLatency bool       `protobuf:"varint,34,opt,name=use_real_latency,json=useRealLatency,proto3" json:"use_real_latency,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *TmConfig) Reset() {
	*x = TmConfig{}
	mi := &file_protos_client_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TmConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TmConfig) ProtoMessage() {}

func (x *TmConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_client_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TmConfig.ProtoReflect.Descriptor instead.
func (*TmConfig) Descriptor() ([]byte, []int) {
	return file_protos_client_proto_rawDescGZIP(), []int{0}
}

func (x *TmConfig) GetInboundManager() *InboundManagerConfig {
	if x != nil {
		return x.InboundManager
	}
	return nil
}

func (x *TmConfig) GetDns() *DnsConfig {
	if x != nil {
		return x.Dns
	}
	return nil
}

func (x *TmConfig) GetPolicy() *PolicyConfig {
	if x != nil {
		return x.Policy
	}
	return nil
}

func (x *TmConfig) GetSelectors() *SelectorsConfig {
	if x != nil {
		return x.Selectors
	}
	return nil
}

func (x *TmConfig) GetRouter() *RouterConfig {
	if x != nil {
		return x.Router
	}
	return nil
}

func (x *TmConfig) GetLog() *LoggerConfig {
	if x != nil {
		return x.Log
	}
	return nil
}

func (x *TmConfig) GetDispatcher() *DispatcherConfig {
	if x != nil {
		return x.Dispatcher
	}
	return nil
}

func (x *TmConfig) GetGeo() *GeoConfig {
	if x != nil {
		return x.Geo
	}
	return nil
}

func (x *TmConfig) GetGrpc() *GrpcConfig {
	if x != nil {
		return x.Grpc
	}
	return nil
}

func (x *TmConfig) GetTun() *TunConfig {
	if x != nil {
		return x.Tun
	}
	return nil
}

func (x *TmConfig) GetSysProxy() *SysProxyConfig {
	if x != nil {
		return x.SysProxy
	}
	return nil
}

func (x *TmConfig) GetDbPath() string {
	if x != nil {
		return x.DbPath
	}
	return ""
}

func (x *TmConfig) GetServiceSecret() string {
	if x != nil {
		return x.ServiceSecret
	}
	return ""
}

func (x *TmConfig) GetServicePort() uint32 {
	if x != nil {
		return x.ServicePort
	}
	return 0
}

func (x *TmConfig) GetDefaultNicMonitor() bool {
	if x != nil {
		return x.DefaultNicMonitor
	}
	return false
}

func (x *TmConfig) GetSubscription() *SubscriptionConfig {
	if x != nil {
		return x.Subscription
	}
	return nil
}

func (x *TmConfig) GetHysteria2RejectQuic() bool {
	if x != nil {
		return x.Hysteria2RejectQuic
	}
	return false
}

func (x *TmConfig) GetOutbound() *OutboundConfig {
	if x != nil {
		return x.Outbound
	}
	return nil
}

func (x *TmConfig) GetRedirectStdErr() string {
	if x != nil {
		return x.RedirectStdErr
	}
	return ""
}

func (x *TmConfig) GetRejectIpv6() bool {
	if x != nil {
		return x.RejectIpv6
	}
	return false
}

func (x *TmConfig) GetWfp() *WfpConfig {
	if x != nil {
		return x.Wfp
	}
	return nil
}

func (x *TmConfig) GetUseRealLatency() bool {
	if x != nil {
		return x.UseRealLatency
	}
	return false
}

type GrpcConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// listen address
	Address       string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port          uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	ClientCert    []byte `protobuf:"bytes,3,opt,name=client_cert,json=clientCert,proto3" json:"client_cert,omitempty"`
	Uid           int32  `protobuf:"varint,4,opt,name=uid,proto3" json:"uid,omitempty"`
	Gid           int32  `protobuf:"varint,5,opt,name=gid,proto3" json:"gid,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GrpcConfig) Reset() {
	*x = GrpcConfig{}
	mi := &file_protos_client_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GrpcConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GrpcConfig) ProtoMessage() {}

func (x *GrpcConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_client_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GrpcConfig.ProtoReflect.Descriptor instead.
func (*GrpcConfig) Descriptor() ([]byte, []int) {
	return file_protos_client_proto_rawDescGZIP(), []int{1}
}

func (x *GrpcConfig) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *GrpcConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *GrpcConfig) GetClientCert() []byte {
	if x != nil {
		return x.ClientCert
	}
	return nil
}

func (x *GrpcConfig) GetUid() int32 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *GrpcConfig) GetGid() int32 {
	if x != nil {
		return x.Gid
	}
	return 0
}

type SubscriptionConfig struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	LastUpdateTime uint32                 `protobuf:"varint,21,opt,name=last_update_time,json=lastUpdateTime,proto3" json:"last_update_time,omitempty"`
	Interval       uint32                 `protobuf:"varint,22,opt,name=interval,proto3" json:"interval,omitempty"`
	PeriodicUpdate bool                   `protobuf:"varint,23,opt,name=periodic_update,json=periodicUpdate,proto3" json:"periodic_update,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *SubscriptionConfig) Reset() {
	*x = SubscriptionConfig{}
	mi := &file_protos_client_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubscriptionConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscriptionConfig) ProtoMessage() {}

func (x *SubscriptionConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_client_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscriptionConfig.ProtoReflect.Descriptor instead.
func (*SubscriptionConfig) Descriptor() ([]byte, []int) {
	return file_protos_client_proto_rawDescGZIP(), []int{2}
}

func (x *SubscriptionConfig) GetLastUpdateTime() uint32 {
	if x != nil {
		return x.LastUpdateTime
	}
	return 0
}

func (x *SubscriptionConfig) GetInterval() uint32 {
	if x != nil {
		return x.Interval
	}
	return 0
}

func (x *SubscriptionConfig) GetPeriodicUpdate() bool {
	if x != nil {
		return x.PeriodicUpdate
	}
	return false
}

var File_protos_client_proto protoreflect.FileDescriptor

const file_protos_client_proto_rawDesc = "" +
	"\n" +
	"\x13protos/client.proto\x12\x01x\x1a\x14protos/inbound.proto\x1a\x15protos/outbound.proto\x1a\x10protos/dns.proto\x1a\x13protos/router.proto\x1a\x13protos/policy.proto\x1a\x13protos/logger.proto\x1a\x10protos/geo.proto\x1a\x17protos/dispatcher.proto\x1a\x10protos/tun.proto\x1a\x15protos/sysproxy.proto\"\xa1\a\n" +
	"\bTmConfig\x12@\n" +
	"\x0finbound_manager\x18\x01 \x01(\v2\x17.x.InboundManagerConfigR\x0einboundManager\x12\x1e\n" +
	"\x03dns\x18\x03 \x01(\v2\f.x.DnsConfigR\x03dns\x12'\n" +
	"\x06policy\x18\x04 \x01(\v2\x0f.x.PolicyConfigR\x06policy\x120\n" +
	"\tselectors\x18\x05 \x01(\v2\x12.x.SelectorsConfigR\tselectors\x12'\n" +
	"\x06router\x18\x06 \x01(\v2\x0f.x.RouterConfigR\x06router\x12!\n" +
	"\x03log\x18\a \x01(\v2\x0f.x.LoggerConfigR\x03log\x123\n" +
	"\n" +
	"dispatcher\x18\b \x01(\v2\x13.x.DispatcherConfigR\n" +
	"dispatcher\x12\x1e\n" +
	"\x03geo\x18\r \x01(\v2\f.x.GeoConfigR\x03geo\x12!\n" +
	"\x04grpc\x18\x0f \x01(\v2\r.x.GrpcConfigR\x04grpc\x12\x1e\n" +
	"\x03tun\x18\x11 \x01(\v2\f.x.TunConfigR\x03tun\x12.\n" +
	"\tsys_proxy\x18\x12 \x01(\v2\x11.x.SysProxyConfigR\bsysProxy\x12\x17\n" +
	"\adb_path\x18\x13 \x01(\tR\x06dbPath\x12%\n" +
	"\x0eservice_secret\x18\x16 \x01(\tR\rserviceSecret\x12!\n" +
	"\fservice_port\x18\x17 \x01(\rR\vservicePort\x12.\n" +
	"\x13default_nic_monitor\x18\x14 \x01(\bR\x11defaultNicMonitor\x129\n" +
	"\fsubscription\x18\x15 \x01(\v2\x15.x.SubscriptionConfigR\fsubscription\x122\n" +
	"\x15hysteria2_reject_quic\x18\x18 \x01(\bR\x13hysteria2RejectQuic\x12-\n" +
	"\boutbound\x18\x1e \x01(\v2\x11.x.OutboundConfigR\boutbound\x12(\n" +
	"\x10redirect_std_err\x18\x1f \x01(\tR\x0eredirectStdErr\x12\x1f\n" +
	"\vreject_ipv6\x18  \x01(\bR\n" +
	"rejectIpv6\x12\x1e\n" +
	"\x03wfp\x18! \x01(\v2\f.x.WfpConfigR\x03wfp\x12(\n" +
	"\x10use_real_latency\x18\" \x01(\bR\x0euseRealLatency\"\x7f\n" +
	"\n" +
	"GrpcConfig\x12\x18\n" +
	"\aaddress\x18\x01 \x01(\tR\aaddress\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\x12\x1f\n" +
	"\vclient_cert\x18\x03 \x01(\fR\n" +
	"clientCert\x12\x10\n" +
	"\x03uid\x18\x04 \x01(\x05R\x03uid\x12\x10\n" +
	"\x03gid\x18\x05 \x01(\x05R\x03gid\"\x83\x01\n" +
	"\x12SubscriptionConfig\x12(\n" +
	"\x10last_update_time\x18\x15 \x01(\rR\x0elastUpdateTime\x12\x1a\n" +
	"\binterval\x18\x16 \x01(\rR\binterval\x12'\n" +
	"\x0fperiodic_update\x18\x17 \x01(\bR\x0eperiodicUpdateB*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_client_proto_rawDescOnce sync.Once
	file_protos_client_proto_rawDescData []byte
)

func file_protos_client_proto_rawDescGZIP() []byte {
	file_protos_client_proto_rawDescOnce.Do(func() {
		file_protos_client_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_client_proto_rawDesc), len(file_protos_client_proto_rawDesc)))
	})
	return file_protos_client_proto_rawDescData
}

var file_protos_client_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_protos_client_proto_goTypes = []any{
	(*TmConfig)(nil),             // 0: x.TmConfig
	(*GrpcConfig)(nil),           // 1: x.GrpcConfig
	(*SubscriptionConfig)(nil),   // 2: x.SubscriptionConfig
	(*InboundManagerConfig)(nil), // 3: x.InboundManagerConfig
	(*DnsConfig)(nil),            // 4: x.DnsConfig
	(*PolicyConfig)(nil),         // 5: x.PolicyConfig
	(*SelectorsConfig)(nil),      // 6: x.SelectorsConfig
	(*RouterConfig)(nil),         // 7: x.RouterConfig
	(*LoggerConfig)(nil),         // 8: x.LoggerConfig
	(*DispatcherConfig)(nil),     // 9: x.DispatcherConfig
	(*GeoConfig)(nil),            // 10: x.GeoConfig
	(*TunConfig)(nil),            // 11: x.TunConfig
	(*SysProxyConfig)(nil),       // 12: x.SysProxyConfig
	(*OutboundConfig)(nil),       // 13: x.OutboundConfig
	(*WfpConfig)(nil),            // 14: x.WfpConfig
}
var file_protos_client_proto_depIdxs = []int32{
	3,  // 0: x.TmConfig.inbound_manager:type_name -> x.InboundManagerConfig
	4,  // 1: x.TmConfig.dns:type_name -> x.DnsConfig
	5,  // 2: x.TmConfig.policy:type_name -> x.PolicyConfig
	6,  // 3: x.TmConfig.selectors:type_name -> x.SelectorsConfig
	7,  // 4: x.TmConfig.router:type_name -> x.RouterConfig
	8,  // 5: x.TmConfig.log:type_name -> x.LoggerConfig
	9,  // 6: x.TmConfig.dispatcher:type_name -> x.DispatcherConfig
	10, // 7: x.TmConfig.geo:type_name -> x.GeoConfig
	1,  // 8: x.TmConfig.grpc:type_name -> x.GrpcConfig
	11, // 9: x.TmConfig.tun:type_name -> x.TunConfig
	12, // 10: x.TmConfig.sys_proxy:type_name -> x.SysProxyConfig
	2,  // 11: x.TmConfig.subscription:type_name -> x.SubscriptionConfig
	13, // 12: x.TmConfig.outbound:type_name -> x.OutboundConfig
	14, // 13: x.TmConfig.wfp:type_name -> x.WfpConfig
	14, // [14:14] is the sub-list for method output_type
	14, // [14:14] is the sub-list for method input_type
	14, // [14:14] is the sub-list for extension type_name
	14, // [14:14] is the sub-list for extension extendee
	0,  // [0:14] is the sub-list for field type_name
}

func init() { file_protos_client_proto_init() }
func file_protos_client_proto_init() {
	if File_protos_client_proto != nil {
		return
	}
	file_protos_inbound_proto_init()
	file_protos_outbound_proto_init()
	file_protos_dns_proto_init()
	file_protos_router_proto_init()
	file_protos_policy_proto_init()
	file_protos_logger_proto_init()
	file_protos_geo_proto_init()
	file_protos_dispatcher_proto_init()
	file_protos_tun_proto_init()
	file_protos_sysproxy_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_client_proto_rawDesc), len(file_protos_client_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_client_proto_goTypes,
		DependencyIndexes: file_protos_client_proto_depIdxs,
		MessageInfos:      file_protos_client_proto_msgTypes,
	}.Build()
	File_protos_client_proto = out.File
	file_protos_client_proto_goTypes = nil
	file_protos_client_proto_depIdxs = nil
}
