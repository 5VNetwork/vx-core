package geo

import (
	reflect "reflect"
	sync "sync"

	geo "github.com/5vnetwork/vx-core/common/geo"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GreatDomainSetConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name         string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	OppositeName string   `protobuf:"bytes,4,opt,name=opposite_name,json=oppositeName,proto3" json:"opposite_name,omitempty"`
	ExNames      []string `protobuf:"bytes,5,rep,name=ex_names,json=exNames,proto3" json:"ex_names,omitempty"`
	InNames      []string `protobuf:"bytes,6,rep,name=in_names,json=inNames,proto3" json:"in_names,omitempty"`
}

func (x *GreatDomainSetConfig) Reset() {
	*x = GreatDomainSetConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_geo_geo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GreatDomainSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GreatDomainSetConfig) ProtoMessage() {}

func (x *GreatDomainSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_internal_geo_geo_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_internal_geo_geo_proto_rawDescGZIP(), []int{0}
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
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// either one of the following two fields should be set
	Domains []*geo.Domain  `protobuf:"bytes,2,rep,name=domains,proto3" json:"domains,omitempty"`
	Geosite *GeositeConfig `protobuf:"bytes,3,opt,name=geosite,proto3" json:"geosite,omitempty"`
}

func (x *AtomicDomainSetConfig) Reset() {
	*x = AtomicDomainSetConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_geo_geo_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AtomicDomainSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AtomicDomainSetConfig) ProtoMessage() {}

func (x *AtomicDomainSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_internal_geo_geo_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_internal_geo_geo_proto_rawDescGZIP(), []int{1}
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

type GreatIPSetConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name         string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	OppositeName string   `protobuf:"bytes,4,opt,name=opposite_name,json=oppositeName,proto3" json:"opposite_name,omitempty"`
	ExNames      []string `protobuf:"bytes,5,rep,name=ex_names,json=exNames,proto3" json:"ex_names,omitempty"`
	InNames      []string `protobuf:"bytes,6,rep,name=in_names,json=inNames,proto3" json:"in_names,omitempty"`
}

func (x *GreatIPSetConfig) Reset() {
	*x = GreatIPSetConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_geo_geo_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GreatIPSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GreatIPSetConfig) ProtoMessage() {}

func (x *GreatIPSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_internal_geo_geo_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_internal_geo_geo_proto_rawDescGZIP(), []int{2}
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
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string       `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Cidrs []*geo.CIDR  `protobuf:"bytes,2,rep,name=cidrs,proto3" json:"cidrs,omitempty"`
	Geoip *GeoIPConfig `protobuf:"bytes,1,opt,name=geoip,proto3" json:"geoip,omitempty"`
}

func (x *AtomicIPSetConfig) Reset() {
	*x = AtomicIPSetConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_geo_geo_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AtomicIPSetConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AtomicIPSetConfig) ProtoMessage() {}

func (x *AtomicIPSetConfig) ProtoReflect() protoreflect.Message {
	mi := &file_internal_geo_geo_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_internal_geo_geo_proto_rawDescGZIP(), []int{3}
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

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GreatDomainSets  []*GreatDomainSetConfig  `protobuf:"bytes,3,rep,name=great_domain_sets,json=greatDomainSets,proto3" json:"great_domain_sets,omitempty"`
	GreatIpSets      []*GreatIPSetConfig      `protobuf:"bytes,4,rep,name=great_ip_sets,json=greatIpSets,proto3" json:"great_ip_sets,omitempty"`
	AtomicDomainSets []*AtomicDomainSetConfig `protobuf:"bytes,5,rep,name=atomic_domain_sets,json=atomicDomainSets,proto3" json:"atomic_domain_sets,omitempty"`
	AtomicIpSets     []*AtomicIPSetConfig     `protobuf:"bytes,6,rep,name=atomic_ip_sets,json=atomicIpSets,proto3" json:"atomic_ip_sets,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_geo_geo_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_internal_geo_geo_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_internal_geo_geo_proto_rawDescGZIP(), []int{4}
}

