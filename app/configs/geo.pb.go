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

type GreatDomainSetConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	OppositeName  string                 `protobuf:"bytes,4,opt,name=opposite_name,json=oppositeName,proto3" json:"opposite_name,omitempty"`
	ExNames       []string               `protobuf:"bytes,5,rep,name=ex_names,json=exNames,proto3" json:"ex_names,omitempty"`
	InNames       []string               `protobuf:"bytes,6,rep,name=in_names,json=inNames,proto3" json:"in_names,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GreatDomainSetConfig) Reset() {
	*x = GreatDomainSetConfig{}
	mi := &file_protos_geo_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GreatDomainSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GreatDomainSetConfig) ProtoMessage() {}

func (x *GreatDomainSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_geo_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GreatDomainSetConfig.ProtoReflect.Descriptor instead.
func (*GreatDomainSetConfig) Descriptor() ([]byte, []int) {
	return file_protos_geo_proto_rawDescGZIP(), []int{0}
}

func (x *GreatDomainSetConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GreatDomainSetConfig) GetOppositeName() string {
	if x != nil {
		return x.OppositeName
	}
	return ""
}

func (x *GreatDomainSetConfig) GetExNames() []string {
	if x != nil {
		return x.ExNames
	}
	return nil
}

func (x *GreatDomainSetConfig) GetInNames() []string {
	if x != nil {
		return x.InNames
	}
	return nil
}

type AtomicDomainSetConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Name  string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// either one of the following two fields should be set
	Domains        []*geo.Domain  `protobuf:"bytes,2,rep,name=domains,proto3" json:"domains,omitempty"`
	Geosite        *GeositeConfig `protobuf:"bytes,3,opt,name=geosite,proto3" json:"geosite,omitempty"`
	UseBloomFilter bool           `protobuf:"varint,4,opt,name=use_bloom_filter,json=useBloomFilter,proto3" json:"use_bloom_filter,omitempty"`
	ClashFiles     []string       `protobuf:"bytes,5,rep,name=clash_files,json=clashFiles,proto3" json:"clash_files,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *AtomicDomainSetConfig) Reset() {
	*x = AtomicDomainSetConfig{}
	mi := &file_protos_geo_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AtomicDomainSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AtomicDomainSetConfig) ProtoMessage() {}

func (x *AtomicDomainSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_geo_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AtomicDomainSetConfig.ProtoReflect.Descriptor instead.
func (*AtomicDomainSetConfig) Descriptor() ([]byte, []int) {
	return file_protos_geo_proto_rawDescGZIP(), []int{1}
}

func (x *AtomicDomainSetConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *AtomicDomainSetConfig) GetDomains() []*geo.Domain {
	if x != nil {
		return x.Domains
	}
	return nil
}

func (x *AtomicDomainSetConfig) GetGeosite() *GeositeConfig {
	if x != nil {
		return x.Geosite
	}
	return nil
}

func (x *AtomicDomainSetConfig) GetUseBloomFilter() bool {
	if x != nil {
		return x.UseBloomFilter
	}
	return false
}

func (x *AtomicDomainSetConfig) GetClashFiles() []string {
	if x != nil {
		return x.ClashFiles
	}
	return nil
}

type DomainSetConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Tags          []string               `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DomainSetConfig) Reset() {
	*x = DomainSetConfig{}
	mi := &file_protos_geo_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DomainSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DomainSetConfig) ProtoMessage() {}

func (x *DomainSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_geo_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DomainSetConfig.ProtoReflect.Descriptor instead.
func (*DomainSetConfig) Descriptor() ([]byte, []int) {
	return file_protos_geo_proto_rawDescGZIP(), []int{2}
}

func (x *DomainSetConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *DomainSetConfig) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

type GreatIPSetConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	OppositeName  string                 `protobuf:"bytes,4,opt,name=opposite_name,json=oppositeName,proto3" json:"opposite_name,omitempty"`
	ExNames       []string               `protobuf:"bytes,5,rep,name=ex_names,json=exNames,proto3" json:"ex_names,omitempty"`
	InNames       []string               `protobuf:"bytes,6,rep,name=in_names,json=inNames,proto3" json:"in_names,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GreatIPSetConfig) Reset() {
	*x = GreatIPSetConfig{}
	mi := &file_protos_geo_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GreatIPSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GreatIPSetConfig) ProtoMessage() {}

