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

type ForceALPN int32

const (
	// 优先使用用户在 TLS 设置中手动制定了 APLN 的值，否则使用传输协议的默认 ALPN 设置。
	ForceALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY ForceALPN = 0
	// 不发送 ALPN TLS 扩展
	ForceALPN_NO_ALPN ForceALPN = 1
	// 以 uTLS 的特征模板中的 ALPN 设置为准
	ForceALPN_UTLS_PRESET ForceALPN = 2
)

// Enum value maps for ForceALPN.
var (
	ForceALPN_name = map[int32]string{
		0: "TRANSPORT_PREFERENCE_TAKE_PRIORITY",
		1: "NO_ALPN",
		2: "UTLS_PRESET",
	}
	ForceALPN_value = map[string]int32{
		"TRANSPORT_PREFERENCE_TAKE_PRIORITY": 0,
		"NO_ALPN":                            1,
		"UTLS_PRESET":                        2,
	}
)

func (x ForceALPN) Enum() *ForceALPN {
	p := new(ForceALPN)
	*p = x
	return p
}

func (x ForceALPN) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ForceALPN) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_tls_tls_proto_enumTypes[0].Descriptor()
}

func (ForceALPN) Type() protoreflect.EnumType {
	return &file_protos_tls_tls_proto_enumTypes[0]
}

func (x ForceALPN) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ForceALPN.Descriptor instead.
func (ForceALPN) EnumDescriptor() ([]byte, []int) {
	return file_protos_tls_tls_proto_rawDescGZIP(), []int{0}
}

type TlsConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// certs to be provided to peer
	Certificates []*Certificate `protobuf:"bytes,1,rep,name=certificates,proto3" json:"certificates,omitempty"`
	// certs to be used as root CA
	RootCas [][]byte `protobuf:"bytes,2,rep,name=root_cas,json=rootCas,proto3" json:"root_cas,omitempty"`
	// certs to issue certificates which will be provided to peer
	IssueCas []*Certificate `protobuf:"bytes,3,rep,name=issue_cas,json=issueCas,proto3" json:"issue_cas,omitempty"`
	// if not specified, the server name will be destination domain of the connection if
	// it is a domain
	ServerName              string   `protobuf:"bytes,4,opt,name=server_name,json=serverName,proto3" json:"server_name,omitempty"`
	DisableSystemRoot       bool     `protobuf:"varint,5,opt,name=disable_system_root,json=disableSystemRoot,proto3" json:"disable_system_root,omitempty"`
	AllowInsecure           bool     `protobuf:"varint,6,opt,name=allow_insecure,json=allowInsecure,proto3" json:"allow_insecure,omitempty"`
	NextProtocol            []string `protobuf:"bytes,7,rep,name=next_protocol,json=nextProtocol,proto3" json:"next_protocol,omitempty"`
	EnableSessionResumption bool     `protobuf:"varint,8,opt,name=enable_session_resumption,json=enableSessionResumption,proto3" json:"enable_session_resumption,omitempty"`
	// A list of byte slice, each of which is a hash of a cert chain.
	PinnedPeerCertificateChainSha256 [][]byte `protobuf:"bytes,9,rep,name=pinned_peer_certificate_chain_sha256,json=pinnedPeerCertificateChainSha256,proto3" json:"pinned_peer_certificate_chain_sha256,omitempty"`
	VerifyClientCertificate          bool     `protobuf:"varint,10,opt,name=verify_client_certificate,json=verifyClientCertificate,proto3" json:"verify_client_certificate,omitempty"`
	// utls-related
	Imitate string `protobuf:"bytes,11,opt,name=imitate,proto3" json:"imitate,omitempty"`
	// utls-related
	NoSNI bool `protobuf:"varint,12,opt,name=noSNI,proto3" json:"noSNI,omitempty"`
	// utls-related
	ForceAlpn    ForceALPN `protobuf:"varint,13,opt,name=force_alpn,json=forceAlpn,proto3,enum=x.tls.ForceALPN" json:"force_alpn,omitempty"`
	MasterKeyLog string    `protobuf:"bytes,14,opt,name=master_key_log,json=masterKeyLog,proto3" json:"master_key_log,omitempty"`
	// server ech key
	EchKey []byte `protobuf:"bytes,15,opt,name=ech_key,json=echKey,proto3" json:"ech_key,omitempty"`
	// client ech config
	EchConfig []byte `protobuf:"bytes,16,opt,name=ech_config,json=echConfig,proto3" json:"ech_config,omitempty"`
	// client only
	// enable ech
	EnableEch     bool `protobuf:"varint,17,opt,name=enable_ech,json=enableEch,proto3" json:"enable_ech,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TlsConfig) Reset() {
	*x = TlsConfig{}
	mi := &file_protos_tls_tls_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TlsConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TlsConfig) ProtoMessage() {}

func (x *TlsConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protos_tls_tls_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TlsConfig.ProtoReflect.Descriptor instead.
func (*TlsConfig) Descriptor() ([]byte, []int) {
	return file_protos_tls_tls_proto_rawDescGZIP(), []int{0}
}

func (x *TlsConfig) GetCertificates() []*Certificate {
	if x != nil {
		return x.Certificates
	}
	return nil
}

func (x *TlsConfig) GetRootCas() [][]byte {
	if x != nil {
		return x.RootCas
	}
	return nil
}

func (x *TlsConfig) GetIssueCas() []*Certificate {
	if x != nil {
		return x.IssueCas
	}
	return nil
}

func (x *TlsConfig) GetServerName() string {
	if x != nil {
		return x.ServerName
	}
	return ""
}

func (x *TlsConfig) GetDisableSystemRoot() bool {
	if x != nil {
		return x.DisableSystemRoot
	}
	return false
}

func (x *TlsConfig) GetAllowInsecure() bool {
	if x != nil {
		return x.AllowInsecure
	}
	return false
}

func (x *TlsConfig) GetNextProtocol() []string {
	if x != nil {
		return x.NextProtocol
	}
	return nil
}

func (x *TlsConfig) GetEnableSessionResumption() bool {
	if x != nil {
		return x.EnableSessionResumption
	}
	return false
}

func (x *TlsConfig) GetPinnedPeerCertificateChainSha256() [][]byte {
	if x != nil {
		return x.PinnedPeerCertificateChainSha256
	}
	return nil
}

func (x *TlsConfig) GetVerifyClientCertificate() bool {
	if x != nil {
		return x.VerifyClientCertificate
	}
	return false
}

func (x *TlsConfig) GetImitate() string {
	if x != nil {
		return x.Imitate
	}
	return ""
}

func (x *TlsConfig) GetNoSNI() bool {
	if x != nil {
		return x.NoSNI
	}
	return false
}

func (x *TlsConfig) GetForceAlpn() ForceALPN {
	if x != nil {
		return x.ForceAlpn
	}
	return ForceALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY
}

func (x *TlsConfig) GetMasterKeyLog() string {
	if x != nil {
		return x.MasterKeyLog
	}
	return ""
}

func (x *TlsConfig) GetEchKey() []byte {
	if x != nil {
		return x.EchKey
	}
	return nil
}

func (x *TlsConfig) GetEchConfig() []byte {
	if x != nil {
		return x.EchConfig
	}
	return nil
}

func (x *TlsConfig) GetEnableEch() bool {
	if x != nil {
		return x.EnableEch
	}
	return false
}

var File_protos_tls_tls_proto protoreflect.FileDescriptor

const file_protos_tls_tls_proto_rawDesc = "" +
	"\n" +
	"\x14protos/tls/tls.proto\x12\x05x.tls\x1a\x1cprotos/tls/certificate.proto\"\xd2\x05\n" +
	"\tTlsConfig\x126\n" +
	"\fcertificates\x18\x01 \x03(\v2\x12.x.tls.CertificateR\fcertificates\x12\x19\n" +
	"\broot_cas\x18\x02 \x03(\fR\arootCas\x12/\n" +
	"\tissue_cas\x18\x03 \x03(\v2\x12.x.tls.CertificateR\bissueCas\x12\x1f\n" +
	"\vserver_name\x18\x04 \x01(\tR\n" +
	"serverName\x12.\n" +
	"\x13disable_system_root\x18\x05 \x01(\bR\x11disableSystemRoot\x12%\n" +
	"\x0eallow_insecure\x18\x06 \x01(\bR\rallowInsecure\x12#\n" +
	"\rnext_protocol\x18\a \x03(\tR\fnextProtocol\x12:\n" +
	"\x19enable_session_resumption\x18\b \x01(\bR\x17enableSessionResumption\x12N\n" +
	"$pinned_peer_certificate_chain_sha256\x18\t \x03(\fR pinnedPeerCertificateChainSha256\x12:\n" +
	"\x19verify_client_certificate\x18\n" +
	" \x01(\bR\x17verifyClientCertificate\x12\x18\n" +
	"\aimitate\x18\v \x01(\tR\aimitate\x12\x14\n" +
	"\x05noSNI\x18\f \x01(\bR\x05noSNI\x12/\n" +
	"\n" +
	"force_alpn\x18\r \x01(\x0e2\x10.x.tls.ForceALPNR\tforceAlpn\x12$\n" +
	"\x0emaster_key_log\x18\x0e \x01(\tR\fmasterKeyLog\x12\x17\n" +
	"\aech_key\x18\x0f \x01(\fR\x06echKey\x12\x1d\n" +
	"\n" +
	"ech_config\x18\x10 \x01(\fR\techConfig\x12\x1d\n" +
	"\n" +
	"enable_ech\x18\x11 \x01(\bR\tenableEch*Q\n" +
	"\tForceALPN\x12&\n" +
	"\"TRANSPORT_PREFERENCE_TAKE_PRIORITY\x10\x00\x12\v\n" +
	"\aNO_ALPN\x10\x01\x12\x0f\n" +
	"\vUTLS_PRESET\x10\x02B5Z3github.com/5vnetwork/vx-core/transport/security/tlsb\x06proto3"

var (
	file_protos_tls_tls_proto_rawDescOnce sync.Once
	file_protos_tls_tls_proto_rawDescData []byte
)

func file_protos_tls_tls_proto_rawDescGZIP() []byte {
	file_protos_tls_tls_proto_rawDescOnce.Do(func() {
		file_protos_tls_tls_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_tls_tls_proto_rawDesc), len(file_protos_tls_tls_proto_rawDesc)))
	})
	return file_protos_tls_tls_proto_rawDescData
}

var file_protos_tls_tls_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protos_tls_tls_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_tls_tls_proto_goTypes = []any{
	(ForceALPN)(0),      // 0: x.tls.ForceALPN
	(*TlsConfig)(nil),   // 1: x.tls.TlsConfig
	(*Certificate)(nil), // 2: x.tls.Certificate
}
var file_protos_tls_tls_proto_depIdxs = []int32{
	2, // 0: x.tls.TlsConfig.certificates:type_name -> x.tls.Certificate
	2, // 1: x.tls.TlsConfig.issue_cas:type_name -> x.tls.Certificate
	0, // 2: x.tls.TlsConfig.force_alpn:type_name -> x.tls.ForceALPN
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_protos_tls_tls_proto_init() }
func file_protos_tls_tls_proto_init() {
	if File_protos_tls_tls_proto != nil {
		return
	}
	file_protos_tls_certificate_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_tls_tls_proto_rawDesc), len(file_protos_tls_tls_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_tls_tls_proto_goTypes,
		DependencyIndexes: file_protos_tls_tls_proto_depIdxs,
		EnumInfos:         file_protos_tls_tls_proto_enumTypes,
		MessageInfos:      file_protos_tls_tls_proto_msgTypes,
	}.Build()
	File_protos_tls_tls_proto = out.File
	file_protos_tls_tls_proto_goTypes = nil
	file_protos_tls_tls_proto_depIdxs = nil
}
