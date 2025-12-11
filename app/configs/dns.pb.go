package configs

import (
	geo "github.com/5vnetwork/vx-core/common/geo"
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

type DnsType int32

const (
	DnsType_DnsType_A    DnsType = 0
	DnsType_DnsType_AAAA DnsType = 1
)

// Enum value maps for DnsType.
var (
	DnsType_name = map[int32]string{
		0: "DnsType_A",
		1: "DnsType_AAAA",
	}
	DnsType_value = map[string]int32{
		"DnsType_A":    0,
		"DnsType_AAAA": 1,
	}
)

func (x DnsType) Enum() *DnsType {
	p := new(DnsType)
	*p = x
	return p
}

func (x DnsType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DnsType) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_dns_proto_enumTypes[0].Descriptor()
}

func (DnsType) Type() protoreflect.EnumType {
	return &file_protos_dns_proto_enumTypes[0]
}

func (x DnsType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DnsType.Descriptor instead.
func (DnsType) EnumDescriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{0}
}

type DnsConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Records       []*Record              `protobuf:"bytes,1,rep,name=records,proto3" json:"records,omitempty"`
	DnsServers    []*DnsServerConfig     `protobuf:"bytes,3,rep,name=dns_servers,json=dnsServers,proto3" json:"dns_servers,omitempty"`
	DnsRules      []*DnsRuleConfig       `protobuf:"bytes,4,rep,name=dns_rules,json=dnsRules,proto3" json:"dns_rules,omitempty"`
	EnableFakeDns bool                   `protobuf:"varint,5,opt,name=enable_fake_dns,json=enableFakeDns,proto3" json:"enable_fake_dns,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DnsConfig) Reset() {
	*x = DnsConfig{}
	mi := &file_protos_dns_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DnsConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DnsConfig) ProtoMessage() {}

func (x *DnsConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DnsConfig.ProtoReflect.Descriptor instead.
func (*DnsConfig) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{0}
}

func (x *DnsConfig) GetRecords() []*Record {
	if x != nil {
		return x.Records
	}
	return nil
}

func (x *DnsConfig) GetDnsServers() []*DnsServerConfig {
	if x != nil {
		return x.DnsServers
	}
	return nil
}

func (x *DnsConfig) GetDnsRules() []*DnsRuleConfig {
	if x != nil {
		return x.DnsRules
	}
	return nil
}

func (x *DnsConfig) GetEnableFakeDns() bool {
	if x != nil {
		return x.EnableFakeDns
	}
	return false
}

type DnsRules struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Rules         []*DnsRuleConfig       `protobuf:"bytes,1,rep,name=rules,proto3" json:"rules,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DnsRules) Reset() {
	*x = DnsRules{}
	mi := &file_protos_dns_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DnsRules) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DnsRules) ProtoMessage() {}

func (x *DnsRules) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DnsRules.ProtoReflect.Descriptor instead.
func (*DnsRules) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{1}
}

func (x *DnsRules) GetRules() []*DnsRuleConfig {
	if x != nil {
		return x.Rules
	}
	return nil
}

type DnsRuleConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	DnsServerName string                 `protobuf:"bytes,1,opt,name=dns_server_name,json=dnsServerName,proto3" json:"dns_server_name,omitempty"`
	// used to construct preferred domains
	Domains []*geo.Domain `protobuf:"bytes,10,rep,name=domains,proto3" json:"domains,omitempty"`
	// used to construct preferred domains
	DomainTags []string `protobuf:"bytes,11,rep,name=domain_tags,json=domainTags,proto3" json:"domain_tags,omitempty"`
	// types
	IncludedTypes []DnsType `protobuf:"varint,12,rep,packed,name=included_types,json=includedTypes,proto3,enum=x.DnsType" json:"included_types,omitempty"`
	// for debug and display purpose
	RuleName      string `protobuf:"bytes,20,opt,name=rule_name,json=ruleName,proto3" json:"rule_name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DnsRuleConfig) Reset() {
	*x = DnsRuleConfig{}
	mi := &file_protos_dns_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DnsRuleConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DnsRuleConfig) ProtoMessage() {}

func (x *DnsRuleConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DnsRuleConfig.ProtoReflect.Descriptor instead.
func (*DnsRuleConfig) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{2}
}

func (x *DnsRuleConfig) GetDnsServerName() string {
	if x != nil {
		return x.DnsServerName
	}
	return ""
}

func (x *DnsRuleConfig) GetDomains() []*geo.Domain {
	if x != nil {
		return x.Domains
	}
	return nil
}

func (x *DnsRuleConfig) GetDomainTags() []string {
	if x != nil {
		return x.DomainTags
	}
	return nil
}

func (x *DnsRuleConfig) GetIncludedTypes() []DnsType {
	if x != nil {
		return x.IncludedTypes
	}
	return nil
}

func (x *DnsRuleConfig) GetRuleName() string {
	if x != nil {
		return x.RuleName
	}
	return ""
}

type DnsServerConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Type:
	//
	//	*DnsServerConfig_PlainDnsServer
	//	*DnsServerConfig_DohDnsServer
	//	*DnsServerConfig_QuicDnsServer
	//	*DnsServerConfig_FakeDnsServer
	//	*DnsServerConfig_TlsDnsServer
	Type          isDnsServerConfig_Type `protobuf_oneof:"type"`
	Name          string                 `protobuf:"bytes,10,opt,name=name,proto3" json:"name,omitempty"`
	ClientIp      string                 `protobuf:"bytes,11,opt,name=client_ip,json=clientIp,proto3" json:"client_ip,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DnsServerConfig) Reset() {
	*x = DnsServerConfig{}
	mi := &file_protos_dns_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DnsServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DnsServerConfig) ProtoMessage() {}

func (x *DnsServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DnsServerConfig.ProtoReflect.Descriptor instead.
func (*DnsServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{3}
}

func (x *DnsServerConfig) GetType() isDnsServerConfig_Type {
	if x != nil {
		return x.Type
	}
	return nil
}

func (x *DnsServerConfig) GetPlainDnsServer() *PlainDnsServer {
	if x != nil {
		if x, ok := x.Type.(*DnsServerConfig_PlainDnsServer); ok {
			return x.PlainDnsServer
		}
	}
	return nil
}

func (x *DnsServerConfig) GetDohDnsServer() *DohDnsServer {
	if x != nil {
		if x, ok := x.Type.(*DnsServerConfig_DohDnsServer); ok {
			return x.DohDnsServer
		}
	}
	return nil
}

func (x *DnsServerConfig) GetQuicDnsServer() *QuicDnsServer {
	if x != nil {
		if x, ok := x.Type.(*DnsServerConfig_QuicDnsServer); ok {
			return x.QuicDnsServer
		}
	}
	return nil
}

func (x *DnsServerConfig) GetFakeDnsServer() *FakeDnsServer {
	if x != nil {
		if x, ok := x.Type.(*DnsServerConfig_FakeDnsServer); ok {
			return x.FakeDnsServer
		}
	}
	return nil
}

func (x *DnsServerConfig) GetTlsDnsServer() *TlsDnsServer {
	if x != nil {
		if x, ok := x.Type.(*DnsServerConfig_TlsDnsServer); ok {
			return x.TlsDnsServer
		}
	}
	return nil
}

func (x *DnsServerConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *DnsServerConfig) GetClientIp() string {
	if x != nil {
		return x.ClientIp
	}
	return ""
}

type isDnsServerConfig_Type interface {
	isDnsServerConfig_Type()
}

type DnsServerConfig_PlainDnsServer struct {
	PlainDnsServer *PlainDnsServer `protobuf:"bytes,1,opt,name=plain_dns_server,json=plainDnsServer,proto3,oneof"`
}

type DnsServerConfig_DohDnsServer struct {
	DohDnsServer *DohDnsServer `protobuf:"bytes,2,opt,name=doh_dns_server,json=dohDnsServer,proto3,oneof"`
}

type DnsServerConfig_QuicDnsServer struct {
	QuicDnsServer *QuicDnsServer `protobuf:"bytes,3,opt,name=quic_dns_server,json=quicDnsServer,proto3,oneof"`
}

type DnsServerConfig_FakeDnsServer struct {
	FakeDnsServer *FakeDnsServer `protobuf:"bytes,4,opt,name=fake_dns_server,json=fakeDnsServer,proto3,oneof"`
}

type DnsServerConfig_TlsDnsServer struct {
	TlsDnsServer *TlsDnsServer `protobuf:"bytes,5,opt,name=tls_dns_server,json=tlsDnsServer,proto3,oneof"`
}

func (*DnsServerConfig_PlainDnsServer) isDnsServerConfig_Type() {}

func (*DnsServerConfig_DohDnsServer) isDnsServerConfig_Type() {}

func (*DnsServerConfig_QuicDnsServer) isDnsServerConfig_Type() {}

func (*DnsServerConfig_FakeDnsServer) isDnsServerConfig_Type() {}

func (*DnsServerConfig_TlsDnsServer) isDnsServerConfig_Type() {}

type PlainDnsServer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Addresses     []string               `protobuf:"bytes,2,rep,name=addresses,proto3" json:"addresses,omitempty"`
	UseDefaultDns bool                   `protobuf:"varint,3,opt,name=use_default_dns,json=useDefaultDns,proto3" json:"use_default_dns,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PlainDnsServer) Reset() {
	*x = PlainDnsServer{}
	mi := &file_protos_dns_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PlainDnsServer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlainDnsServer) ProtoMessage() {}

func (x *PlainDnsServer) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlainDnsServer.ProtoReflect.Descriptor instead.
func (*PlainDnsServer) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{4}
}

func (x *PlainDnsServer) GetAddresses() []string {
	if x != nil {
		return x.Addresses
	}
	return nil
}

func (x *PlainDnsServer) GetUseDefaultDns() bool {
	if x != nil {
		return x.UseDefaultDns
	}
	return false
}

type TlsDnsServer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Addresses     []string               `protobuf:"bytes,2,rep,name=addresses,proto3" json:"addresses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TlsDnsServer) Reset() {
	*x = TlsDnsServer{}
	mi := &file_protos_dns_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TlsDnsServer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TlsDnsServer) ProtoMessage() {}