func (x *GreatIPSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_geo_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GreatIPSetConfig.ProtoReflect.Descriptor instead.
func (*GreatIPSetConfig) Descriptor() ([]byte, []int) {
	return file_protos_geo_proto_rawDescGZIP(), []int{3}
}

func (x *GreatIPSetConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GreatIPSetConfig) GetOppositeName() string {
	if x != nil {
		return x.OppositeName
	}
	return ""
}

func (x *GreatIPSetConfig) GetExNames() []string {
	if x != nil {
		return x.ExNames
	}
	return nil
}

func (x *GreatIPSetConfig) GetInNames() []string {
	if x != nil {
		return x.InNames
	}
	return nil
}

type AtomicIPSetConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Cidrs         []*geo.CIDR            `protobuf:"bytes,2,rep,name=cidrs,proto3" json:"cidrs,omitempty"`
	Geoip         *GeoIPConfig           `protobuf:"bytes,1,opt,name=geoip,proto3" json:"geoip,omitempty"`
	Inverse       bool                   `protobuf:"varint,4,opt,name=inverse,proto3" json:"inverse,omitempty"`
	ClashFiles    []string               `protobuf:"bytes,5,rep,name=clash_files,json=clashFiles,proto3" json:"clash_files,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AtomicIPSetConfig) Reset() {
	*x = AtomicIPSetConfig{}
	mi := &file_protos_geo_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AtomicIPSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AtomicIPSetConfig) ProtoMessage() {}

func (x *AtomicIPSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_geo_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AtomicIPSetConfig.ProtoReflect.Descriptor instead.
func (*AtomicIPSetConfig) Descriptor() ([]byte, []int) {
	return file_protos_geo_proto_rawDescGZIP(), []int{4}
}

func (x *AtomicIPSetConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *AtomicIPSetConfig) GetCidrs() []*geo.CIDR {
	if x != nil {
		return x.Cidrs
	}
	return nil
}

func (x *AtomicIPSetConfig) GetGeoip() *GeoIPConfig {
	if x != nil {
		return x.Geoip
	}
	return nil
}

func (x *AtomicIPSetConfig) GetInverse() bool {
	if x != nil {
		return x.Inverse
	}
	return false
}

func (x *AtomicIPSetConfig) GetClashFiles() []string {
	if x != nil {
		return x.ClashFiles
	}
	return nil
}

type GeoConfig struct {
	state      protoimpl.MessageState `protogen:"open.v1"`
	DomainSets []*DomainSetConfig     `protobuf:"bytes,1,rep,name=domain_sets,json=domainSets,proto3" json:"domain_sets,omitempty"`
	// There should be no same name in great_domain_sets and atomic_domain_sets.
	// Same for great_ip_sets and atomic_ip_sets.
	GreatDomainSets  []*GreatDomainSetConfig  `protobuf:"bytes,3,rep,name=great_domain_sets,json=greatDomainSets,proto3" json:"great_domain_sets,omitempty"`
	GreatIpSets      []*GreatIPSetConfig      `protobuf:"bytes,4,rep,name=great_ip_sets,json=greatIpSets,proto3" json:"great_ip_sets,omitempty"`
	AtomicDomainSets []*AtomicDomainSetConfig `protobuf:"bytes,5,rep,name=atomic_domain_sets,json=atomicDomainSets,proto3" json:"atomic_domain_sets,omitempty"`
	AtomicIpSets     []*AtomicIPSetConfig     `protobuf:"bytes,6,rep,name=atomic_ip_sets,json=atomicIpSets,proto3" json:"atomic_ip_sets,omitempty"`
	AppSets          []*AppSetConfig          `protobuf:"bytes,7,rep,name=app_sets,json=appSets,proto3" json:"app_sets,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *GeoConfig) Reset() {
	*x = GeoConfig{}
	mi := &file_protos_geo_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GeoConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoConfig) ProtoMessage() {}

