package proxy

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

// Config for HTTP proxy server.
type HttpServerConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HttpServerConfig) Reset() {
	*x = HttpServerConfig{}
	mi := &file_protos_proxy_http_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HttpServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HttpServerConfig) ProtoMessage() {}

func (x *HttpServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_http_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HttpServerConfig.ProtoReflect.Descriptor instead.
func (*HttpServerConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_http_proto_rawDescGZIP(), []int{0}
}

// ClientConfig is the protobuf config for HTTP proxy client.
type HttpClientConfig struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	H1SkipWaitForReply bool                   `protobuf:"varint,2,opt,name=h1_skip_wait_for_reply,json=h1SkipWaitForReply,proto3" json:"h1_skip_wait_for_reply,omitempty"`
	Account            *Account               `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *HttpClientConfig) Reset() {
	*x = HttpClientConfig{}
	mi := &file_protos_proxy_http_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HttpClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HttpClientConfig) ProtoMessage() {}

func (x *HttpClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_http_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HttpClientConfig.ProtoReflect.Descriptor instead.
func (*HttpClientConfig) Descriptor() ([]byte, []int) {
	return file_protos_proxy_http_proto_rawDescGZIP(), []int{1}
}

func (x *HttpClientConfig) GetH1SkipWaitForReply() bool {
	if x != nil {
		return x.H1SkipWaitForReply
	}
	return false
}

func (x *HttpClientConfig) GetAccount() *Account {
	if x != nil {
		return x.Account
	}
	return nil
}

type Account struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Username      string                 `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password      string                 `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Account) Reset() {
	*x = Account{}
	mi := &file_protos_proxy_http_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Account) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Account) ProtoMessage() {}

func (x *Account) ProtoReflect() protoreflect.Message {
	mi := &file_protos_proxy_http_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Account.ProtoReflect.Descriptor instead.
func (*Account) Descriptor() ([]byte, []int) {
	return file_protos_proxy_http_proto_rawDescGZIP(), []int{2}
}

func (x *Account) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *Account) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

var File_protos_proxy_http_proto protoreflect.FileDescriptor

const file_protos_proxy_http_proto_rawDesc = "" +
	"\n" +
	"\x17protos/proxy/http.proto\x12\ax.proxy\"\x12\n" +
	"\x10HttpServerConfig\"r\n" +
	"\x10HttpClientConfig\x122\n" +
	"\x16h1_skip_wait_for_reply\x18\x02 \x01(\bR\x12h1SkipWaitForReply\x12*\n" +
	"\aaccount\x18\x01 \x01(\v2\x10.x.proxy.AccountR\aaccount\"A\n" +
	"\aAccount\x12\x1a\n" +
	"\busername\x18\x01 \x01(\tR\busername\x12\x1a\n" +
	"\bpassword\x18\x02 \x01(\tR\bpasswordB0Z.github.com/5vnetwork/vx-core/app/configs/proxyb\x06proto3"

var (
	file_protos_proxy_http_proto_rawDescOnce sync.Once
	file_protos_proxy_http_proto_rawDescData []byte
)

func file_protos_proxy_http_proto_rawDescGZIP() []byte {
	file_protos_proxy_http_proto_rawDescOnce.Do(func() {
		file_protos_proxy_http_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_proxy_http_proto_rawDesc), len(file_protos_proxy_http_proto_rawDesc)))
	})
	return file_protos_proxy_http_proto_rawDescData
}

var file_protos_proxy_http_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_protos_proxy_http_proto_goTypes = []any{
	(*HttpServerConfig)(nil), // 0: x.proxy.HttpServerConfig
	(*HttpClientConfig)(nil), // 1: x.proxy.HttpClientConfig
	(*Account)(nil),          // 2: x.proxy.Account
}
var file_protos_proxy_http_proto_depIdxs = []int32{
	2, // 0: x.proxy.HttpClientConfig.account:type_name -> x.proxy.Account
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protos_proxy_http_proto_init() }
func file_protos_proxy_http_proto_init() {
	if File_protos_proxy_http_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_proxy_http_proto_rawDesc), len(file_protos_proxy_http_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_proxy_http_proto_goTypes,
		DependencyIndexes: file_protos_proxy_http_proto_depIdxs,
		MessageInfos:      file_protos_proxy_http_proto_msgTypes,
	}.Build()
	File_protos_proxy_http_proto = out.File
	file_protos_proxy_http_proto_goTypes = nil
	file_protos_proxy_http_proto_depIdxs = nil
}
