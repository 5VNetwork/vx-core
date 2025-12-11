package configs

import (
	geo "github.com/5vnetwork/vx-core/common/geo"
	net "github.com/5vnetwork/vx-core/common/net"
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

// how to select handlers to be used
type SelectorConfig_SelectingStrategy int32

const (
	SelectorConfig_ALL             SelectorConfig_SelectingStrategy = 0
	SelectorConfig_ALL_OK          SelectorConfig_SelectingStrategy = 1
	SelectorConfig_LEAST_PING      SelectorConfig_SelectingStrategy = 2
	SelectorConfig_MOST_THROUGHPUT SelectorConfig_SelectingStrategy = 3
	// select good ones
	SelectorConfig_TOP_PING       SelectorConfig_SelectingStrategy = 4
	SelectorConfig_TOP_THROUGHPUT SelectorConfig_SelectingStrategy = 5
)

// Enum value maps for SelectorConfig_SelectingStrategy.
var (
	SelectorConfig_SelectingStrategy_name = map[int32]string{
		0: "ALL",
		1: "ALL_OK",
		2: "LEAST_PING",
		3: "MOST_THROUGHPUT",
		4: "TOP_PING",
		5: "TOP_THROUGHPUT",
	}
	SelectorConfig_SelectingStrategy_value = map[string]int32{
		"ALL":             0,
		"ALL_OK":          1,
		"LEAST_PING":      2,
		"MOST_THROUGHPUT": 3,
		"TOP_PING":        4,
		"TOP_THROUGHPUT":  5,
	}
)

func (x SelectorConfig_SelectingStrategy) Enum() *SelectorConfig_SelectingStrategy {
	p := new(SelectorConfig_SelectingStrategy)
	*p = x
	return p
}

func (x SelectorConfig_SelectingStrategy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SelectorConfig_SelectingStrategy) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_router_proto_enumTypes[0].Descriptor()
}

func (SelectorConfig_SelectingStrategy) Type() protoreflect.EnumType {
	return &file_protos_router_proto_enumTypes[0]
}

func (x SelectorConfig_SelectingStrategy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SelectorConfig_SelectingStrategy.Descriptor instead.
func (SelectorConfig_SelectingStrategy) EnumDescriptor() ([]byte, []int) {
	return file_protos_router_proto_rawDescGZIP(), []int{3, 0}
}

// if there are many handlers, how to balance them
type SelectorConfig_BalanceStrategy int32

const (
	SelectorConfig_RANDOM SelectorConfig_BalanceStrategy = 0
	// balance based on app first, if no app, based on root domain
	SelectorConfig_MEMORY SelectorConfig_BalanceStrategy = 1
)

// Enum value maps for SelectorConfig_BalanceStrategy.
var (
	SelectorConfig_BalanceStrategy_name = map[int32]string{
		0: "RANDOM",
		1: "MEMORY",
	}
	SelectorConfig_BalanceStrategy_value = map[string]int32{
		"RANDOM": 0,
		"MEMORY": 1,
	}
)

func (x SelectorConfig_BalanceStrategy) Enum() *SelectorConfig_BalanceStrategy {
	p := new(SelectorConfig_BalanceStrategy)
	*p = x
	return p
}

func (x SelectorConfig_BalanceStrategy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SelectorConfig_BalanceStrategy) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_router_proto_enumTypes[1].Descriptor()
}

func (SelectorConfig_BalanceStrategy) Type() protoreflect.EnumType {
	return &file_protos_router_proto_enumTypes[1]
}

