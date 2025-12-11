package tls

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

type Certificate struct {
	state               protoimpl.MessageState `protogen:"open.v1"`
	Certificate         []byte                 `protobuf:"bytes,1,opt,name=certificate,proto3" json:"certificate,omitempty"`
	Key                 []byte                 `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	CertificateFilepath string                 `protobuf:"bytes,4,opt,name=certificate_filepath,json=certificateFilepath,proto3" json:"certificate_filepath,omitempty"`
	KeyFilepath         string                 `protobuf:"bytes,5,opt,name=key_filepath,json=keyFilepath,proto3" json:"key_filepath,omitempty"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *Certificate) Reset() {
	*x = Certificate{}
	mi := &file_protos_tls_certificate_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Certificate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Certificate) ProtoMessage() {}

func (x *Certificate) ProtoReflect() protoreflect.Message {
	mi := &file_protos_tls_certificate_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Certificate.ProtoReflect.Descriptor instead.
func (*Certificate) Descriptor() ([]byte, []int) {
	return file_protos_tls_certificate_proto_rawDescGZIP(), []int{0}
}

func (x *Certificate) GetCertificate() []byte {
	if x != nil {
		return x.Certificate
	}
	return nil
}

func (x *Certificate) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *Certificate) GetCertificateFilepath() string {
	if x != nil {
		return x.CertificateFilepath
	}
	return ""
}

func (x *Certificate) GetKeyFilepath() string {
	if x != nil {
		return x.KeyFilepath
	}
	return ""
}

var File_protos_tls_certificate_proto protoreflect.FileDescriptor

const file_protos_tls_certificate_proto_rawDesc = "" +
	"\n" +
	"\x1cprotos/tls/certificate.proto\x12\x05x.tls\"\x97\x01\n" +
	"\vCertificate\x12 \n" +
	"\vcertificate\x18\x01 \x01(\fR\vcertificate\x12\x10\n" +
	"\x03key\x18\x02 \x01(\fR\x03key\x121\n" +
	"\x14certificate_filepath\x18\x04 \x01(\tR\x13certificateFilepath\x12!\n" +
	"\fkey_filepath\x18\x05 \x01(\tR\vkeyFilepathB5Z3github.com/5vnetwork/vx-core/transport/security/tlsb\x06proto3"

var (
	file_protos_tls_certificate_proto_rawDescOnce sync.Once
	file_protos_tls_certificate_proto_rawDescData []byte
)

func file_protos_tls_certificate_proto_rawDescGZIP() []byte {
	file_protos_tls_certificate_proto_rawDescOnce.Do(func() {
		file_protos_tls_certificate_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_tls_certificate_proto_rawDesc), len(file_protos_tls_certificate_proto_rawDesc)))
	})
	return file_protos_tls_certificate_proto_rawDescData
}

var file_protos_tls_certificate_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_tls_certificate_proto_goTypes = []any{
	(*Certificate)(nil), // 0: x.tls.Certificate
}
var file_protos_tls_certificate_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protos_tls_certificate_proto_init() }
func file_protos_tls_certificate_proto_init() {
	if File_protos_tls_certificate_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_tls_certificate_proto_rawDesc), len(file_protos_tls_certificate_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_tls_certificate_proto_goTypes,
		DependencyIndexes: file_protos_tls_certificate_proto_depIdxs,
		MessageInfos:      file_protos_tls_certificate_proto_msgTypes,
	}.Build()
	File_protos_tls_certificate_proto = out.File
	file_protos_tls_certificate_proto_goTypes = nil
	file_protos_tls_certificate_proto_depIdxs = nil
}