func (x *TlsDnsServer) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TlsDnsServer.ProtoReflect.Descriptor instead.
func (*TlsDnsServer) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{5}
}

func (x *TlsDnsServer) GetAddresses() []string {
	if x != nil {
		return x.Addresses
	}
	return nil
}

type DohDnsServer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Url           string                 `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DohDnsServer) Reset() {
	*x = DohDnsServer{}
	mi := &file_protos_dns_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DohDnsServer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DohDnsServer) ProtoMessage() {}

func (x *DohDnsServer) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DohDnsServer.ProtoReflect.Descriptor instead.
func (*DohDnsServer) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{6}
}

func (x *DohDnsServer) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

type QuicDnsServer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       string                 `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	ServerName    string                 `protobuf:"bytes,3,opt,name=serverName,proto3" json:"serverName,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QuicDnsServer) Reset() {
	*x = QuicDnsServer{}
	mi := &file_protos_dns_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QuicDnsServer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QuicDnsServer) ProtoMessage() {}

func (x *QuicDnsServer) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QuicDnsServer.ProtoReflect.Descriptor instead.
func (*QuicDnsServer) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{7}
}

func (x *QuicDnsServer) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *QuicDnsServer) GetServerName() string {
	if x != nil {
		return x.ServerName
	}
	return ""
}

type Record struct {
	state  protoimpl.MessageState `protogen:"open.v1"`
	Domain string                 `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	Ip     []string               `protobuf:"bytes,3,rep,name=ip,proto3" json:"ip,omitempty"`
	// ProxiedDomain indicates the mapped domain has the same IP address on this
	// domain.
	ProxiedDomain string `protobuf:"bytes,4,opt,name=proxied_domain,json=proxiedDomain,proto3" json:"proxied_domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Record) Reset() {
	*x = Record{}
	mi := &file_protos_dns_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Record) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Record) ProtoMessage() {}

func (x *Record) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Record.ProtoReflect.Descriptor instead.
func (*Record) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{8}
}

func (x *Record) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *Record) GetIp() []string {
	if x != nil {
		return x.Ip
	}
	return nil
}

func (x *Record) GetProxiedDomain() string {
	if x != nil {
		return x.ProxiedDomain
	}
	return ""
}

type DnsHijackConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// dns handler name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// dns servers to use for direct dns queries
	DirectAddresses []string `protobuf:"bytes,2,rep,name=direct_addresses,json=directAddresses,proto3" json:"direct_addresses,omitempty"`
	UseDefaultDns   bool     `protobuf:"varint,3,opt,name=use_default_dns,json=useDefaultDns,proto3" json:"use_default_dns,omitempty"`
	ProxyAddresses  []string `protobuf:"bytes,4,rep,name=proxy_addresses,json=proxyAddresses,proto3" json:"proxy_addresses,omitempty"`
	// used as inbound tag for proxy dns server
	DnsConnProxy   string `protobuf:"bytes,5,opt,name=dns_conn_proxy,json=dnsConnProxy,proto3" json:"dns_conn_proxy,omitempty"`
	DnsConnDirect  string `protobuf:"bytes,6,opt,name=dns_conn_direct,json=dnsConnDirect,proto3" json:"dns_conn_direct,omitempty"`
	DnsServerProxy string `protobuf:"bytes,7,opt,name=dns_server_proxy,json=dnsServerProxy,proto3" json:"dns_server_proxy,omitempty"`
	// domain sets that contains all proxy domains
	// all domains in the sets will be resolved by proxy dns server
	ProxyTags                       []string `protobuf:"bytes,8,rep,name=proxy_tags,json=proxyTags,proto3" json:"proxy_tags,omitempty"`
	DirectReturnNilAaaaIfNotSupport bool     `protobuf:"varint,9,opt,name=direct_return_nil_aaaa_if_not_support,json=directReturnNilAaaaIfNotSupport,proto3" json:"direct_return_nil_aaaa_if_not_support,omitempty"`
	ProxyReturnNilAaaaIfNotSupport  bool     `protobuf:"varint,10,opt,name=proxy_return_nil_aaaa_if_not_support,json=proxyReturnNilAaaaIfNotSupport,proto3" json:"proxy_return_nil_aaaa_if_not_support,omitempty"`
	unknownFields                   protoimpl.UnknownFields
	sizeCache                       protoimpl.SizeCache
}

func (x *DnsHijackConfig) Reset() {
	*x = DnsHijackConfig{}
	mi := &file_protos_dns_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DnsHijackConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DnsHijackConfig) ProtoMessage() {}

func (x *DnsHijackConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DnsHijackConfig.ProtoReflect.Descriptor instead.
func (*DnsHijackConfig) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{9}
}

func (x *DnsHijackConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *DnsHijackConfig) GetDirectAddresses() []string {
	if x != nil {
		return x.DirectAddresses
	}
	return nil
}

func (x *DnsHijackConfig) GetUseDefaultDns() bool {
	if x != nil {
		return x.UseDefaultDns
	}
	return false
}

func (x *DnsHijackConfig) GetProxyAddresses() []string {
	if x != nil {
		return x.ProxyAddresses
	}
	return nil
}

func (x *DnsHijackConfig) GetDnsConnProxy() string {
	if x != nil {
		return x.DnsConnProxy
	}
	return ""
}

func (x *DnsHijackConfig) GetDnsConnDirect() string {
	if x != nil {
		return x.DnsConnDirect
	}
	return ""
}

func (x *DnsHijackConfig) GetDnsServerProxy() string {
	if x != nil {
		return x.DnsServerProxy
	}
	return ""
}

func (x *DnsHijackConfig) GetProxyTags() []string {
	if x != nil {
		return x.ProxyTags
	}
	return nil
}

func (x *DnsHijackConfig) GetDirectReturnNilAaaaIfNotSupport() bool {
	if x != nil {
		return x.DirectReturnNilAaaaIfNotSupport
	}
	return false
}

func (x *DnsHijackConfig) GetProxyReturnNilAaaaIfNotSupport() bool {
	if x != nil {
		return x.ProxyReturnNilAaaaIfNotSupport
	}
	return false
}

type FakeDnsServer struct {
	state         protoimpl.MessageState      `protogen:"open.v1"`
	PoolConfigs   []*FakeDnsServer_PoolConfig `protobuf:"bytes,1,rep,name=pool_configs,json=poolConfigs,proto3" json:"pool_configs,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FakeDnsServer) Reset() {
	*x = FakeDnsServer{}
	mi := &file_protos_dns_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FakeDnsServer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FakeDnsServer) ProtoMessage() {}