func (x *GeoConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_geo_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoConfig.ProtoReflect.Descriptor instead.
func (*GeoConfig) Descriptor() ([]byte, []int) {
	return file_protos_geo_proto_rawDescGZIP(), []int{5}
}

func (x *GeoConfig) GetDomainSets() []*DomainSetConfig {
	if x != nil {
		return x.DomainSets
	}
	return nil
}

func (x *GeoConfig) GetGreatDomainSets() []*GreatDomainSetConfig {
	if x != nil {
		return x.GreatDomainSets
	}
	return nil
}

func (x *GeoConfig) GetGreatIpSets() []*GreatIPSetConfig {
	if x != nil {
		return x.GreatIpSets
	}
	return nil
}

func (x *GeoConfig) GetAtomicDomainSets() []*AtomicDomainSetConfig {
	if x != nil {
		return x.AtomicDomainSets
	}
	return nil
}

func (x *GeoConfig) GetAtomicIpSets() []*AtomicIPSetConfig {
	if x != nil {
		return x.AtomicIpSets
	}
	return nil
}

func (x *GeoConfig) GetAppSets() []*AppSetConfig {
	if x != nil {
		return x.AppSets
	}
	return nil
}

type GeositeConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Codes         []string               `protobuf:"bytes,1,rep,name=codes,proto3" json:"codes,omitempty"`
	Attributes    []string               `protobuf:"bytes,2,rep,name=attributes,proto3" json:"attributes,omitempty"`
	Filepath      string                 `protobuf:"bytes,3,opt,name=filepath,proto3" json:"filepath,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GeositeConfig) Reset() {
	*x = GeositeConfig{}
	mi := &file_protos_geo_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GeositeConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeositeConfig) ProtoMessage() {}

func (x *GeositeConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_geo_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeositeConfig.ProtoReflect.Descriptor instead.
func (*GeositeConfig) Descriptor() ([]byte, []int) {
	return file_protos_geo_proto_rawDescGZIP(), []int{6}
}

func (x *GeositeConfig) GetCodes() []string {
	if x != nil {
		return x.Codes
	}
	return nil
}

func (x *GeositeConfig) GetAttributes() []string {
	if x != nil {
		return x.Attributes
	}
	return nil
}

func (x *GeositeConfig) GetFilepath() string {
	if x != nil {
		return x.Filepath
	}
	return ""
}

type GeoIPConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Filepath      string                 `protobuf:"bytes,1,opt,name=filepath,proto3" json:"filepath,omitempty"`
	Codes         []string               `protobuf:"bytes,2,rep,name=codes,proto3" json:"codes,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GeoIPConfig) Reset() {
	*x = GeoIPConfig{}
	mi := &file_protos_geo_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GeoIPConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoIPConfig) ProtoMessage() {}

