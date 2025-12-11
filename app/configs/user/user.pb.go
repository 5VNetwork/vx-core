package user

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type UserConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	UserLevel uint32 `protobuf:"varint,2,opt,name=user_level,json=userLevel,proto3" json:"user_level,omitempty"`
	// uuid is string format
	Secret string `protobuf:"bytes,3,opt,name=secret,proto3" json:"secret,omitempty"`
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
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Users []*UserConfig `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
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

var file_protos_user_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x01, 0x78, 0x22, 0x53, 0x0a, 0x0a, 0x55, 0x73, 0x65, 0x72, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x6c, 0x65, 0x76,
	0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x75, 0x73, 0x65, 0x72, 0x4c, 0x65,
	0x76, 0x65, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x22, 0x38, 0x0a, 0x11, 0x55,
	0x73, 0x65, 0x72, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x23, 0x0a, 0x05, 0x75, 0x73, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x0d, 0x2e, 0x78, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x05,
	0x75, 0x73, 0x65, 0x72, 0x73, 0x42, 0x11, 0x5a, 0x0f, 0x74, 0x6d, 0x2f, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x73, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protos_user_proto_rawDescOnce sync.Once
	file_protos_user_proto_rawDescData = file_protos_user_proto_rawDesc
)

func file_protos_user_proto_rawDescGZIP() []byte {
	file_protos_user_proto_rawDescOnce.Do(func() {
		file_protos_user_proto_rawDescData = protoimpl.X.CompressGZIP(file_protos_user_proto_rawDescData)
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
			RawDescriptor: file_protos_user_proto_rawDesc,
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
	file_protos_user_proto_rawDesc = nil
	file_protos_user_proto_goTypes = nil
	file_protos_user_proto_depIdxs = nil
}