func (x *Config) GetGreatDomainSets() []*GreatDomainSetConfig {
	if x != nil {
		return x.GreatDomainSets
	}
	return nil
}

func (x *Config) GetGreatIpSets() []*GreatIPSetConfig {
	if x != nil {
		return x.GreatIpSets
	}
	return nil
}

func (x *Config) GetAtomicDomainSets() []*AtomicDomainSetConfig {
	if x != nil {
		return x.AtomicDomainSets
	}
	return nil
}

func (x *Config) GetAtomicIpSets() []*AtomicIPSetConfig {
	if x != nil {
		return x.AtomicIpSets
	}
	return nil
}

type GeositeConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Codes      []string `protobuf:"bytes,1,rep,name=codes,proto3" json:"codes,omitempty"`
	Attributes []string `protobuf:"bytes,2,rep,name=attributes,proto3" json:"attributes,omitempty"`
	Filepath   string   `protobuf:"bytes,3,opt,name=filepath,proto3" json:"filepath,omitempty"`
}

func (x *GeositeConfig) Reset() {
	*x = GeositeConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_geo_geo_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeositeConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeositeConfig) ProtoMessage() {}

func (x *GeositeConfig) ProtoReflect() protoreflect.Message {
	mi := &file_internal_geo_geo_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_internal_geo_geo_proto_rawDescGZIP(), []int{5}
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
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Filepath string   `protobuf:"bytes,1,opt,name=filepath,proto3" json:"filepath,omitempty"`
	Codes    []string `protobuf:"bytes,2,rep,name=codes,proto3" json:"codes,omitempty"`
	Inverse  bool     `protobuf:"varint,3,opt,name=inverse,proto3" json:"inverse,omitempty"`
}

func (x *GeoIPConfig) Reset() {
	*x = GeoIPConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_geo_geo_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeoIPConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoIPConfig) ProtoMessage() {}