func (x SelectorConfig_BalanceStrategy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SelectorConfig_BalanceStrategy.Descriptor instead.
func (SelectorConfig_BalanceStrategy) EnumDescriptor() ([]byte, []int) {
	return file_protos_router_proto_rawDescGZIP(), []int{3, 1}
}

type AppId_Type int32

const (
	// The value is used as is. "keyword"
	AppId_Keyword AppId_Type = 0
	AppId_Prefix  AppId_Type = 1
	AppId_Exact   AppId_Type = 2
)

// Enum value maps for AppId_Type.
var (
	AppId_Type_name = map[int32]string{
		0: "Keyword",
		1: "Prefix",
		2: "Exact",
	}
	AppId_Type_value = map[string]int32{
		"Keyword": 0,
		"Prefix":  1,
		"Exact":   2,
	}
)

func (x AppId_Type) Enum() *AppId_Type {
	p := new(AppId_Type)
	*p = x
	return p
}

func (x AppId_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AppId_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_router_proto_enumTypes[2].Descriptor()
}

func (AppId_Type) Type() protoreflect.EnumType {
	return &file_protos_router_proto_enumTypes[2]
}

func (x AppId_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AppId_Type.Descriptor instead.
func (AppId_Type) EnumDescriptor() ([]byte, []int) {
	return file_protos_router_proto_rawDescGZIP(), []int{4, 0}
}

type RouterConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// defaults to As_Is
	Rules         []*RuleConfig `protobuf:"bytes,1,rep,name=rules,proto3" json:"rules,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RouterConfig) Reset() {
	*x = RouterConfig{}
	mi := &file_protos_router_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RouterConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RouterConfig) ProtoMessage() {}

func (x *RouterConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_router_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RouterConfig.ProtoReflect.Descriptor instead.
func (*RouterConfig) Descriptor() ([]byte, []int) {
	return file_protos_router_proto_rawDescGZIP(), []int{0}
}

func (x *RouterConfig) GetRules() []*RuleConfig {
	if x != nil {
		return x.Rules
	}
	return nil
}

type SelectorsConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Selectors     []*SelectorConfig      `protobuf:"bytes,1,rep,name=selectors,proto3" json:"selectors,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SelectorsConfig) Reset() {
	*x = SelectorsConfig{}
	mi := &file_protos_router_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SelectorsConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SelectorsConfig) ProtoMessage() {}

func (x *SelectorsConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_router_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SelectorsConfig.ProtoReflect.Descriptor instead.
func (*SelectorsConfig) Descriptor() ([]byte, []int) {
	return file_protos_router_proto_rawDescGZIP(), []int{1}
}

func (x *SelectorsConfig) GetSelectors() []*SelectorConfig {
	if x != nil {
		return x.Selectors
	}
	return nil
}

type RuleConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// either outbound_tag or banlancer_tag should be specified but not both
	// when both are empty, it means blocks
	OutboundTag string `protobuf:"bytes,1,opt,name=outbound_tag,json=outboundTag,proto3" json:"outbound_tag,omitempty"`
	SelectorTag string `protobuf:"bytes,2,opt,name=selector_tag,json=selectorTag,proto3" json:"selector_tag,omitempty"`
	// used to match source ip
	SrcCidrs []string `protobuf:"bytes,5,rep,name=src_cidrs,json=srcCidrs,proto3" json:"src_cidrs,omitempty"`
	// used to match source ip
	SrcIpTags []string `protobuf:"bytes,6,rep,name=src_ip_tags,json=srcIpTags,proto3" json:"src_ip_tags,omitempty"`
	// used to match dst ip
	DstCidrs []string `protobuf:"bytes,8,rep,name=dst_cidrs,json=dstCidrs,proto3" json:"dst_cidrs,omitempty"`
	// used to match dst ip
	DstIpTags []string `protobuf:"bytes,9,rep,name=dst_ip_tags,json=dstIpTags,proto3" json:"dst_ip_tags,omitempty"`
	// resolve domain to ip when ip is not available
	ResolveDomain bool `protobuf:"varint,26,opt,name=resolve_domain,json=resolveDomain,proto3" json:"resolve_domain,omitempty"`
	// used to match domain
	GeoDomains []*geo.Domain `protobuf:"bytes,11,rep,name=geo_domains,json=geoDomains,proto3" json:"geo_domains,omitempty"`
	DomainTags []string      `protobuf:"bytes,13,rep,name=domain_tags,json=domainTags,proto3" json:"domain_tags,omitempty"`
	// skip sniff for connectiosn use ip targets
	SkipSniff     bool             `protobuf:"varint,27,opt,name=skip_sniff,json=skipSniff,proto3" json:"skip_sniff,omitempty"`
	Usernames     []string         `protobuf:"bytes,14,rep,name=usernames,proto3" json:"usernames,omitempty"`
	InboundTags   []string         `protobuf:"bytes,15,rep,name=inbound_tags,json=inboundTags,proto3" json:"inbound_tags,omitempty"`
	Networks      []net.Network    `protobuf:"varint,16,rep,packed,name=networks,proto3,enum=x.common.net.Network" json:"networks,omitempty"`
	SrcPortRanges []*net.PortRange `protobuf:"bytes,17,rep,name=src_port_ranges,json=srcPortRanges,proto3" json:"src_port_ranges,omitempty"`
	DstPortRanges []*net.PortRange `protobuf:"bytes,18,rep,name=dst_port_ranges,json=dstPortRanges,proto3" json:"dst_port_ranges,omitempty"`
	AppIds        []*AppId         `protobuf:"bytes,19,rep,name=app_ids,json=appIds,proto3" json:"app_ids,omitempty"`
	Ipv6          bool             `protobuf:"varint,20,opt,name=ipv6,proto3" json:"ipv6,omitempty"`
	// for debugging
	RuleName      string   `protobuf:"bytes,21,opt,name=rule_name,json=ruleName,proto3" json:"rule_name,omitempty"`
	FakeIp        bool     `protobuf:"varint,22,opt,name=fake_ip,json=fakeIp,proto3" json:"fake_ip,omitempty"`
	MatchAll      bool     `protobuf:"varint,23,opt,name=match_all,json=matchAll,proto3" json:"match_all,omitempty"`
	AppTags       []string `protobuf:"bytes,24,rep,name=app_tags,json=appTags,proto3" json:"app_tags,omitempty"`
	AllTags       []string `protobuf:"bytes,25,rep,name=all_tags,json=allTags,proto3" json:"all_tags,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RuleConfig) Reset() {
	*x = RuleConfig{}
	mi := &file_protos_router_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RuleConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RuleConfig) ProtoMessage() {}

func (x *RuleConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_router_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RuleConfig.ProtoReflect.Descriptor instead.
func (*RuleConfig) Descriptor() ([]byte, []int) {
	return file_protos_router_proto_rawDescGZIP(), []int{2}
}

func (x *RuleConfig) GetOutboundTag() string {
	if x != nil {
		return x.OutboundTag
	}
	return ""
}

func (x *RuleConfig) GetSelectorTag() string {
	if x != nil {
		return x.SelectorTag
	}
	return ""
}

func (x *RuleConfig) GetSrcCidrs() []string {
	if x != nil {
		return x.SrcCidrs
	}
	return nil
}

func (x *RuleConfig) GetSrcIpTags() []string {
	if x != nil {
		return x.SrcIpTags
	}
	return nil
}

func (x *RuleConfig) GetDstCidrs() []string {
	if x != nil {
		return x.DstCidrs
	}
	return nil
}

func (x *RuleConfig) GetDstIpTags() []string {
	if x != nil {
		return x.DstIpTags
	}
	return nil
}

func (x *RuleConfig) GetResolveDomain() bool {
	if x != nil {
		return x.ResolveDomain
	}
	return false
}

func (x *RuleConfig) GetGeoDomains() []*geo.Domain {
	if x != nil {
		return x.GeoDomains
	}
	return nil
}

func (x *RuleConfig) GetDomainTags() []string {
	if x != nil {
		return x.DomainTags
	}
	return nil
}

func (x *RuleConfig) GetSkipSniff() bool {
	if x != nil {
		return x.SkipSniff
	}
	return false
}

func (x *RuleConfig) GetUsernames() []string {
	if x != nil {
		return x.Usernames
	}
	return nil
}

func (x *RuleConfig) GetInboundTags() []string {
	if x != nil {
		return x.InboundTags
	}
	return nil
}

func (x *RuleConfig) GetNetworks() []net.Network {
	if x != nil {
		return x.Networks
	}
	return nil
}

func (x *RuleConfig) GetSrcPortRanges() []*net.PortRange {
	if x != nil {
		return x.SrcPortRanges
	}
	return nil
}

func (x *RuleConfig) GetDstPortRanges() []*net.PortRange {
	if x != nil {
		return x.DstPortRanges
	}
	return nil
}

func (x *RuleConfig) GetAppIds() []*AppId {
	if x != nil {
		return x.AppIds
	}
	return nil
}

func (x *RuleConfig) GetIpv6() bool {
	if x != nil {
		return x.Ipv6
	}
	return false
}

func (x *RuleConfig) GetRuleName() string {
	if x != nil {
		return x.RuleName
	}
	return ""
}

func (x *RuleConfig) GetFakeIp() bool {
	if x != nil {
		return x.FakeIp
	}
	return false
}

func (x *RuleConfig) GetMatchAll() bool {
	if x != nil {
		return x.MatchAll
	}
	return false
}

func (x *RuleConfig) GetAppTags() []string {
	if x != nil {
		return x.AppTags
	}
	return nil
}

func (x *RuleConfig) GetAllTags() []string {
	if x != nil {
		return x.AllTags
	}
	return nil
}

type SelectorConfig struct {
	state           protoimpl.MessageState           `protogen:"open.v1"`
	Tag             string                           `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Filter          *SelectorConfig_Filter           `protobuf:"bytes,2,opt,name=filter,proto3" json:"filter,omitempty"`
	Strategy        SelectorConfig_SelectingStrategy `protobuf:"varint,3,opt,name=strategy,proto3,enum=x.SelectorConfig_SelectingStrategy" json:"strategy,omitempty"`
	BalanceStrategy SelectorConfig_BalanceStrategy   `protobuf:"varint,4,opt,name=balance_strategy,json=balanceStrategy,proto3,enum=x.SelectorConfig_BalanceStrategy" json:"balance_strategy,omitempty"`
	// if not empty, these handlers will be used as land handlers
	// ids of the outbound handler
	LandHandlers  []int64 `protobuf:"varint,6,rep,packed,name=land_handlers,json=landHandlers,proto3" json:"land_handlers,omitempty"`
	SelectFromOm  bool    `protobuf:"varint,7,opt,name=select_from_om,json=selectFromOm,proto3" json:"select_from_om,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SelectorConfig) Reset() {
	*x = SelectorConfig{}
	mi := &file_protos_router_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SelectorConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SelectorConfig) ProtoMessage() {}

func (x *SelectorConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_router_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SelectorConfig.ProtoReflect.Descriptor instead.
func (*SelectorConfig) Descriptor() ([]byte, []int) {
	return file_protos_router_proto_rawDescGZIP(), []int{3}
}

func (x *SelectorConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *SelectorConfig) GetFilter() *SelectorConfig_Filter {
	if x != nil {
		return x.Filter
	}
	return nil
}

func (x *SelectorConfig) GetStrategy() SelectorConfig_SelectingStrategy {
	if x != nil {
		return x.Strategy
	}
	return SelectorConfig_ALL
}

func (x *SelectorConfig) GetBalanceStrategy() SelectorConfig_BalanceStrategy {
	if x != nil {
		return x.BalanceStrategy
	}
	return SelectorConfig_RANDOM
}

func (x *SelectorConfig) GetLandHandlers() []int64 {
	if x != nil {
		return x.LandHandlers
	}
	return nil
}

func (x *SelectorConfig) GetSelectFromOm() bool {
	if x != nil {
		return x.SelectFromOm
	}
	return false
}

type AppId struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Type          AppId_Type             `protobuf:"varint,1,opt,name=type,proto3,enum=x.AppId_Type" json:"type,omitempty"`
	Value         string                 `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AppId) Reset() {
	*x = AppId{}
	mi := &file_protos_router_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppId) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppId) ProtoMessage() {}