func (x *FakeDnsServer) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FakeDnsServer.ProtoReflect.Descriptor instead.
func (*FakeDnsServer) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{10}
}

func (x *FakeDnsServer) GetPoolConfigs() []*FakeDnsServer_PoolConfig {
	if x != nil {
		return x.PoolConfigs
	}
	return nil
}

type FakeDnsServer_PoolConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Cidr  string                 `protobuf:"bytes,1,opt,name=cidr,proto3" json:"cidr,omitempty"`
	// FakeDNS 所记忆的「IP - 域名映射」数量。当域名数量超过此数值时，会依据 LRU
	// 规则淘汰老旧域名。
	LruSize       uint32 `protobuf:"varint,2,opt,name=lru_size,json=lruSize,proto3" json:"lru_size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FakeDnsServer_PoolConfig) Reset() {
	*x = FakeDnsServer_PoolConfig{}
	mi := &file_protos_dns_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FakeDnsServer_PoolConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FakeDnsServer_PoolConfig) ProtoMessage() {}

func (x *FakeDnsServer_PoolConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_dns_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FakeDnsServer_PoolConfig.ProtoReflect.Descriptor instead.
func (*FakeDnsServer_PoolConfig) Descriptor() ([]byte, []int) {
	return file_protos_dns_proto_rawDescGZIP(), []int{10, 0}
}

func (x *FakeDnsServer_PoolConfig) GetCidr() string {
	if x != nil {
		return x.Cidr
	}
	return ""
}

func (x *FakeDnsServer_PoolConfig) GetLruSize() uint32 {
	if x != nil {
		return x.LruSize
	}
	return 0
}

var File_protos_dns_proto protoreflect.FileDescriptor

const file_protos_dns_proto_rawDesc = "" +
	"\n" +
	"\x10protos/dns.proto\x12\x01x\x1a\x14common/geo/geo.proto\"\xbc\x01\n" +
	"\tDnsConfig\x12#\n" +
	"\arecords\x18\x01 \x03(\v2\t.x.RecordR\arecords\x123\n" +
	"\vdns_servers\x18\x03 \x03(\v2\x12.x.DnsServerConfigR\n" +
	"dnsServers\x12-\n" +
	"\tdns_rules\x18\x04 \x03(\v2\x10.x.DnsRuleConfigR\bdnsRules\x12&\n" +
	"\x0fenable_fake_dns\x18\x05 \x01(\bR\renableFakeDns\"2\n" +
	"\bDnsRules\x12&\n" +
	"\x05rules\x18\x01 \x03(\v2\x10.x.DnsRuleConfigR\x05rules\"\xd8\x01\n" +
	"\rDnsRuleConfig\x12&\n" +
	"\x0fdns_server_name\x18\x01 \x01(\tR\rdnsServerName\x12.\n" +
	"\adomains\x18\n" +
	" \x03(\v2\x14.x.common.geo.DomainR\adomains\x12\x1f\n" +
	"\vdomain_tags\x18\v \x03(\tR\n" +
	"domainTags\x121\n" +
	"\x0eincluded_types\x18\f \x03(\x0e2\n" +
	".x.DnsTypeR\rincludedTypes\x12\x1b\n" +
	"\trule_name\x18\x14 \x01(\tR\bruleName\"\xf3\x02\n" +
	"\x0fDnsServerConfig\x12=\n" +
	"\x10plain_dns_server\x18\x01 \x01(\v2\x11.x.PlainDnsServerH\x00R\x0eplainDnsServer\x127\n" +
	"\x0edoh_dns_server\x18\x02 \x01(\v2\x0f.x.DohDnsServerH\x00R\fdohDnsServer\x12:\n" +
	"\x0fquic_dns_server\x18\x03 \x01(\v2\x10.x.QuicDnsServerH\x00R\rquicDnsServer\x12:\n" +
	"\x0ffake_dns_server\x18\x04 \x01(\v2\x10.x.FakeDnsServerH\x00R\rfakeDnsServer\x127\n" +
	"\x0etls_dns_server\x18\x05 \x01(\v2\x0f.x.TlsDnsServerH\x00R\ftlsDnsServer\x12\x12\n" +
	"\x04name\x18\n" +
	" \x01(\tR\x04name\x12\x1b\n" +
	"\tclient_ip\x18\v \x01(\tR\bclientIpB\x06\n" +
	"\x04type\"V\n" +
	"\x0ePlainDnsServer\x12\x1c\n" +
	"\taddresses\x18\x02 \x03(\tR\taddresses\x12&\n" +
	"\x0fuse_default_dns\x18\x03 \x01(\bR\ruseDefaultDns\",\n" +
	"\fTlsDnsServer\x12\x1c\n" +
	"\taddresses\x18\x02 \x03(\tR\taddresses\" \n" +
	"\fDohDnsServer\x12\x10\n" +
	"\x03url\x18\x02 \x01(\tR\x03url\"I\n" +
	"\rQuicDnsServer\x12\x18\n" +
	"\aaddress\x18\x02 \x01(\tR\aaddress\x12\x1e\n" +
	"\n" +
	"serverName\x18\x03 \x01(\tR\n" +
	"serverName\"W\n" +
	"\x06Record\x12\x16\n" +
	"\x06domain\x18\x02 \x01(\tR\x06domain\x12\x0e\n" +
	"\x02ip\x18\x03 \x03(\tR\x02ip\x12%\n" +
	"\x0eproxied_domain\x18\x04 \x01(\tR\rproxiedDomain\"\xd6\x03\n" +
	"\x0fDnsHijackConfig\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12)\n" +
	"\x10direct_addresses\x18\x02 \x03(\tR\x0fdirectAddresses\x12&\n" +
	"\x0fuse_default_dns\x18\x03 \x01(\bR\ruseDefaultDns\x12'\n" +
	"\x0fproxy_addresses\x18\x04 \x03(\tR\x0eproxyAddresses\x12$\n" +
	"\x0edns_conn_proxy\x18\x05 \x01(\tR\fdnsConnProxy\x12&\n" +
	"\x0fdns_conn_direct\x18\x06 \x01(\tR\rdnsConnDirect\x12(\n" +
	"\x10dns_server_proxy\x18\a \x01(\tR\x0ednsServerProxy\x12\x1d\n" +
	"\n" +
	"proxy_tags\x18\b \x03(\tR\tproxyTags\x12N\n" +
	"%direct_return_nil_aaaa_if_not_support\x18\t \x01(\bR\x1fdirectReturnNilAaaaIfNotSupport\x12L\n" +
	"$proxy_return_nil_aaaa_if_not_support\x18\n" +
	" \x01(\bR\x1eproxyReturnNilAaaaIfNotSupport\"\x8c\x01\n" +
	"\rFakeDnsServer\x12>\n" +
	"\fpool_configs\x18\x01 \x03(\v2\x1b.x.FakeDnsServer.PoolConfigR\vpoolConfigs\x1a;\n" +
	"\n" +
	"PoolConfig\x12\x12\n" +
	"\x04cidr\x18\x01 \x01(\tR\x04cidr\x12\x19\n" +
	"\blru_size\x18\x02 \x01(\rR\alruSize**\n" +
	"\aDnsType\x12\r\n" +
	"\tDnsType_A\x10\x00\x12\x10\n" +
	"\fDnsType_AAAA\x10\x01B*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_dns_proto_rawDescOnce sync.Once
	file_protos_dns_proto_rawDescData []byte
)

func file_protos_dns_proto_rawDescGZIP() []byte {
	file_protos_dns_proto_rawDescOnce.Do(func() {
		file_protos_dns_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_dns_proto_rawDesc), len(file_protos_dns_proto_rawDesc)))
	})
	return file_protos_dns_proto_rawDescData
}

var file_protos_dns_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protos_dns_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_protos_dns_proto_goTypes = []any{
	(DnsType)(0),                     // 0: x.DnsType
	(*DnsConfig)(nil),                // 1: x.DnsConfig
	(*DnsRules)(nil),                 // 2: x.DnsRules
	(*DnsRuleConfig)(nil),            // 3: x.DnsRuleConfig
	(*DnsServerConfig)(nil),          // 4: x.DnsServerConfig
	(*PlainDnsServer)(nil),           // 5: x.PlainDnsServer
	(*TlsDnsServer)(nil),             // 6: x.TlsDnsServer
	(*DohDnsServer)(nil),             // 7: x.DohDnsServer
	(*QuicDnsServer)(nil),            // 8: x.QuicDnsServer
	(*Record)(nil),                   // 9: x.Record
	(*DnsHijackConfig)(nil),          // 10: x.DnsHijackConfig
	(*FakeDnsServer)(nil),            // 11: x.FakeDnsServer
	(*FakeDnsServer_PoolConfig)(nil), // 12: x.FakeDnsServer.PoolConfig
	(*geo.Domain)(nil),               // 13: x.common.geo.Domain
}
var file_protos_dns_proto_depIdxs = []int32{
	9,  // 0: x.DnsConfig.records:type_name -> x.Record
	4,  // 1: x.DnsConfig.dns_servers:type_name -> x.DnsServerConfig
	3,  // 2: x.DnsConfig.dns_rules:type_name -> x.DnsRuleConfig
	3,  // 3: x.DnsRules.rules:type_name -> x.DnsRuleConfig
	13, // 4: x.DnsRuleConfig.domains:type_name -> x.common.geo.Domain
	0,  // 5: x.DnsRuleConfig.included_types:type_name -> x.DnsType
	5,  // 6: x.DnsServerConfig.plain_dns_server:type_name -> x.PlainDnsServer
	7,  // 7: x.DnsServerConfig.doh_dns_server:type_name -> x.DohDnsServer
	8,  // 8: x.DnsServerConfig.quic_dns_server:type_name -> x.QuicDnsServer
	11, // 9: x.DnsServerConfig.fake_dns_server:type_name -> x.FakeDnsServer
	6,  // 10: x.DnsServerConfig.tls_dns_server:type_name -> x.TlsDnsServer
	12, // 11: x.FakeDnsServer.pool_configs:type_name -> x.FakeDnsServer.PoolConfig
	12, // [12:12] is the sub-list for method output_type
	12, // [12:12] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_protos_dns_proto_init() }
func file_protos_dns_proto_init() {
	if File_protos_dns_proto != nil {
		return
	}
	file_protos_dns_proto_msgTypes[3].OneofWrappers = []any{
		(*DnsServerConfig_PlainDnsServer)(nil),
		(*DnsServerConfig_DohDnsServer)(nil),
		(*DnsServerConfig_QuicDnsServer)(nil),
		(*DnsServerConfig_FakeDnsServer)(nil),
		(*DnsServerConfig_TlsDnsServer)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_dns_proto_rawDesc), len(file_protos_dns_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_dns_proto_goTypes,
		DependencyIndexes: file_protos_dns_proto_depIdxs,
		EnumInfos:         file_protos_dns_proto_enumTypes,
		MessageInfos:      file_protos_dns_proto_msgTypes,
	}.Build()
	File_protos_dns_proto = out.File
	file_protos_dns_proto_goTypes = nil
	file_protos_dns_proto_depIdxs = nil
}
