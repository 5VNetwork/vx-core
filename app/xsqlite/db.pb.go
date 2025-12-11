package xsqlite

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

type Receipt struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Receipt) Reset() {
	*x = Receipt{}
	mi := &file_protos_db_db_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Receipt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Receipt) ProtoMessage() {}

func (x *Receipt) ProtoReflect() protoreflect.Message {
	mi := &file_protos_db_db_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Receipt.ProtoReflect.Descriptor instead.
func (*Receipt) Descriptor() ([]byte, []int) {
	return file_protos_db_db_proto_rawDescGZIP(), []int{0}
}

type GetHandlerRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetHandlerRequest) Reset() {
	*x = GetHandlerRequest{}
	mi := &file_protos_db_db_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetHandlerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetHandlerRequest) ProtoMessage() {}

func (x *GetHandlerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_db_db_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetHandlerRequest.ProtoReflect.Descriptor instead.
func (*GetHandlerRequest) Descriptor() ([]byte, []int) {
	return file_protos_db_db_proto_rawDescGZIP(), []int{1}
}

func (x *GetHandlerRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type GetAllHandlersRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetAllHandlersRequest) Reset() {
	*x = GetAllHandlersRequest{}
	mi := &file_protos_db_db_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetAllHandlersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAllHandlersRequest) ProtoMessage() {}

func (x *GetAllHandlersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_db_db_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAllHandlersRequest.ProtoReflect.Descriptor instead.
func (*GetAllHandlersRequest) Descriptor() ([]byte, []int) {
	return file_protos_db_db_proto_rawDescGZIP(), []int{2}
}

type GetHandlersByGroupRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Group         string                 `protobuf:"bytes,1,opt,name=group,proto3" json:"group,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetHandlersByGroupRequest) Reset() {
	*x = GetHandlersByGroupRequest{}
	mi := &file_protos_db_db_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetHandlersByGroupRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetHandlersByGroupRequest) ProtoMessage() {}

func (x *GetHandlersByGroupRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_db_db_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetHandlersByGroupRequest.ProtoReflect.Descriptor instead.
func (*GetHandlersByGroupRequest) Descriptor() ([]byte, []int) {
	return file_protos_db_db_proto_rawDescGZIP(), []int{3}
}

func (x *GetHandlersByGroupRequest) GetGroup() string {
	if x != nil {
		return x.Group
	}
	return ""
}

type GetBatchedHandlersRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	BatchSize     uint32                 `protobuf:"varint,1,opt,name=batch_size,json=batchSize,proto3" json:"batch_size,omitempty"`
	Offset        uint32                 `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetBatchedHandlersRequest) Reset() {
	*x = GetBatchedHandlersRequest{}
	mi := &file_protos_db_db_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetBatchedHandlersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBatchedHandlersRequest) ProtoMessage() {}

func (x *GetBatchedHandlersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_db_db_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBatchedHandlersRequest.ProtoReflect.Descriptor instead.
func (*GetBatchedHandlersRequest) Descriptor() ([]byte, []int) {
	return file_protos_db_db_proto_rawDescGZIP(), []int{4}
}

func (x *GetBatchedHandlersRequest) GetBatchSize() uint32 {
	if x != nil {
		return x.BatchSize
	}
	return 0
}

func (x *GetBatchedHandlersRequest) GetOffset() uint32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

type DbHandlers struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Handlers      []*DbOutboundHandler   `protobuf:"bytes,1,rep,name=handlers,proto3" json:"handlers,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DbHandlers) Reset() {
	*x = DbHandlers{}
	mi := &file_protos_db_db_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DbHandlers) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DbHandlers) ProtoMessage() {}

func (x *DbHandlers) ProtoReflect() protoreflect.Message {
	mi := &file_protos_db_db_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DbHandlers.ProtoReflect.Descriptor instead.
func (*DbHandlers) Descriptor() ([]byte, []int) {
	return file_protos_db_db_proto_rawDescGZIP(), []int{5}
}

func (x *DbHandlers) GetHandlers() []*DbOutboundHandler {
	if x != nil {
		return x.Handlers
	}
	return nil
}

type DbOutboundHandler struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Id               int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Tag              string                 `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	Ok               int32                  `protobuf:"varint,3,opt,name=ok,proto3" json:"ok,omitempty"`
	Speed            float32                `protobuf:"fixed32,4,opt,name=speed,proto3" json:"speed,omitempty"`
	SpeedTestTime    int32                  `protobuf:"varint,5,opt,name=speed_test_time,json=speedTestTime,proto3" json:"speed_test_time,omitempty"`
	Ping             int32                  `protobuf:"varint,6,opt,name=ping,proto3" json:"ping,omitempty"`
	PingTestTime     int32                  `protobuf:"varint,7,opt,name=ping_test_time,json=pingTestTime,proto3" json:"ping_test_time,omitempty"`
	Support6         int32                  `protobuf:"varint,8,opt,name=support6,proto3" json:"support6,omitempty"`
	Support6TestTime int32                  `protobuf:"varint,9,opt,name=support6_test_time,json=support6TestTime,proto3" json:"support6_test_time,omitempty"`
	Config           []byte                 `protobuf:"bytes,10,opt,name=config,proto3" json:"config,omitempty"`
	Selected         bool                   `protobuf:"varint,11,opt,name=selected,proto3" json:"selected,omitempty"`
	SubId            int64                  `protobuf:"varint,12,opt,name=sub_id,json=subId,proto3" json:"sub_id,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *DbOutboundHandler) Reset() {
	*x = DbOutboundHandler{}
	mi := &file_protos_db_db_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DbOutboundHandler) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DbOutboundHandler) ProtoMessage() {}

func (x *DbOutboundHandler) ProtoReflect() protoreflect.Message {
	mi := &file_protos_db_db_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DbOutboundHandler.ProtoReflect.Descriptor instead.
func (*DbOutboundHandler) Descriptor() ([]byte, []int) {
	return file_protos_db_db_proto_rawDescGZIP(), []int{6}
}

func (x *DbOutboundHandler) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *DbOutboundHandler) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *DbOutboundHandler) GetOk() int32 {
	if x != nil {
		return x.Ok
	}
	return 0
}

func (x *DbOutboundHandler) GetSpeed() float32 {
	if x != nil {
		return x.Speed
	}
	return 0
}

func (x *DbOutboundHandler) GetSpeedTestTime() int32 {
	if x != nil {
		return x.SpeedTestTime
	}
	return 0
}

func (x *DbOutboundHandler) GetPing() int32 {
	if x != nil {
		return x.Ping
	}
	return 0
}

func (x *DbOutboundHandler) GetPingTestTime() int32 {
	if x != nil {
		return x.PingTestTime
	}
	return 0
}

func (x *DbOutboundHandler) GetSupport6() int32 {
	if x != nil {
		return x.Support6
	}
	return 0
}

func (x *DbOutboundHandler) GetSupport6TestTime() int32 {
	if x != nil {
		return x.Support6TestTime
	}
	return 0
}

func (x *DbOutboundHandler) GetConfig() []byte {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *DbOutboundHandler) GetSelected() bool {
	if x != nil {
		return x.Selected
	}
	return false
}

func (x *DbOutboundHandler) GetSubId() int64 {
	if x != nil {
		return x.SubId
	}
	return 0
}

type UpdateHandlerRequest struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Id               int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Ok               *int32                 `protobuf:"varint,2,opt,name=ok,proto3,oneof" json:"ok,omitempty"`
	Speed            *float32               `protobuf:"fixed32,3,opt,name=speed,proto3,oneof" json:"speed,omitempty"`
	Ping             *int32                 `protobuf:"varint,4,opt,name=ping,proto3,oneof" json:"ping,omitempty"`
	Support6         *int32                 `protobuf:"varint,5,opt,name=support6,proto3,oneof" json:"support6,omitempty"`
	SpeedTestTime    *int32                 `protobuf:"varint,6,opt,name=speed_test_time,json=speedTestTime,proto3,oneof" json:"speed_test_time,omitempty"`
	PingTestTime     *int32                 `protobuf:"varint,7,opt,name=ping_test_time,json=pingTestTime,proto3,oneof" json:"ping_test_time,omitempty"`
	Support6TestTime *int32                 `protobuf:"varint,8,opt,name=support6_test_time,json=support6TestTime,proto3,oneof" json:"support6_test_time,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *UpdateHandlerRequest) Reset() {
	*x = UpdateHandlerRequest{}
	mi := &file_protos_db_db_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateHandlerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateHandlerRequest) ProtoMessage() {}

func (x *UpdateHandlerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_db_db_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateHandlerRequest.ProtoReflect.Descriptor instead.
func (*UpdateHandlerRequest) Descriptor() ([]byte, []int) {
	return file_protos_db_db_proto_rawDescGZIP(), []int{7}
}

func (x *UpdateHandlerRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *UpdateHandlerRequest) GetOk() int32 {
	if x != nil && x.Ok != nil {
		return *x.Ok
	}
	return 0
}

func (x *UpdateHandlerRequest) GetSpeed() float32 {
	if x != nil && x.Speed != nil {
		return *x.Speed
	}
	return 0
}

func (x *UpdateHandlerRequest) GetPing() int32 {
	if x != nil && x.Ping != nil {
		return *x.Ping
	}
	return 0
}

func (x *UpdateHandlerRequest) GetSupport6() int32 {
	if x != nil && x.Support6 != nil {
		return *x.Support6
	}
	return 0
}

func (x *UpdateHandlerRequest) GetSpeedTestTime() int32 {
	if x != nil && x.SpeedTestTime != nil {
		return *x.SpeedTestTime
	}
	return 0
}

func (x *UpdateHandlerRequest) GetPingTestTime() int32 {
	if x != nil && x.PingTestTime != nil {
		return *x.PingTestTime
	}
	return 0
}

func (x *UpdateHandlerRequest) GetSupport6TestTime() int32 {
	if x != nil && x.Support6TestTime != nil {
		return *x.Support6TestTime
	}
	return 0
}

var File_protos_db_db_proto protoreflect.FileDescriptor

const file_protos_db_db_proto_rawDesc = "" +
	"\n" +
	"\x12protos/db/db.proto\x12\x04x.db\"\t\n" +
	"\aReceipt\"#\n" +
	"\x11GetHandlerRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\"\x17\n" +
	"\x15GetAllHandlersRequest\"1\n" +
	"\x19GetHandlersByGroupRequest\x12\x14\n" +
	"\x05group\x18\x01 \x01(\tR\x05group\"R\n" +
	"\x19GetBatchedHandlersRequest\x12\x1d\n" +
	"\n" +
	"batch_size\x18\x01 \x01(\rR\tbatchSize\x12\x16\n" +
	"\x06offset\x18\x02 \x01(\rR\x06offset\"A\n" +
	"\n" +
	"DbHandlers\x123\n" +
	"\bhandlers\x18\x01 \x03(\v2\x17.x.db.DbOutboundHandlerR\bhandlers\"\xd2\x02\n" +
	"\x11DbOutboundHandler\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x10\n" +
	"\x03tag\x18\x02 \x01(\tR\x03tag\x12\x0e\n" +
	"\x02ok\x18\x03 \x01(\x05R\x02ok\x12\x14\n" +
	"\x05speed\x18\x04 \x01(\x02R\x05speed\x12&\n" +
	"\x0fspeed_test_time\x18\x05 \x01(\x05R\rspeedTestTime\x12\x12\n" +
	"\x04ping\x18\x06 \x01(\x05R\x04ping\x12$\n" +
	"\x0eping_test_time\x18\a \x01(\x05R\fpingTestTime\x12\x1a\n" +
	"\bsupport6\x18\b \x01(\x05R\bsupport6\x12,\n" +
	"\x12support6_test_time\x18\t \x01(\x05R\x10support6TestTime\x12\x16\n" +
	"\x06config\x18\n" +
	" \x01(\fR\x06config\x12\x1a\n" +
	"\bselected\x18\v \x01(\bR\bselected\x12\x15\n" +
	"\x06sub_id\x18\f \x01(\x03R\x05subId\"\x80\x03\n" +
	"\x14UpdateHandlerRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x13\n" +
	"\x02ok\x18\x02 \x01(\x05H\x00R\x02ok\x88\x01\x01\x12\x19\n" +
	"\x05speed\x18\x03 \x01(\x02H\x01R\x05speed\x88\x01\x01\x12\x17\n" +
	"\x04ping\x18\x04 \x01(\x05H\x02R\x04ping\x88\x01\x01\x12\x1f\n" +
	"\bsupport6\x18\x05 \x01(\x05H\x03R\bsupport6\x88\x01\x01\x12+\n" +
	"\x0fspeed_test_time\x18\x06 \x01(\x05H\x04R\rspeedTestTime\x88\x01\x01\x12)\n" +
	"\x0eping_test_time\x18\a \x01(\x05H\x05R\fpingTestTime\x88\x01\x01\x121\n" +
	"\x12support6_test_time\x18\b \x01(\x05H\x06R\x10support6TestTime\x88\x01\x01B\x05\n" +
	"\x03_okB\b\n" +
	"\x06_speedB\a\n" +
	"\x05_pingB\v\n" +
	"\t_support6B\x12\n" +
	"\x10_speed_test_timeB\x11\n" +
	"\x0f_ping_test_timeB\x15\n" +
	"\x13_support6_test_time2\xda\x02\n" +
	"\tDbService\x12>\n" +
	"\n" +
	"GetHandler\x12\x17.x.db.GetHandlerRequest\x1a\x17.x.db.DbOutboundHandler\x12?\n" +
	"\x0eGetAllHandlers\x12\x1b.x.db.GetAllHandlersRequest\x1a\x10.x.db.DbHandlers\x12G\n" +
	"\x12GetHandlersByGroup\x12\x1f.x.db.GetHandlersByGroupRequest\x1a\x10.x.db.DbHandlers\x12G\n" +
	"\x12GetBatchedHandlers\x12\x1f.x.db.GetBatchedHandlersRequest\x1a\x10.x.db.DbHandlers\x12:\n" +
	"\rUpdateHandler\x12\x1a.x.db.UpdateHandlerRequest\x1a\r.x.db.ReceiptB*Z(github.com/5vnetwork/vx-core/app/xsqliteb\x06proto3"

var (
	file_protos_db_db_proto_rawDescOnce sync.Once
	file_protos_db_db_proto_rawDescData []byte
)

func file_protos_db_db_proto_rawDescGZIP() []byte {
	file_protos_db_db_proto_rawDescOnce.Do(func() {
		file_protos_db_db_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_db_db_proto_rawDesc), len(file_protos_db_db_proto_rawDesc)))
	})
	return file_protos_db_db_proto_rawDescData
}

