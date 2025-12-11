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

type UserConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	UserLevel     uint32                 `protobuf:"varint,2,opt,name=user_level,json=userLevel,proto3" json:"user_level,omitempty"`
	Secret        string                 `protobuf:"bytes,3,opt,name=secret,proto3" json:"secret,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UserConfig) Reset() {
	*x = UserConfig{}
	mi := &file_protos_user_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserConfig) ProtoMessage() {}

func (x *UserConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_user_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserConfig.ProtoReflect.Descriptor instead.
func (*UserConfig) Descriptor() ([]byte, []int) {
	return file_protos_user_proto_rawDescGZIP(), []int{0}
}

func (x *UserConfig) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *UserConfig) GetUserLevel() uint32 {
	if x != nil {
		return x.UserLevel
	}
	return 0
}

func (x *UserConfig) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

type UserManagerConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Users         []*UserConfig          `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UserManagerConfig) Reset() {
	*x = UserManagerConfig{}
	mi := &file_protos_user_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserManagerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserManagerConfig) ProtoMessage() {}

func (x *UserManagerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_user_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserManagerConfig.ProtoReflect.Descriptor instead.
func (*UserManagerConfig) Descriptor() ([]byte, []int) {
	return file_protos_user_proto_rawDescGZIP(), []int{1}
}

func (x *UserManagerConfig) GetUsers() []*UserConfig {
	if x != nil {
		return x.Users
	}
	return nil
}

var File_protos_user_proto protoreflect.FileDescriptor

const file_protos_user_proto_rawDesc = "" +
	"\n" +
	"\x11protos/user.proto\x12\x01x\"S\n" +
	"\n" +
	"UserConfig\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x1d\n" +
	"\n" +
	"user_level\x18\x02 \x01(\rR\tuserLevel\x12\x16\n" +
	"\x06secret\x18\x03 \x01(\tR\x06secret\"8\n" +
	"\x11UserManagerConfig\x12#\n" +
	"\x05users\x18\x01 \x03(\v2\r.x.UserConfigR\x05usersB*Z(github.com/5vnetwork/vx-core/app/configsb\x06proto3"

var (
	file_protos_user_proto_rawDescOnce sync.Once
	file_protos_user_proto_rawDescData []byte
)

func file_protos_user_proto_rawDescGZIP() []byte {
	file_protos_user_proto_rawDescOnce.Do(func() {
		file_protos_user_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_user_proto_rawDesc), len(file_protos_user_proto_rawDesc)))
	})
	return file_protos_user_proto_rawDescData
}

var file_protos_user_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protos_user_proto_goTypes = []any{
	(*UserConfig)(nil),        // 0: x.UserConfig
	(*UserManagerConfig)(nil), // 1: x.UserManagerConfig
}
var file_protos_user_proto_depIdxs = []int32{
	0, // 0: x.UserManagerConfig.users:type_name -> x.UserConfig
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protos_user_proto_init() }
func file_protos_user_proto_init() {
	if File_protos_user_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_user_proto_rawDesc), len(file_protos_user_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_user_proto_goTypes,
		DependencyIndexes: file_protos_user_proto_depIdxs,
		MessageInfos:      file_protos_user_proto_msgTypes,
	}.Build()
	File_protos_user_proto = out.File
	file_protos_user_proto_goTypes = nil
	file_protos_user_proto_depIdxs = nil
}
