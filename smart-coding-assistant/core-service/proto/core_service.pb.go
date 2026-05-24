// 由 protoc-gen-go 自动生成。请勿手动编辑。
// 版本:
// 	protoc-gen-go v1.36.11
// 	protoc        v7.34.1
// 源文件: protos/core_service.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// 验证此生成代码是否足够新。
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// 验证 runtime/protoimpl 是否足够新。
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ChatRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Message       string                 `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Context       map[string]string      `protobuf:"bytes,3,rep,name=context,proto3" json:"context,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChatRequest) Reset() {
	*x = ChatRequest{}
	mi := &file_protos_core_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatRequest) ProtoMessage() {}

func (x *ChatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_core_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 ChatRequest.ProtoReflect.Descriptor。
func (*ChatRequest) Descriptor() ([]byte, []int) {
	return file_protos_core_service_proto_rawDescGZIP(), []int{0}
}

func (x *ChatRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *ChatRequest) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *ChatRequest) GetContext() map[string]string {
	if x != nil {
		return x.Context
	}
	return nil
}

type ChatResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Response      string                 `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	IsFinished    bool                   `protobuf:"varint,2,opt,name=is_finished,json=isFinished,proto3" json:"is_finished,omitempty"`
	Context       map[string]string      `protobuf:"bytes,3,rep,name=context,proto3" json:"context,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChatResponse) Reset() {
	*x = ChatResponse{}
	mi := &file_protos_core_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatResponse) ProtoMessage() {}

func (x *ChatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_core_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 ChatResponse.ProtoReflect.Descriptor。
func (*ChatResponse) Descriptor() ([]byte, []int) {
	return file_protos_core_service_proto_rawDescGZIP(), []int{1}
}

func (x *ChatResponse) GetResponse() string {
	if x != nil {
		return x.Response
	}
	return ""
}

func (x *ChatResponse) GetIsFinished() bool {
	if x != nil {
		return x.IsFinished
	}
	return false
}

func (x *ChatResponse) GetContext() map[string]string {
	if x != nil {
		return x.Context
	}
	return nil
}

type GetHistoryRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Limit         int32                  `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
	Offset        int32                  `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetHistoryRequest) Reset() {
	*x = GetHistoryRequest{}
	mi := &file_protos_core_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetHistoryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetHistoryRequest) ProtoMessage() {}

