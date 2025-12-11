package proxy

import (
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

type DokodemoConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port          uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	Networks      []net.Network          `protobuf:"varint,7,rep,packed,name=networks,proto3,enum=x.common.net.Network" json:"networks,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DokodemoConfig) Reset() {
	*x = DokodemoConfig{}
	mi := &file_protos_proxy_dokodemo_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DokodemoConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DokodemoConfig) ProtoMessage() {}

func (x *DokodemoConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_dokodemo_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DokodemoConfig.ProtoReflect.Descriptor instead.
func (*DokodemoConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_dokodemo_proto_rawDescGZIP(), []int{0}
}

func (x *DokodemoConfig) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *DokodemoConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *DokodemoConfig) GetNetworks() []net.Network {
	if x != nil {
		return x.Networks
	}
	return nil
}

var File_protos_proxy_dokodemo_proto protoreflect.FileDescriptor

const file_protos_proxy_dokodemo_proto_rawDesc = "" +
	"\n" +
	"\x1bprotos/proxy/dokodemo.proto\x12\ax.proxy\x1a\x14common/net/net.proto\"q\n" +
	"\x0eDokodemoConfig\x12\x18\n" +
	"\aaddress\x18\x01 \x01(\tR\aaddress\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\x121\n" +
	"\bnetworks\x18\a \x03(\x0e2\x15.x.common.net.NetworkR\bnetworksB0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_dokodemo_proto_rawDescOnce sync.Once
	file_protos_proxy_dokodemo_proto_rawDescData []byte
)

func file_protos_proxy_dokodemo_proto_rawDescGZIP() []byte {
	file_protos_proxy_dokodemo_proto_rawDescOnce.Do(func() {
		file_protos_proxy_dokodemo_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_dokodemo_proto_rawDesc), len(file_protos_proxy_dokodemo_proto_rawDesc)))
	})
	return file_protos_proxy_dokodemo_proto_rawDescData
}

var file_protos_proxy_dokodemo_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_proxy_dokodemo_proto_goTypes = []any{
	(*DokodemoConfig)(nil), // 0: x.proxy.DokodemoConfig
	(net.Network)(0),       // 1: x.common.net.Network
}
var file_protos_proxy_dokodemo_proto_depIdxs = []int32{
	1, // 0: x.proxy.DokodemoConfig.networks:type_name -> x.common.net.Network
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protos_proxy_dokodemo_proto_init() }
func file_protos_proxy_dokodemo_proto_init() {
	if File_protos_proxy_dokodemo_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_dokodemo_proto_rawDesc), len(file_protos_proxy_dokodemo_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_dokodemo_proto_goTypes,
		DependencyIndexes: file_protos_proxy_dokodemo_proto_depIdxs,
		MessageInfos:      file_protos_proxy_dokodemo_proto_msgTypes,
	}.Build()
	File_protos_proxy_dokodemo_proto = out.File
	file_protos_proxy_dokodemo_proto_goTypes = nil
	file_protos_proxy_dokodemo_proto_depIdxs = nil
}