var file_protos_db_db_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_protos_db_db_proto_goTypes = []any{
	(*Receipt)(nil),                   // 0: x.db.Receipt
	(*GetHandlerRequest)(nil),         // 1: x.db.GetHandlerRequest
	(*GetAllHandlersRequest)(nil),     // 2: x.db.GetAllHandlersRequest
	(*GetHandlersByGroupRequest)(nil), // 3: x.db.GetHandlersByGroupRequest
	(*GetBatchedHandlersRequest)(nil), // 4: x.db.GetBatchedHandlersRequest
	(*DbHandlers)(nil),                // 5: x.db.DbHandlers
	(*DbOutboundHandler)(nil),         // 6: x.db.DbOutboundHandler
	(*UpdateHandlerRequest)(nil),      // 7: x.db.UpdateHandlerRequest
}
var file_protos_db_db_proto_depIdxs = []int32{
	6, // 0: x.db.DbHandlers.handlers:type_name -> x.db.DbOutboundHandler
	1, // 1: x.db.DbService.GetHandler:input_type -> x.db.GetHandlerRequest
	2, // 2: x.db.DbService.GetAllHandlers:input_type -> x.db.GetAllHandlersRequest
	3, // 3: x.db.DbService.GetHandlersByGroup:input_type -> x.db.GetHandlersByGroupRequest
	4, // 4: x.db.DbService.GetBatchedHandlers:input_type -> x.db.GetBatchedHandlersRequest
	7, // 5: x.db.DbService.UpdateHandler:input_type -> x.db.UpdateHandlerRequest
	6, // 6: x.db.DbService.GetHandler:output_type -> x.db.DbOutboundHandler
	5, // 7: x.db.DbService.GetAllHandlers:output_type -> x.db.DbHandlers
	5, // 8: x.db.DbService.GetHandlersByGroup:output_type -> x.db.DbHandlers
	5, // 9: x.db.DbService.GetBatchedHandlers:output_type -> x.db.DbHandlers
	0, // 10: x.db.DbService.UpdateHandler:output_type -> x.db.Receipt
	6, // [6:11] is the sub-list for method output_type
	1, // [1:6] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protos_db_db_proto_init() }
func file_protos_db_db_proto_init() {
	if File_protos_db_db_proto != nil {
		return
	}
	file_protos_db_db_proto_msgTypes[7].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_db_db_proto_rawDesc), len(file_protos_db_db_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_protos_db_db_proto_goTypes,
		DependencyIndexes: file_protos_db_db_proto_depIdxs,
		MessageInfos:      file_protos_db_db_proto_msgTypes,
	}.Build()
	File_protos_db_db_proto = out.File
	file_protos_db_db_proto_goTypes = nil
	file_protos_db_db_proto_depIdxs = nil
}