func (x *GeoIPConfig) ProtoReflect() protoreflect.Message {
	mi := &file_internal_geo_geo_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_internal_geo_geo_proto_rawDescGZIP(), []int{6}
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

func (x *GeoIPConfig) GetInverse() bool {
	if x != nil {
		return x.Inverse
	}
	return false
}

var File_internal_geo_geo_proto protoreflect.FileDescriptor

var file_internal_geo_geo_proto_rawDesc = []byte{
	0x0a, 0x16, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6f, 0x2f, 0x67,
	0x65, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x74, 0x6d, 0x2e, 0x67, 0x65, 0x6f,
	0x1a, 0x1d, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2f, 0x67, 0x65, 0x6f, 0x2f, 0x67, 0x65, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x85, 0x01, 0x0a, 0x14, 0x47, 0x72, 0x65, 0x61, 0x74, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x53,
	0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x23, 0x0a, 0x0d,
	0x6f, 0x70, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0c, 0x6f, 0x70, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x65, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x78, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x05, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x07, 0x65, 0x78, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x19, 0x0a, 0x08,
	0x69, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07,
	0x69, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x8d, 0x01, 0x0a, 0x15, 0x41, 0x74, 0x6f, 0x6d,
	0x69, 0x63, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x2f, 0x0a, 0x07, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x74, 0x6d, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x67, 0x65, 0x6f, 0x2e, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x52, 0x07, 0x64,
	0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x12, 0x2f, 0x0a, 0x07, 0x67, 0x65, 0x6f, 0x73, 0x69, 0x74,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x74, 0x6d, 0x2e, 0x67, 0x65, 0x6f,
	0x2e, 0x47, 0x65, 0x6f, 0x73, 0x69, 0x74, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x07,
	0x67, 0x65, 0x6f, 0x73, 0x69, 0x74, 0x65, 0x22, 0x81, 0x01, 0x0a, 0x10, 0x47, 0x72, 0x65, 0x61,
	0x74, 0x49, 0x50, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x23, 0x0a, 0x0d, 0x6f, 0x70, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x65, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6f, 0x70, 0x70, 0x6f, 0x73, 0x69, 0x74,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x78, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x65, 0x78, 0x4e, 0x61, 0x6d, 0x65, 0x73,
	0x12, 0x19, 0x0a, 0x08, 0x69, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x06, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x07, 0x69, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x7d, 0x0a, 0x11, 0x41,
	0x74, 0x6f, 0x6d, 0x69, 0x63, 0x49, 0x50, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x29, 0x0a, 0x05, 0x63, 0x69, 0x64, 0x72, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x74, 0x6d, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e,
	0x67, 0x65, 0x6f, 0x2e, 0x43, 0x49, 0x44, 0x52, 0x52, 0x05, 0x63, 0x69, 0x64, 0x72, 0x73, 0x12,
	0x29, 0x0a, 0x05, 0x67, 0x65, 0x6f, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13,
	0x2e, 0x74, 0x6d, 0x2e, 0x67, 0x65, 0x6f, 0x2e, 0x47, 0x65, 0x6f, 0x49, 0x50, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x05, 0x67, 0x65, 0x6f, 0x69, 0x70, 0x22, 0x9e, 0x02, 0x0a, 0x06, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x48, 0x0a, 0x11, 0x67, 0x72, 0x65, 0x61, 0x74, 0x5f, 0x64,
	0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x5f, 0x73, 0x65, 0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x74, 0x6d, 0x2e, 0x67, 0x65, 0x6f, 0x2e, 0x47, 0x72, 0x65, 0x61, 0x74, 0x44,
	0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0f,
	0x67, 0x72, 0x65, 0x61, 0x74, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x53, 0x65, 0x74, 0x73, 0x12,
	0x3c, 0x0a, 0x0d, 0x67, 0x72, 0x65, 0x61, 0x74, 0x5f, 0x69, 0x70, 0x5f, 0x73, 0x65, 0x74, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x74, 0x6d, 0x2e, 0x67, 0x65, 0x6f, 0x2e,
	0x47, 0x72, 0x65, 0x61, 0x74, 0x49, 0x50, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x52, 0x0b, 0x67, 0x72, 0x65, 0x61, 0x74, 0x49, 0x70, 0x53, 0x65, 0x74, 0x73, 0x12, 0x4b, 0x0a,
	0x12, 0x61, 0x74, 0x6f, 0x6d, 0x69, 0x63, 0x5f, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x5f, 0x73,
	0x65, 0x74, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x74, 0x6d, 0x2e, 0x67,
	0x65, 0x6f, 0x2e, 0x41, 0x74, 0x6f, 0x6d, 0x69, 0x63, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x53,
	0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x10, 0x61, 0x74, 0x6f, 0x6d, 0x69, 0x63,
	0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x53, 0x65, 0x74, 0x73, 0x12, 0x3f, 0x0a, 0x0e, 0x61, 0x74,
	0x6f, 0x6d, 0x69, 0x63, 0x5f, 0x69, 0x70, 0x5f, 0x73, 0x65, 0x74, 0x73, 0x18, 0x06, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x19, 0x2e, 0x74, 0x6d, 0x2e, 0x67, 0x65, 0x6f, 0x2e, 0x41, 0x74, 0x6f, 0x6d,
	0x69, 0x63, 0x49, 0x50, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0c, 0x61,
	0x74, 0x6f, 0x6d, 0x69, 0x63, 0x49, 0x70, 0x53, 0x65, 0x74, 0x73, 0x22, 0x61, 0x0a, 0x0d, 0x47,
	0x65, 0x6f, 0x73, 0x69, 0x74, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x14, 0x0a, 0x05,
	0x63, 0x6f, 0x64, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x63, 0x6f, 0x64,
	0x65, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x70, 0x61, 0x74, 0x68, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x70, 0x61, 0x74, 0x68, 0x22, 0x59,
	0x0a, 0x0b, 0x47, 0x65, 0x6f, 0x49, 0x50, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1a, 0x0a,
	0x08, 0x66, 0x69, 0x6c, 0x65, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x66, 0x69, 0x6c, 0x65, 0x70, 0x61, 0x74, 0x68, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x64,
	0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x63, 0x6f, 0x64, 0x65, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x69, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x69, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x65, 0x42, 0x11, 0x5a, 0x0f, 0x74, 0x6d, 0x2f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6f, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_geo_geo_proto_rawDescOnce sync.Once
	file_internal_geo_geo_proto_rawDescData = file_internal_geo_geo_proto_rawDesc
)

func file_internal_geo_geo_proto_rawDescGZIP() []byte {
	file_internal_geo_geo_proto_rawDescOnce.Do(func() {
		file_internal_geo_geo_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_geo_geo_proto_rawDescData)
	})
	return file_internal_geo_geo_proto_rawDescData
}

