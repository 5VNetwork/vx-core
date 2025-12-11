package log

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

type LogLevel int32

const (
	LogLevel_DEBUG   LogLevel = 0
	LogLevel_INFO    LogLevel = 1
	LogLevel_WARNING LogLevel = 2
	LogLevel_ERROR   LogLevel = 3
	LogLevel_SEVERE  LogLevel = 4
	LogLevel_NONE    LogLevel = 5
)

// Enum value maps for LogLevel.
var (
	LogLevel_name = map[int32]string{
		0: "DEBUG",
		1: "INFO",
		2: "WARNING",
		3: "ERROR",
		4: "SEVERE",
		5: "NONE",
	}
	LogLevel_value = map[string]int32{
		"DEBUG":   0,
		"INFO":    1,
		"WARNING": 2,
		"ERROR":   3,
		"SEVERE":  4,
		"NONE":    5,
	}
)

func (x LogLevel) Enum() *LogLevel {
	p := new(LogLevel)
	*p = x
	return p
}

func (x LogLevel) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LogLevel) Descriptor() protoreflect.EnumDescriptor {
	return file_common_log_log_proto_enumTypes[0].Descriptor()
}

func (LogLevel) Type() protoreflect.EnumType {
	return &file_common_log_log_proto_enumTypes[0]
}

func (x LogLevel) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LogLevel.Descriptor instead.
func (LogLevel) EnumDescriptor() ([]byte, []int) {
	return file_common_log_log_proto_rawDescGZIP(), []int{0}
}

var File_common_log_log_proto protoreflect.FileDescriptor

const file_common_log_log_proto_rawDesc = "" +
	"\n" +
	"\x14common/log/log.proto\x12\fx.common.log*M\n" +
	"\bLogLevel\x12\t\n" +
	"\x05DEBUG\x10\x00\x12\b\n" +
	"\x04INFO\x10\x01\x12\v\n" +
	"\aWARNING\x10\x02\x12\t\n" +
	"\x05ERROR\x10\x03\x12\n" +
	"\n" +
	"\x06SEVERE\x10\x04\x12\b\n" +
	"\x04NONE\x10\x05B)Z'github.com/5vnetwork/vx-core/common/logb\x06proto3"

var (
	file_common_log_log_proto_rawDescOnce sync.Once
	file_common_log_log_proto_rawDescData []byte
)

func file_common_log_log_proto_rawDescGZIP() []byte {
	file_common_log_log_proto_rawDescOnce.Do(func() {
		file_common_log_log_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_log_log_proto_rawDesc), len(file_common_log_log_proto_rawDesc)))
	})
	return file_common_log_log_proto_rawDescData
}

var file_common_log_log_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_common_log_log_proto_goTypes = []any{
	(LogLevel)(0), // 0: x.common.log.LogLevel
}
var file_common_log_log_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_log_log_proto_init() }
func file_common_log_log_proto_init() {
	if File_common_log_log_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_log_log_proto_rawDesc), len(file_common_log_log_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_log_log_proto_goTypes,
		DependencyIndexes: file_common_log_log_proto_depIdxs,
		EnumInfos:         file_common_log_log_proto_enumTypes,
	}.Build()
	File_common_log_log_proto = out.File
	file_common_log_log_proto_goTypes = nil
	file_common_log_log_proto_depIdxs = nil
}