func (x *GetHistoryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_core_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 GetHistoryRequest.ProtoReflect.Descriptor。
func (*GetHistoryRequest) Descriptor() ([]byte, []int) {
	return file_protos_core_service_proto_rawDescGZIP(), []int{2}
}

func (x *GetHistoryRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *GetHistoryRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *GetHistoryRequest) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

type GetHistoryResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Messages      []*ChatMessage         `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetHistoryResponse) Reset() {
	*x = GetHistoryResponse{}
	mi := &file_protos_core_service_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetHistoryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetHistoryResponse) ProtoMessage() {}

func (x *GetHistoryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_core_service_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 GetHistoryResponse.ProtoReflect.Descriptor。
func (*GetHistoryResponse) Descriptor() ([]byte, []int) {
	return file_protos_core_service_proto_rawDescGZIP(), []int{3}
}

func (x *GetHistoryResponse) GetMessages() []*ChatMessage {
	if x != nil {
		return x.Messages
	}
	return nil
}

type ChatMessage struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	UserId        string                 `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Role          string                 `protobuf:"bytes,3,opt,name=role,proto3" json:"role,omitempty"` // 用户或助手
	Content       string                 `protobuf:"bytes,4,opt,name=content,proto3" json:"content,omitempty"`
	Timestamp     int64                  `protobuf:"varint,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChatMessage) Reset() {
	*x = ChatMessage{}
	mi := &file_protos_core_service_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatMessage) ProtoMessage() {}

func (x *ChatMessage) ProtoReflect() protoreflect.Message {
	mi := &file_protos_core_service_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 ChatMessage.ProtoReflect.Descriptor。
func (*ChatMessage) Descriptor() ([]byte, []int) {
	return file_protos_core_service_proto_rawDescGZIP(), []int{4}
}

func (x *ChatMessage) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ChatMessage) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *ChatMessage) GetRole() string {
	if x != nil {
		return x.Role
	}
	return ""
}

func (x *ChatMessage) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *ChatMessage) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

var File_protos_core_service_proto protoreflect.FileDescriptor

const file_protos_core_service_proto_rawDesc = "" +
	"\n" +
	"\x19protos/core_service.proto\x12\x04core\"\xb6\x01\n" +
	"\vChatRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12\x18\n" +
	"\amessage\x18\x02 \x01(\tR\amessage\x128\n" +
	"\acontext\x18\x03 \x03(\v2\x1e.core.ChatRequest.ContextEntryR\acontext\x1a:\n" +
	"\fContextEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\xc2\x01\n" +
	"\fChatResponse\x12\x1a\n" +
	"\bresponse\x18\x01 \x01(\tR\bresponse\x12\x1f\n" +
	"\vis_finished\x18\x02 \x01(\bR\n" +
	"isFinished\x129\n" +
	"\acontext\x18\x03 \x03(\v2\x1f.core.ChatResponse.ContextEntryR\acontext\x1a:\n" +
	"\fContextEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"Z\n" +
	"\x11GetHistoryRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12\x14\n" +
	"\x05limit\x18\x02 \x01(\x05R\x05limit\x12\x16\n" +
	"\x06offset\x18\x03 \x01(\x05R\x06offset\"C\n" +
	"\x12GetHistoryResponse\x12-\n" +
	"\bmessages\x18\x01 \x03(\v2\x11.core.ChatMessageR\bmessages\"\x82\x01\n" +
	"\vChatMessage\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x17\n" +
	"\auser_id\x18\x02 \x01(\tR\x06userId\x12\x12\n" +
	"\x04role\x18\x03 \x01(\tR\x04role\x12\x18\n" +
	"\acontent\x18\x04 \x01(\tR\acontent\x12\x1c\n" +
	"\ttimestamp\x18\x05 \x01(\x03R\ttimestamp2}\n" +
	"\vCoreService\x12-\n" +
	"\x04Chat\x12\x11.core.ChatRequest\x1a\x12.core.ChatResponse\x12?\n" +
	"\n" +
	"GetHistory\x12\x17.core.GetHistoryRequest\x1a\x18.core.GetHistoryResponseB\tZ\a./protob\x06proto3"

var (
	file_protos_core_service_proto_rawDescOnce sync.Once
	file_protos_core_service_proto_rawDescData []byte
)

func file_protos_core_service_proto_rawDescGZIP() []byte {
	file_protos_core_service_proto_rawDescOnce.Do(func() {
		file_protos_core_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_core_service_proto_rawDesc), len(file_protos_core_service_proto_rawDesc)))
	})
	return file_protos_core_service_proto_rawDescData
}

var file_protos_core_service_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_protos_core_service_proto_goTypes = []any{
	(*ChatRequest)(nil),        // 0: core.ChatRequest
	(*ChatResponse)(nil),       // 1: core.ChatResponse
	(*GetHistoryRequest)(nil),  // 2: core.GetHistoryRequest
	(*GetHistoryResponse)(nil), // 3: core.GetHistoryResponse
	(*ChatMessage)(nil),        // 4: core.ChatMessage
	nil,                        // 5: core.ChatRequest.ContextEntry
	nil,                        // 6: core.ChatResponse.ContextEntry
}
var file_protos_core_service_proto_depIdxs = []int32{
	5, // 0: core.ChatRequest.context:type_name -> core.ChatRequest.ContextEntry
	6, // 1: core.ChatResponse.context:type_name -> core.ChatResponse.ContextEntry
	4, // 2: core.GetHistoryResponse.messages:type_name -> core.ChatMessage
	0, // 3: core.CoreService.Chat:input_type -> core.ChatRequest
	2, // 4: core.CoreService.GetHistory:input_type -> core.GetHistoryRequest
	1, // 5: core.CoreService.Chat:output_type -> core.ChatResponse
	3, // 6: core.CoreService.GetHistory:output_type -> core.GetHistoryResponse
	5, // [5:7] 是方法输出类型的子列表
	3, // [3:5] 是方法输入类型的子列表
	3, // [3:3] 是扩展类型名称的子列表
	3, // [3:3] 是扩展被扩展者的子列表
	0, // [0:3] 是字段类型名称的子列表
}

func init() { file_protos_core_service_proto_init() }
func file_protos_core_service_proto_init() {
	if File_protos_core_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_core_service_proto_rawDesc), len(file_protos_core_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_protos_core_service_proto_goTypes,
		DependencyIndexes: file_protos_core_service_proto_depIdxs,
		MessageInfos:      file_protos_core_service_proto_msgTypes,
	}.Build()
	File_protos_core_service_proto = out.File
	file_protos_core_service_proto_goTypes = nil
	file_protos_core_service_proto_depIdxs = nil
}