var file_internal_geo_geo_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_internal_geo_geo_proto_goTypes = []any{
	(*GreatDomainSetConfig)(nil),  // 0: x.geo.GreatDomainSetConfig
	(*AtomicDomainSetConfig)(nil), // 1: x.geo.AtomicDomainSetConfig
	(*GreatIPSetConfig)(nil),      // 2: x.geo.GreatIPSetConfig
	(*AtomicIPSetConfig)(nil),     // 3: x.geo.AtomicIPSetConfig
	(*Config)(nil),                // 4: x.geo.Config
	(*GeositeConfig)(nil),         // 5: x.geo.GeositeConfig
	(*GeoIPConfig)(nil),           // 6: x.geo.GeoIPConfig
	(*geo.Domain)(nil),            // 7: x.common.geo.Domain
	(*geo.CIDR)(nil),              // 8: x.common.geo.CIDR
}
var file_internal_geo_geo_proto_depIdxs = []int32{
	7, // 0: x.geo.AtomicDomainSetConfig.domains:type_name -> x.common.geo.Domain
	5, // 1: x.geo.AtomicDomainSetConfig.geosite:type_name -> x.geo.GeositeConfig
	8, // 2: x.geo.AtomicIPSetConfig.cidrs:type_name -> x.common.geo.CIDR
	6, // 3: x.geo.AtomicIPSetConfig.geoip:type_name -> x.geo.GeoIPConfig
	0, // 4: x.geo.Config.great_domain_sets:type_name -> x.geo.GreatDomainSetConfig
	2, // 5: x.geo.Config.great_ip_sets:type_name -> x.geo.GreatIPSetConfig
	1, // 6: x.geo.Config.atomic_domain_sets:type_name -> x.geo.AtomicDomainSetConfig
	3, // 7: x.geo.Config.atomic_ip_sets:type_name -> x.geo.AtomicIPSetConfig
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_internal_geo_geo_proto_init() }
func file_internal_geo_geo_proto_init() {
	if File_internal_geo_geo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_geo_geo_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*GreatDomainSetConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_geo_geo_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*AtomicDomainSetConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_geo_geo_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*GreatIPSetConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_geo_geo_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*AtomicIPSetConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_geo_geo_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*Config); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_geo_geo_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*GeositeConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_geo_geo_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*GeoIPConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_geo_geo_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_geo_geo_proto_goTypes,
		DependencyIndexes: file_internal_geo_geo_proto_depIdxs,
		MessageInfos:      file_internal_geo_geo_proto_msgTypes,
	}.Build()
	File_internal_geo_geo_proto = out.File
	file_internal_geo_geo_proto_rawDesc = nil
	file_internal_geo_geo_proto_goTypes = nil
	file_internal_geo_geo_proto_depIdxs = nil
}
