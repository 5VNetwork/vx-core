package reality

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

type RealityConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Show  bool                   `protobuf:"varint,1,opt,name=show,proto3" json:"show,omitempty"`
	// server only
	Dest string `protobuf:"bytes,2,opt,name=dest,proto3" json:"dest,omitempty"`
	// string type = 3;
	// server only
	Xver uint64 `protobuf:"varint,4,opt,name=xver,proto3" json:"xver,omitempty"`
	// server only
	ServerNames []string `protobuf:"bytes,5,rep,name=server_names,json=serverNames,proto3" json:"server_names,omitempty"`
	// server only
	PrivateKey []byte `protobuf:"bytes,6,opt,name=private_key,json=privateKey,proto3" json:"private_key,omitempty"`
	// server only
	MinClientVer []byte `protobuf:"bytes,7,opt,name=min_client_ver,json=minClientVer,proto3" json:"min_client_ver,omitempty"`
	// server only
	MaxClientVer []byte `protobuf:"bytes,8,opt,name=max_client_ver,json=maxClientVer,proto3" json:"max_client_ver,omitempty"`
	// server only
	// miliseconds
	MaxTimeDiff uint64 `protobuf:"varint,9,opt,name=max_time_diff,json=maxTimeDiff,proto3" json:"max_time_diff,omitempty"`
	// server only
	ShortIds    [][]byte `protobuf:"bytes,10,rep,name=short_ids,json=shortIds,proto3" json:"short_ids,omitempty"`
	Fingerprint string   `protobuf:"bytes,21,opt,name=Fingerprint,proto3" json:"Fingerprint,omitempty"`
	ServerName  string   `protobuf:"bytes,22,opt,name=server_name,json=serverName,proto3" json:"server_name,omitempty"`
	Pbk         string   `protobuf:"bytes,28,opt,name=pbk,proto3" json:"pbk,omitempty"`
	// must be 8 bytes
	Sid           string  `protobuf:"bytes,29,opt,name=sid,proto3" json:"sid,omitempty"`
	SpiderX       string  `protobuf:"bytes,25,opt,name=spider_x,json=spiderX,proto3" json:"spider_x,omitempty"`
	MasterKeyLog  string  `protobuf:"bytes,27,opt,name=master_key_log,json=masterKeyLog,proto3" json:"master_key_log,omitempty"`
	ShortId       []byte  `protobuf:"bytes,24,opt,name=short_id,json=shortId,proto3" json:"short_id,omitempty"`
	SpiderY       []int64 `protobuf:"varint,26,rep,packed,name=spider_y,json=spiderY,proto3" json:"spider_y,omitempty"`
	PublicKey     []byte  `protobuf:"bytes,23,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RealityConfig) Reset() {
	*x = RealityConfig{}
	mi := &file_transport_security_reality_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RealityConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RealityConfig) ProtoMessage() {}

func (x *RealityConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_security_reality_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RealityConfig.ProtoReflect.Descriptor instead.
func (*RealityConfig) Descriptor() ([]byte, []int) {
	return file_transport_security_reality_config_proto_rawDescGZIP(), []int{0}
}

func (x *RealityConfig) GetShow() bool {
	if x != nil {
		return x.Show
	}
	return false
}

func (x *RealityConfig) GetDest() string {
	if x != nil {
		return x.Dest
	}
	return ""
}

func (x *RealityConfig) GetXver() uint64 {
	if x != nil {
		return x.Xver
	}
	return 0
}

func (x *RealityConfig) GetServerNames() []string {
	if x != nil {
		return x.ServerNames
	}
	return nil
}

func (x *RealityConfig) GetPrivateKey() []byte {
	if x != nil {
		return x.PrivateKey
	}
	return nil
}

func (x *RealityConfig) GetMinClientVer() []byte {
	if x != nil {
		return x.MinClientVer
	}
	return nil
}

func (x *RealityConfig) GetMaxClientVer() []byte {
	if x != nil {
		return x.MaxClientVer
	}
	return nil
}

func (x *RealityConfig) GetMaxTimeDiff() uint64 {
	if x != nil {
		return x.MaxTimeDiff
	}
	return 0
}

func (x *RealityConfig) GetShortIds() [][]byte {
	if x != nil {
		return x.ShortIds
	}
	return nil
}

func (x *RealityConfig) GetFingerprint() string {
	if x != nil {
		return x.Fingerprint
	}
	return ""
}

func (x *RealityConfig) GetServerName() string {
	if x != nil {
		return x.ServerName
	}
	return ""
}

func (x *RealityConfig) GetPbk() string {
	if x != nil {
		return x.Pbk
	}
	return ""
}

func (x *RealityConfig) GetSid() string {
	if x != nil {
		return x.Sid
	}
	return ""
}

func (x *RealityConfig) GetSpiderX() string {
	if x != nil {
		return x.SpiderX
	}
	return ""
}

func (x *RealityConfig) GetMasterKeyLog() string {
	if x != nil {
		return x.MasterKeyLog
	}
	return ""
}

func (x *RealityConfig) GetShortId() []byte {
	if x != nil {
		return x.ShortId
	}
	return nil
}

func (x *RealityConfig) GetSpiderY() []int64 {
	if x != nil {
		return x.SpiderY
	}
	return nil
}

func (x *RealityConfig) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

var File_transport_security_reality_config_proto protoreflect.FileDescriptor

const file_transport_security_reality_config_proto_rawDesc = "" +
	"\n" +
	"'transport/security/reality/config.proto\x12\x1cx.transport.security.reality\"\x99\x04\n" +
	"\rRealityConfig\x12\x12\n" +
	"\x04show\x18\x01 \x01(\bR\x04show\x12\x12\n" +
	"\x04dest\x18\x02 \x01(\tR\x04dest\x12\x12\n" +
	"\x04xver\x18\x04 \x01(\x04R\x04xver\x12!\n" +
	"\fserver_names\x18\x05 \x03(\tR\vserverNames\x12\x1f\n" +
	"\vprivate_key\x18\x06 \x01(\fR\n" +
	"privateKey\x12$\n" +
	"\x0emin_client_ver\x18\a \x01(\fR\fminClientVer\x12$\n" +
	"\x0emax_client_ver\x18\b \x01(\fR\fmaxClientVer\x12\"\n" +
	"\rmax_time_diff\x18\t \x01(\x04R\vmaxTimeDiff\x12\x1b\n" +
	"\tshort_ids\x18\n" +
	" \x03(\fR\bshortIds\x12 \n" +
	"\vFingerprint\x18\x15 \x01(\tR\vFingerprint\x12\x1f\n" +
	"\vserver_name\x18\x16 \x01(\tR\n" +
	"serverName\x12\x10\n" +
	"\x03pbk\x18\x1c \x01(\tR\x03pbk\x12\x10\n" +
	"\x03sid\x18\x1d \x01(\tR\x03sid\x12\x19\n" +
	"\bspider_x\x18\x19 \x01(\tR\aspiderX\x12$\n" +
	"\x0emaster_key_log\x18\x1b \x01(\tR\fmasterKeyLog\x12\x19\n" +
	"\bshort_id\x18\x18 \x01(\fR\ashortId\x12\x19\n" +
	"\bspider_y\x18\x1a \x03(\x03R\aspiderY\x12\x1d\n" +
	"\n" +
	"public_key\x18\x17 \x01(\fR\tpublicKeyB9Z7github.com/5vnetwork/vx-core/transport/security/realityb\x06proto3"

var (
	file_transport_security_reality_config_proto_rawDescOnce sync.Once
	file_transport_security_reality_config_proto_rawDescData []byte
)

func file_transport_security_reality_config_proto_rawDescGZIP() []byte {
	file_transport_security_reality_config_proto_rawDescOnce.Do(func() {
		file_transport_security_reality_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_security_reality_config_proto_rawDesc), len(file_transport_security_reality_config_proto_rawDesc)))
	})
	return file_transport_security_reality_config_proto_rawDescData
}

var file_transport_security_reality_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_security_reality_config_proto_goTypes = []any{
	(*RealityConfig)(nil), // 0: x.transport.security.reality.RealityConfig
}
var file_transport_security_reality_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_transport_security_reality_config_proto_init() }
func file_transport_security_reality_config_proto_init() {
	if File_transport_security_reality_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_security_reality_config_proto_rawDesc), len(file_transport_security_reality_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_security_reality_config_proto_goTypes,
		DependencyIndexes: file_transport_security_reality_config_proto_depIdxs,
		MessageInfos:      file_transport_security_reality_config_proto_msgTypes,
	}.Build()
	File_transport_security_reality_config_proto = out.File
	file_transport_security_reality_config_proto_goTypes = nil
	file_transport_security_reality_config_proto_depIdxs = nil
}
