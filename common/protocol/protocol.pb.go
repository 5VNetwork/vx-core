package protocol

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

type SecurityType int32

const (
	SecurityType_UNKNOWN           SecurityType = 0
	SecurityType_LEGACY            SecurityType = 1
	SecurityType_AUTO              SecurityType = 2
	SecurityType_AES128_GCM        SecurityType = 3
	SecurityType_CHACHA20_POLY1305 SecurityType = 4
	SecurityType_NONE              SecurityType = 5
	SecurityType_ZERO              SecurityType = 6
)

// Enum value maps for SecurityType.
var (
	SecurityType_name = map[int32]string{
		0: "UNKNOWN",
		1: "LEGACY",
		2: "AUTO",
		3: "AES128_GCM",
		4: "CHACHA20_POLY1305",
		5: "NONE",
		6: "ZERO",
	}
	SecurityType_value = map[string]int32{
		"UNKNOWN":           0,
		"LEGACY":            1,
		"AUTO":              2,
		"AES128_GCM":        3,
		"CHACHA20_POLY1305": 4,
		"NONE":              5,
		"ZERO":              6,
	}
)

func (x SecurityType) Enum() *SecurityType {
	p := new(SecurityType)
	*p = x
	return p
}

func (x SecurityType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SecurityType) Descriptor() protoreflect.EnumDescriptor {
	return file_common_protocol_protocol_proto_enumTypes[0].Descriptor()
}

func (SecurityType) Type() protoreflect.EnumType {
	return &file_common_protocol_protocol_proto_enumTypes[0]
}

func (x SecurityType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SecurityType.Descriptor instead.
func (SecurityType) EnumDescriptor() ([]byte, []int) {
	return file_common_protocol_protocol_proto_rawDescGZIP(), []int{0}
}

var File_common_protocol_protocol_proto protoreflect.FileDescriptor

const file_common_protocol_protocol_proto_rawDesc = "" +
	"\n" +
	"\x1ecommon/protocol/protocol.proto\x12\x11x.common.protocol*l\n" +
	"\fSecurityType\x12\v\n" +
	"\aUNKNOWN\x10\x00\x12\n" +
	"\n" +
	"\x06LEGACY\x10\x01\x12\b\n" +
	"\x04AUTO\x10\x02\x12\x0e\n" +
	"\n" +
	"AES128_GCM\x10\x03\x12\x15\n" +
	"\x11CHACHA20_POLY1305\x10\x04\x12\b\n" +
	"\x04NONE\x10\x05\x12\b\n" +
	"\x04ZERO\x10\x06B.Z,github.com/5vnetwork/vx-core/common/protocolb\x06proto3"

var (
	file_common_protocol_protocol_proto_rawDescOnce sync.Once
	file_common_protocol_protocol_proto_rawDescData []byte
)

func file_common_protocol_protocol_proto_rawDescGZIP() []byte {
	file_common_protocol_protocol_proto_rawDescOnce.Do(func() {
		file_common_protocol_protocol_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_protocol_protocol_proto_rawDesc), len(file_common_protocol_protocol_proto_rawDesc)))
	})
	return file_common_protocol_protocol_proto_rawDescData
}

var file_common_protocol_protocol_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_common_protocol_protocol_proto_goTypes = []any{
	(SecurityType)(0), // 0: x.common.protocol.SecurityType
}
var file_common_protocol_protocol_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_protocol_protocol_proto_init() }
func file_common_protocol_protocol_proto_init() {
	if File_common_protocol_protocol_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_protocol_protocol_proto_rawDesc), len(file_common_protocol_protocol_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_protocol_protocol_proto_goTypes,
		DependencyIndexes: file_common_protocol_protocol_proto_depIdxs,
		EnumInfos:         file_common_protocol_protocol_proto_enumTypes,
	}.Build()
	File_common_protocol_protocol_proto = out.File
	file_common_protocol_protocol_proto_goTypes = nil
	file_common_protocol_protocol_proto_depIdxs = nil
}