func (x *AppId) ProtoReflect() protoreflect.Message {
	mi := &file_protos_router_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppId.ProtoReflect.Descriptor instead.
func (*AppId) Descriptor() ([]byte, []int) {
	return file_protos_router_proto_rawDescGZIP(), []int{4}
}

func (x *AppId) GetType() AppId_Type {
	if x != nil {
		return x.Type
	}
	return AppId_Keyword
}

func (x *AppId) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type SelectorConfig_Filter struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// If an outbound's tag has prefix of any of the prefixes, match!
	Prefixes []string `protobuf:"bytes,1,rep,name=prefixes,proto3" json:"prefixes,omitempty"`
	// A outbound handler will match if its tag is one of the tags
	Tags      []string `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty"`
	GroupTags []string `protobuf:"bytes,3,rep,name=group_tags,json=groupTags,proto3" json:"group_tags,omitempty"`
	// If true, a handler will be selected if it does not match all conditions
	Inverse       bool    `protobuf:"varint,4,opt,name=inverse,proto3" json:"inverse,omitempty"`
	SubIds        []int64 `protobuf:"varint,5,rep,packed,name=sub_ids,json=subIds,proto3" json:"sub_ids,omitempty"`
	HandlerIds    []int64 `protobuf:"varint,6,rep,packed,name=handler_ids,json=handlerIds,proto3" json:"handler_ids,omitempty"`
	Selected      bool    `protobuf:"varint,7,opt,name=selected,proto3" json:"selected,omitempty"`
	All           bool    `protobuf:"varint,8,opt,name=all,proto3" json:"all,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SelectorConfig_Filter) Reset() {
	*x = SelectorConfig_Filter{}
	mi := &file_protos_router_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SelectorConfig_Filter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SelectorConfig_Filter) ProtoMessage() {}

func (x *SelectorConfig_Filter) ProtoReflect() protoreflect.Message {
	mi := &file_protos_router_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SelectorConfig_Filter.ProtoReflect.Descriptor instead.
func (*SelectorConfig_Filter) Descriptor() ([]byte, []int) {
	return file_protos_router_proto_rawDescGZIP(), []int{3, 0}
}

func (x *SelectorConfig_Filter) GetPrefixes() []string {
	if x != nil {
		return x.Prefixes
	}
	return nil
}

func (x *SelectorConfig_Filter) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *SelectorConfig_Filter) GetGroupTags() []string {
	if x != nil {
		return x.GroupTags
	}
	return nil
}

func (x *SelectorConfig_Filter) GetInverse() bool {
	if x != nil {
		return x.Inverse
	}
	return false
}

func (x *SelectorConfig_Filter) GetSubIds() []int64 {
	if x != nil {
		return x.SubIds
	}
	return nil
}

func (x *SelectorConfig_Filter) GetHandlerIds() []int64 {
	if x != nil {
		return x.HandlerIds
	}
	return nil
}

func (x *SelectorConfig_Filter) GetSelected() bool {
	if x != nil {
		return x.Selected
	}
	return false
}

func (x *SelectorConfig_Filter) GetAll() bool {
	if x != nil {
		return x.All
	}
	return false
}

var File_protos_router_proto protoreflect.FileDescriptor

const file_protos_router_proto_rawDesc = "" +
	"\n" +
	"\x13protos/router.proto\x12\x01x\x1a\x14common/geo/geo.proto\x1a\x14common/net/net.proto\"3\n" +
	"\fRouterConfig\x12#\n" +
	"\x05rules\x18\x01 \x03(\v2\r.x.RuleConfigR\x05rules\"B\n" +
	"\x0fSelectorsConfig\x12/\n" +
	"\tselectors\x18\x01 \x03(\v2\x11.x.SelectorConfigR\tselectors\"\xa0\x06\n" +
	"\n" +
	"RuleConfig\x12!\n" +
	"\foutbound_tag\x18\x01 \x01(\tR\voutboundTag\x12!\n" +
	"\fselector_tag\x18\x02 \x01(\tR\vselectorTag\x12\x1b\n" +
	"\tsrc_cidrs\x18\x05 \x03(\tR\bsrcCidrs\x12\x1e\n" +
	"\vsrc_ip_tags\x18\x06 \x03(\tR\tsrcIpTags\x12\x1b\n" +
	"\tdst_cidrs\x18\b \x03(\tR\bdstCidrs\x12\x1e\n" +
	"\vdst_ip_tags\x18\t \x03(\tR\tdstIpTags\x12%\n" +
	"\x0eresolve_domain\x18\x1a \x01(\bR\rresolveDomain\x125\n" +
	"\vgeo_domains\x18\v \x03(\v2\x14.x.common.geo.DomainR\n" +
	"geoDomains\x12\x1f\n" +
	"\vdomain_tags\x18\r \x03(\tR\n" +
	"domainTags\x12\x1d\n" +
	"\n" +
	"skip_sniff\x18\x1b \x01(\bR\tskipSniff\x12\x1c\n" +
	"\tusernames\x18\x0e \x03(\tR\tusernames\x12!\n" +
	"\finbound_tags\x18\x0f \x03(\tR\vinboundTags\x121\n" +
	"\bnetworks\x18\x10 \x03(\x0e2\x15.x.common.net.NetworkR\bnetworks\x12?\n" +
	"\x0fsrc_port_ranges\x18\x11 \x03(\v2\x17.x.common.net.PortRangeR\rsrcPortRanges\x12?\n" +
	"\x0fdst_port_ranges\x18\x12 \x03(\v2\x17.x.common.net.PortRangeR\rdstPortRanges\x12!\n" +
	"\aapp_ids\x18\x13 \x03(\v2\b.x.AppIdR\x06appIds\x12\x12\n" +
	"\x04ipv6\x18\x14 \x01(\bR\x04ipv6\x12\x1b\n" +
	"\trule_name\x18\x15 \x01(\tR\bruleName\x12\x17\n" +
	"\afake_ip\x18\x16 \x01(\bR\x06fakeIp\x12\x1b\n" +
	"\tmatch_all\x18\x17 \x01(\bR\bmatchAll\x12\x19\n" +
	"\bapp_tags\x18\x18 \x03(\tR\aappTags\x12\x19\n" +
	"\ball_tags\x18\x19 \x03(\tR\aallTags\"\xa6\x05\n" +
	"\x0eSelectorConfig\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x120\n" +
	"\x06filter\x18\x02 \x01(\v2\x18.x.SelectorConfig.FilterR\x06filter\x12?\n" +
	"\bstrategy\x18\x03 \x01(\x0e2#.x.SelectorConfig.SelectingStrategyR\bstrategy\x12L\n" +
	"\x10balance_strategy\x18\x04 \x01(\x0e2!.x.SelectorConfig.BalanceStrategyR\x0fbalanceStrategy\x12#\n" +
	"\rland_handlers\x18\x06 \x03(\x03R\flandHandlers\x12$\n" +
	"\x0eselect_from_om\x18\a \x01(\bR\fselectFromOm\x1a\xd9\x01\n" +
	"\x06Filter\x12\x1a\n" +
	"\bprefixes\x18\x01 \x03(\tR\bprefixes\x12\x12\n" +
	"\x04tags\x18\x02 \x03(\tR\x04tags\x12\x1d\n" +
	"\n" +
	"group_tags\x18\x03 \x03(\tR\tgroupTags\x12\x18\n" +
	"\ainverse\x18\x04 \x01(\bR\ainverse\x12\x17\n" +
	"\asub_ids\x18\x05 \x03(\x03R\x06subIds\x12\x1f\n" +
	"\vhandler_ids\x18\x06 \x03(\x03R\n" +
	"handlerIds\x12\x1a\n" +
	"\bselected\x18\a \x01(\bR\bselected\x12\x10\n" +
	"\x03all\x18\b \x01(\bR\x03all\"o\n" +
	"\x11SelectingStrategy\x12\a\n" +
	"\x03ALL\x10\x00\x12\n" +
	"\n" +
	"\x06ALL_OK\x10\x01\x12\x0e\n" +
	"\n" +
	"LEAST_PING\x10\x02\x12\x13\n" +
	"\x0fMOST_THROUGHPUT\x10\x03\x12\f\n" +
	"\bTOP_PING\x10\x04\x12\x12\n" +
	"\x0eTOP_THROUGHPUT\x10\x05\")\n" +
	"\x0fBalanceStrategy\x12\n" +
	"\n" +
	"\x06RANDOM\x10\x00\x12\n" +
	"\n" +
	"\x06MEMORY\x10\x01\"l\n" +
	"\x05AppId\x12!\n" +
	"\x04type\x18\x01 \x01(\x0e2\r.x.AppId.TypeR\x04type\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value\"*\n" +
	"\x04Type\x12\v\n" +
	"\aKeyword\x10\x00\x12\n" +
	"\n" +
	"\x06Prefix\x10\x01\x12\t\n" +
	"\x05Exact\x10\x02B*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_router_proto_rawDescOnce sync.Once
	file_protos_router_proto_rawDescData []byte
)

func file_protos_router_proto_rawDescGZIP() []byte {
	file_protos_router_proto_rawDescOnce.Do(func() {
		file_protos_router_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_router_proto_rawDesc), len(file_protos_router_proto_rawDesc)))
	})
	return file_protos_router_proto_rawDescData
}

var file_protos_router_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_protos_router_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_protos_router_proto_goTypes = []any{
	(SelectorConfig_SelectingStrategy)(0), // 0: x.SelectorConfig.SelectingStrategy
	(SelectorConfig_BalanceStrategy)(0),   // 1: x.SelectorConfig.BalanceStrategy
	(AppId_Type)(0),                       // 2: x.AppId.Type
	(*RouterConfig)(nil),                  // 3: x.RouterConfig
	(*SelectorsConfig)(nil),               // 4: x.SelectorsConfig
	(*RuleConfig)(nil),                    // 5: x.RuleConfig
	(*SelectorConfig)(nil),                // 6: x.SelectorConfig
	(*AppId)(nil),                         // 7: x.AppId
	(*SelectorConfig_Filter)(nil),         // 8: x.SelectorConfig.Filter
	(*geo.Domain)(nil),                    // 9: x.common.geo.Domain
	(net.Network)(0),                      // 10: x.common.net.Network
	(*net.PortRange)(nil),                 // 11: x.common.net.PortRange
}
var file_protos_router_proto_depIdxs = []int32{
	5,  // 0: x.RouterConfig.rules:type_name -> x.RuleConfig
	6,  // 1: x.SelectorsConfig.selectors:type_name -> x.SelectorConfig
	9,  // 2: x.RuleConfig.geo_domains:type_name -> x.common.geo.Domain
	10, // 3: x.RuleConfig.networks:type_name -> x.common.net.Network
	11, // 4: x.RuleConfig.src_port_ranges:type_name -> x.common.net.PortRange
	11, // 5: x.RuleConfig.dst_port_ranges:type_name -> x.common.net.PortRange
	7,  // 6: x.RuleConfig.app_ids:type_name -> x.AppId
	8,  // 7: x.SelectorConfig.filter:type_name -> x.SelectorConfig.Filter
	0,  // 8: x.SelectorConfig.strategy:type_name -> x.SelectorConfig.SelectingStrategy
	1,  // 9: x.SelectorConfig.balance_strategy:type_name -> x.SelectorConfig.BalanceStrategy
	2,  // 10: x.AppId.type:type_name -> x.AppId.Type
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_protos_router_proto_init() }
func file_protos_router_proto_init() {
	if File_protos_router_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_router_proto_rawDesc), len(file_protos_router_proto_rawDesc)),
			NumEnums:      3,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_router_proto_goTypes,
		DependencyIndexes: file_protos_router_proto_depIdxs,
		EnumInfos:         file_protos_router_proto_enumTypes,
		MessageInfos:      file_protos_router_proto_msgTypes,
	}.Build()
	File_protos_router_proto = out.File
	file_protos_router_proto_goTypes = nil
	file_protos_router_proto_depIdxs = nil
}