func (x *GeoIPConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_geo_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoIPConfig.ProtoReflect.Descriptor instead.
func (*GeoIPConfig) Descriptor() ([]byte, []int) {
	return file_protos_geo_proto_rawDescGZIP(), []int{7}
}

func (x *GeoIPConfig) GetFilepath() string {
	if x != nil {
		return x.Filepath
	}
	return ""
}

func (x *GeoIPConfig) GetCodes() []string {
	if x != nil {
		return x.Codes
	}
	return nil
}

type AppSetConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	AppIds        []*AppId               `protobuf:"bytes,2,rep,name=app_ids,json=appIds,proto3" json:"app_ids,omitempty"`
	ClashFiles    []string               `protobuf:"bytes,3,rep,name=clash_files,json=clashFiles,proto3" json:"clash_files,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AppSetConfig) Reset() {
	*x = AppSetConfig{}
	mi := &file_protos_geo_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppSetConfig) ProtoMessage() {}

func (x *AppSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_geo_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppSetConfig.ProtoReflect.Descriptor instead.
func (*AppSetConfig) Descriptor() ([]byte, []int) {
	return file_protos_geo_proto_rawDescGZIP(), []int{8}
}

func (x *AppSetConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *AppSetConfig) GetAppIds() []*AppId {
	if x != nil {
		return x.AppIds
	}
	return nil
}

func (x *AppSetConfig) GetClashFiles() []string {
	if x != nil {
		return x.ClashFiles
	}
	return nil
}

var File_protos_geo_proto protoreflect.FileDescriptor

const file_protos_geo_proto_rawDesc = "" +
	"\n" +
	"\x10protos/geo.proto\x12\x01x\x1a\x14common/geo/geo.proto\x1a\x13protos/router.proto\"\x85\x01\n" +
	"\x14GreatDomainSetConfig\x12\x12\n" +
	"\x04name\x18\x03 \x01(\tR\x04name\x12#\n" +
	"\ropposite_name\x18\x04 \x01(\tR\foppositeName\x12\x19\n" +
	"\bex_names\x18\x05 \x03(\tR\aexNames\x12\x19\n" +
	"\bin_names\x18\x06 \x03(\tR\ainNames\"\xd2\x01\n" +
	"\x15AtomicDomainSetConfig\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12.\n" +
	"\adomains\x18\x02 \x03(\v2\x14.x.common.geo.DomainR\adomains\x12*\n" +
	"\ageosite\x18\x03 \x01(\v2\x10.x.GeositeConfigR\ageosite\x12(\n" +
	"\x10use_bloom_filter\x18\x04 \x01(\bR\x0euseBloomFilter\x12\x1f\n" +
	"\vclash_files\x18\x05 \x03(\tR\n" +
	"clashFiles\"9\n" +
	"\x0fDomainSetConfig\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x12\n" +
	"\x04tags\x18\x02 \x03(\tR\x04tags\"\x81\x01\n" +
	"\x10GreatIPSetConfig\x12\x12\n" +
	"\x04name\x18\x03 \x01(\tR\x04name\x12#\n" +
	"\ropposite_name\x18\x04 \x01(\tR\foppositeName\x12\x19\n" +
	"\bex_names\x18\x05 \x03(\tR\aexNames\x12\x19\n" +
	"\bin_names\x18\x06 \x03(\tR\ainNames\"\xb2\x01\n" +
	"\x11AtomicIPSetConfig\x12\x12\n" +
	"\x04name\x18\x03 \x01(\tR\x04name\x12(\n" +
	"\x05cidrs\x18\x02 \x03(\v2\x12.x.common.geo.CIDRR\x05cidrs\x12$\n" +
	"\x05geoip\x18\x01 \x01(\v2\x0e.x.GeoIPConfigR\x05geoip\x12\x18\n" +
	"\ainverse\x18\x04 \x01(\bR\ainverse\x12\x1f\n" +
	"\vclash_files\x18\x05 \x03(\tR\n" +
	"clashFiles\"\xee\x02\n" +
	"\tGeoConfig\x123\n" +
	"\vdomain_sets\x18\x01 \x03(\v2\x12.x.DomainSetConfigR\n" +
	"domainSets\x12C\n" +
	"\x11great_domain_sets\x18\x03 \x03(\v2\x17.x.GreatDomainSetConfigR\x0fgreatDomainSets\x127\n" +
	"\rgreat_ip_sets\x18\x04 \x03(\v2\x13.x.GreatIPSetConfigR\vgreatIpSets\x12F\n" +
	"\x12atomic_domain_sets\x18\x05 \x03(\v2\x18.x.AtomicDomainSetConfigR\x10atomicDomainSets\x12:\n" +
	"\x0eatomic_ip_sets\x18\x06 \x03(\v2\x14.x.AtomicIPSetConfigR\fatomicIpSets\x12*\n" +
	"\bapp_sets\x18\a \x03(\v2\x0f.x.AppSetConfigR\aappSets\"a\n" +
	"\rGeositeConfig\x12\x14\n" +
	"\x05codes\x18\x01 \x03(\tR\x05codes\x12\x1e\n" +
	"\n" +
	"attributes\x18\x02 \x03(\tR\n" +
	"attributes\x12\x1a\n" +
	"\bfilepath\x18\x03 \x01(\tR\bfilepath\"?\n" +
	"\vGeoIPConfig\x12\x1a\n" +
	"\bfilepath\x18\x01 \x01(\tR\bfilepath\x12\x14\n" +
	"\x05codes\x18\x02 \x03(\tR\x05codes\"f\n" +
	"\fAppSetConfig\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12!\n" +
	"\aapp_ids\x18\x02 \x03(\v2\b.x.AppIdR\x06appIds\x12\x1f\n" +
	"\vclash_files\x18\x03 \x03(\tR\n" +
	"clashFilesB*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_geo_proto_rawDescOnce sync.Once
	file_protos_geo_proto_rawDescData []byte
)

func file_protos_geo_proto_rawDescGZIP() []byte {
	file_protos_geo_proto_rawDescOnce.Do(func() {
		file_protos_geo_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_geo_proto_rawDesc), len(file_protos_geo_proto_rawDesc)))
	})
	return file_protos_geo_proto_rawDescData
}

var file_protos_geo_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_protos_geo_proto_goTypes = []any{
	(*GreatDomainSetConfig)(nil),  // 0: x.GreatDomainSetConfig
	(*AtomicDomainSetConfig)(nil), // 1: x.AtomicDomainSetConfig
	(*DomainSetConfig)(nil),       // 2: x.DomainSetConfig
	(*GreatIPSetConfig)(nil),      // 3: x.GreatIPSetConfig
	(*AtomicIPSetConfig)(nil),     // 4: x.AtomicIPSetConfig
	(*GeoConfig)(nil),             // 5: x.GeoConfig
	(*GeositeConfig)(nil),         // 6: x.GeositeConfig
	(*GeoIPConfig)(nil),           // 7: x.GeoIPConfig
	(*AppSetConfig)(nil),          // 8: x.AppSetConfig
	(*geo.Domain)(nil),            // 9: x.common.geo.Domain
	(*geo.CIDR)(nil),              // 10: x.common.geo.CIDR
	(*AppId)(nil),                 // 11: x.AppId
}
var file_protos_geo_proto_depIdxs = []int32{
	9,  // 0: x.AtomicDomainSetConfig.domains:type_name -> x.common.geo.Domain
	6,  // 1: x.AtomicDomainSetConfig.geosite:type_name -> x.GeositeConfig
	10, // 2: x.AtomicIPSetConfig.cidrs:type_name -> x.common.geo.CIDR
	7,  // 3: x.AtomicIPSetConfig.geoip:type_name -> x.GeoIPConfig
	2,  // 4: x.GeoConfig.domain_sets:type_name -> x.DomainSetConfig
	0,  // 5: x.GeoConfig.great_domain_sets:type_name -> x.GreatDomainSetConfig
	3,  // 6: x.GeoConfig.great_ip_sets:type_name -> x.GreatIPSetConfig
	1,  // 7: x.GeoConfig.atomic_domain_sets:type_name -> x.AtomicDomainSetConfig
	4,  // 8: x.GeoConfig.atomic_ip_sets:type_name -> x.AtomicIPSetConfig
	8,  // 9: x.GeoConfig.app_sets:type_name -> x.AppSetConfig
	11, // 10: x.AppSetConfig.app_ids:type_name -> x.AppId
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_protos_geo_proto_init() }
func file_protos_geo_proto_init() {
	if File_protos_geo_proto != nil {
		return
	}
	file_protos_router_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_geo_proto_rawDesc), len(file_protos_geo_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_geo_proto_goTypes,
		DependencyIndexes: file_protos_geo_proto_depIdxs,
		MessageInfos:      file_protos_geo_proto_msgTypes,
	}.Build()
	File_protos_geo_proto = out.File
	file_protos_geo_proto_goTypes = nil
	file_protos_geo_proto_depIdxs = nil
}
