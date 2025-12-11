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

type Level int32

const (
	Level_DEBUG    Level = 0
	Level_INFO     Level = 1
	Level_WARN     Level = 2
	Level_ERROR    Level = 3
	Level_FATAL    Level = 4
	Level_DISABLED Level = 5
)

// Enum value maps for Level.
var (
	Level_name = map[int32]string{
		0: "DEBUG",
		1: "INFO",
		2: "WARN",
		3: "ERROR",
		4: "FATAL",
		5: "DISABLED",
	}
	Level_value = map[string]int32{
		"DEBUG":    0,
		"INFO":     1,
		"WARN":     2,
		"ERROR":    3,
		"FATAL":    4,
		"DISABLED": 5,
	}
)

func (x Level) Enum() *Level {
	p := new(Level)
	*p = x
	return p
}

func (x Level) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Level) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_logger_proto_enumTypes[0].Descriptor()
}

func (Level) Type() protoreflect.EnumType {
	return &file_protos_logger_proto_enumTypes[0]
}

func (x Level) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Level.Descriptor instead.
func (Level) EnumDescriptor() ([]byte, []int) {
	return file_protos_logger_proto_rawDescGZIP(), []int{0}
}

type LoggerConfig struct {
	state    protoimpl.MessageState `protogen:"open.v1"`
	LogLevel Level                  `protobuf:"varint,1,opt,name=log_level,json=logLevel,proto3,enum=x.Level" json:"log_level,omitempty"`
	FilePath string                 `protobuf:"bytes,2,opt,name=file_path,json=filePath,proto3" json:"file_path,omitempty"`
	// whether use console writer for stderr output
	ConsoleWriter bool   `protobuf:"varint,3,opt,name=console_writer,json=consoleWriter,proto3" json:"console_writer,omitempty"`
	ShowColor     bool   `protobuf:"varint,4,opt,name=show_color,json=showColor,proto3" json:"show_color,omitempty"`
	ShowCaller    bool   `protobuf:"varint,5,opt,name=show_caller,json=showCaller,proto3" json:"show_caller,omitempty"`
	LogFileDir    string `protobuf:"bytes,9,opt,name=log_file_dir,json=logFileDir,proto3" json:"log_file_dir,omitempty"`
	Redact        bool   `protobuf:"varint,10,opt,name=redact,proto3" json:"redact,omitempty"`
	UserLog       bool   `protobuf:"varint,6,opt,name=user_log,json=userLog,proto3" json:"user_log,omitempty"`
	LogAppId      bool   `protobuf:"varint,7,opt,name=log_app_id,json=logAppId,proto3" json:"log_app_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LoggerConfig) Reset() {
	*x = LoggerConfig{}
	mi := &file_protos_logger_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LoggerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LoggerConfig) ProtoMessage() {}

func (x *LoggerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_logger_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LoggerConfig.ProtoReflect.Descriptor instead.
func (*LoggerConfig) Descriptor() ([]byte, []int) {
	return file_protos_logger_proto_rawDescGZIP(), []int{0}
}

func (x *LoggerConfig) GetLogLevel() Level {
	if x != nil {
		return x.LogLevel
	}
	return Level_DEBUG
}

func (x *LoggerConfig) GetFilePath() string {
	if x != nil {
		return x.FilePath
	}
	return ""
}

func (x *LoggerConfig) GetConsoleWriter() bool {
	if x != nil {
		return x.ConsoleWriter
	}
	return false
}

func (x *LoggerConfig) GetShowColor() bool {
	if x != nil {
		return x.ShowColor
	}
	return false
}

func (x *LoggerConfig) GetShowCaller() bool {
	if x != nil {
		return x.ShowCaller
	}
	return false
}

func (x *LoggerConfig) GetLogFileDir() string {
	if x != nil {
		return x.LogFileDir
	}
	return ""
}

func (x *LoggerConfig) GetRedact() bool {
	if x != nil {
		return x.Redact
	}
	return false
}

func (x *LoggerConfig) GetUserLog() bool {
	if x != nil {
		return x.UserLog
	}
	return false
}

func (x *LoggerConfig) GetLogAppId() bool {
	if x != nil {
		return x.LogAppId
	}
	return false
}

var File_protos_logger_proto protoreflect.FileDescriptor

const file_protos_logger_proto_rawDesc = "" +
	"\n" +
	"\x13protos/logger.proto\x12\x01x\"\xac\x02\n" +
	"\fLoggerConfig\x12%\n" +
	"\tlog_level\x18\x01 \x01(\x0e2\b.x.LevelR\blogLevel\x12\x1b\n" +
	"\tfile_path\x18\x02 \x01(\tR\bfilePath\x12%\n" +
	"\x0econsole_writer\x18\x03 \x01(\bR\rconsoleWriter\x12\x1d\n" +
	"\n" +
	"show_color\x18\x04 \x01(\bR\tshowColor\x12\x1f\n" +
	"\vshow_caller\x18\x05 \x01(\bR\n" +
	"showCaller\x12 \n" +
	"\flog_file_dir\x18\t \x01(\tR\n" +
	"logFileDir\x12\x16\n" +
	"\x06redact\x18\n" +
	" \x01(\bR\x06redact\x12\x19\n" +
	"\buser_log\x18\x06 \x01(\bR\auserLog\x12\x1c\n" +
	"\n" +
	"log_app_id\x18\a \x01(\bR\blogAppId*J\n" +
	"\x05Level\x12\t\n" +
	"\x05DEBUG\x10\x00\x12\b\n" +
	"\x04INFO\x10\x01\x12\b\n" +
	"\x04WARN\x10\x02\x12\t\n" +
	"\x05ERROR\x10\x03\x12\t\n" +
	"\x05FATAL\x10\x04\x12\f\n" +
	"\bDISABLED\x10\x05B*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_logger_proto_rawDescOnce sync.Once
	file_protos_logger_proto_rawDescData []byte
)

func file_protos_logger_proto_rawDescGZIP() []byte {
	file_protos_logger_proto_rawDescOnce.Do(func() {
		file_protos_logger_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_logger_proto_rawDesc), len(file_protos_logger_proto_rawDesc)))
	})
	return file_protos_logger_proto_rawDescData
}

var file_protos_logger_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protos_logger_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_logger_proto_goTypes = []any{
	(Level)(0),           // 0: x.Level
	(*LoggerConfig)(nil), // 1: x.LoggerConfig
}
var file_protos_logger_proto_depIdxs = []int32{
	0, // 0: x.LoggerConfig.log_level:type_name -> x.Level
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protos_logger_proto_init() }
func file_protos_logger_proto_init() {
	if File_protos_logger_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_logger_proto_rawDesc), len(file_protos_logger_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_logger_proto_goTypes,
		DependencyIndexes: file_protos_logger_proto_depIdxs,
		EnumInfos:         file_protos_logger_proto_enumTypes,
		MessageInfos:      file_protos_logger_proto_msgTypes,
	}.Build()
	File_protos_logger_proto = out.File
	file_protos_logger_proto_goTypes = nil
	file_protos_logger_proto_depIdxs = nil
}
