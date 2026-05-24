// 由 protoc-gen-go 自动生成。请勿手动编辑。
// 版本：
// 	protoc-gen-go v1.36.11
// 	protoc        v7.34.1
// 源文件：protos/memory_service.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// 验证此生成的代码是否足够新。
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// 验证 runtime/protoimpl 是否足够新。
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SaveContextRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Context       map[string]string      `protobuf:"bytes,2,rep,name=context,proto3" json:"context,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Ttl           int64                  `protobuf:"varint,3,opt,name=ttl,proto3" json:"ttl,omitempty"` // 生存时间（秒）
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SaveContextRequest) Reset() {
	*x = SaveContextRequest{}
	mi := &file_protos_memory_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SaveContextRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SaveContextRequest) ProtoMessage() {}

func (x *SaveContextRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 SaveContextRequest.ProtoReflect.Descriptor。
func (*SaveContextRequest) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{0}
}

func (x *SaveContextRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *SaveContextRequest) GetContext() map[string]string {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *SaveContextRequest) GetTtl() int64 {
	if x != nil {
		return x.Ttl
	}
	return 0
}

type SaveContextResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SaveContextResponse) Reset() {
	*x = SaveContextResponse{}
	mi := &file_protos_memory_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SaveContextResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SaveContextResponse) ProtoMessage() {}

func (x *SaveContextResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 SaveContextResponse.ProtoReflect.Descriptor。
func (*SaveContextResponse) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{1}
}

func (x *SaveContextResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *SaveContextResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type GetContextRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Keys          []string               `protobuf:"bytes,2,rep,name=keys,proto3" json:"keys,omitempty"` // 可选：要检索的特定键
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetContextRequest) Reset() {
	*x = GetContextRequest{}
	mi := &file_protos_memory_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetContextRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetContextRequest) ProtoMessage() {}

func (x *GetContextRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 GetContextRequest.ProtoReflect.Descriptor。
func (*GetContextRequest) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{2}
}

func (x *GetContextRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *GetContextRequest) GetKeys() []string {
	if x != nil {
		return x.Keys
	}
	return nil
}

type GetContextResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Context       map[string]string      `protobuf:"bytes,1,rep,name=context,proto3" json:"context,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Success       bool                   `protobuf:"varint,2,opt,name=success,proto3" json:"success,omitempty"`
	Error         string                 `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetContextResponse) Reset() {
	*x = GetContextResponse{}
	mi := &file_protos_memory_service_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetContextResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetContextResponse) ProtoMessage() {}

func (x *GetContextResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 GetContextResponse.ProtoReflect.Descriptor。
func (*GetContextResponse) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{3}
}

func (x *GetContextResponse) GetContext() map[string]string {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *GetContextResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *GetContextResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type DeleteContextRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Keys          []string               `protobuf:"bytes,2,rep,name=keys,proto3" json:"keys,omitempty"` // 可选：要删除的特定键
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteContextRequest) Reset() {
	*x = DeleteContextRequest{}
	mi := &file_protos_memory_service_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteContextRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteContextRequest) ProtoMessage() {}

func (x *DeleteContextRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 DeleteContextRequest.ProtoReflect.Descriptor。
func (*DeleteContextRequest) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteContextRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *DeleteContextRequest) GetKeys() []string {
	if x != nil {
		return x.Keys
	}
	return nil
}

type DeleteContextResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteContextResponse) Reset() {
	*x = DeleteContextResponse{}
	mi := &file_protos_memory_service_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteContextResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteContextResponse) ProtoMessage() {}

func (x *DeleteContextResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 DeleteContextResponse.ProtoReflect.Descriptor。
func (*DeleteContextResponse) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{5}
}

func (x *DeleteContextResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *DeleteContextResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type UpdateContextRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Context       map[string]string      `protobuf:"bytes,2,rep,name=context,proto3" json:"context,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Ttl           int64                  `protobuf:"varint,3,opt,name=ttl,proto3" json:"ttl,omitempty"` // 可选：新的 TTL
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateContextRequest) Reset() {
	*x = UpdateContextRequest{}
	mi := &file_protos_memory_service_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateContextRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateContextRequest) ProtoMessage() {}

func (x *UpdateContextRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 UpdateContextRequest.ProtoReflect.Descriptor。
func (*UpdateContextRequest) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{6}
}

func (x *UpdateContextRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *UpdateContextRequest) GetContext() map[string]string {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *UpdateContextRequest) GetTtl() int64 {
	if x != nil {
		return x.Ttl
	}
	return 0
}

type UpdateContextResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateContextResponse) Reset() {
	*x = UpdateContextResponse{}
	mi := &file_protos_memory_service_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateContextResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateContextResponse) ProtoMessage() {}

func (x *UpdateContextResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 UpdateContextResponse.ProtoReflect.Descriptor。
func (*UpdateContextResponse) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{7}
}

func (x *UpdateContextResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *UpdateContextResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type SaveVectorRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Collection    string                 `protobuf:"bytes,1,opt,name=collection,proto3" json:"collection,omitempty"` // 集合名称
	Vectors       []*VectorData          `protobuf:"bytes,2,rep,name=vectors,proto3" json:"vectors,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SaveVectorRequest) Reset() {
	*x = SaveVectorRequest{}
	mi := &file_protos_memory_service_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SaveVectorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SaveVectorRequest) ProtoMessage() {}

func (x *SaveVectorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 SaveVectorRequest.ProtoReflect.Descriptor。
func (*SaveVectorRequest) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{8}
}

func (x *SaveVectorRequest) GetCollection() string {
	if x != nil {
		return x.Collection
	}
	return ""
}

func (x *SaveVectorRequest) GetVectors() []*VectorData {
	if x != nil {
		return x.Vectors
	}
	return nil
}

type VectorData struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`                                                                                      // 唯一向量 ID
	Vector        []float32              `protobuf:"fixed32,2,rep,packed,name=vector,proto3" json:"vector,omitempty"`                                                                      // 嵌入向量
	Metadata      map[string]string      `protobuf:"bytes,3,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"` // 关联的元数据
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VectorData) Reset() {
	*x = VectorData{}
	mi := &file_protos_memory_service_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VectorData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VectorData) ProtoMessage() {}

func (x *VectorData) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 VectorData.ProtoReflect.Descriptor。
func (*VectorData) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{9}
}

func (x *VectorData) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *VectorData) GetVector() []float32 {
	if x != nil {
		return x.Vector
	}
	return nil
}

func (x *VectorData) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type SaveVectorResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	InsertedIds   []int64                `protobuf:"varint,3,rep,packed,name=inserted_ids,json=insertedIds,proto3" json:"inserted_ids,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SaveVectorResponse) Reset() {
	*x = SaveVectorResponse{}
	mi := &file_protos_memory_service_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SaveVectorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SaveVectorResponse) ProtoMessage() {}

func (x *SaveVectorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 SaveVectorResponse.ProtoReflect.Descriptor。
func (*SaveVectorResponse) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{10}
}

func (x *SaveVectorResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *SaveVectorResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *SaveVectorResponse) GetInsertedIds() []int64 {
	if x != nil {
		return x.InsertedIds
	}
	return nil
}

type SearchSimilarRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Collection    string                 `protobuf:"bytes,1,opt,name=collection,proto3" json:"collection,omitempty"` // 集合名称
	QueryVector   []float32              `protobuf:"fixed32,2,rep,packed,name=query_vector,json=queryVector,proto3" json:"query_vector,omitempty"`
	TopK          int32                  `protobuf:"varint,3,opt,name=top_k,json=topK,proto3" json:"top_k,omitempty"`                                                                  // 返回结果数量
	Filter        map[string]string      `protobuf:"bytes,4,rep,name=filter,proto3" json:"filter,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"` // 元数据过滤条件
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchSimilarRequest) Reset() {
	*x = SearchSimilarRequest{}
	mi := &file_protos_memory_service_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchSimilarRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchSimilarRequest) ProtoMessage() {}

func (x *SearchSimilarRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 SearchSimilarRequest.ProtoReflect.Descriptor。
func (*SearchSimilarRequest) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{11}
}

func (x *SearchSimilarRequest) GetCollection() string {
	if x != nil {
		return x.Collection
	}
	return ""
}

func (x *SearchSimilarRequest) GetQueryVector() []float32 {
	if x != nil {
		return x.QueryVector
	}
	return nil
}

func (x *SearchSimilarRequest) GetTopK() int32 {
	if x != nil {
		return x.TopK
	}
	return 0
}

func (x *SearchSimilarRequest) GetFilter() map[string]string {
	if x != nil {
		return x.Filter
	}
	return nil
}

type SearchSimilarResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	Results       []*SearchResult        `protobuf:"bytes,3,rep,name=results,proto3" json:"results,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchSimilarResponse) Reset() {
	*x = SearchSimilarResponse{}
	mi := &file_protos_memory_service_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchSimilarResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchSimilarResponse) ProtoMessage() {}

func (x *SearchSimilarResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 SearchSimilarResponse.ProtoReflect.Descriptor。
func (*SearchSimilarResponse) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{12}
}

func (x *SearchSimilarResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *SearchSimilarResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *SearchSimilarResponse) GetResults() []*SearchResult {
	if x != nil {
		return x.Results
	}
	return nil
}

type SearchResult struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Vector        []float32              `protobuf:"fixed32,2,rep,packed,name=vector,proto3" json:"vector,omitempty"`
	Metadata      map[string]string      `protobuf:"bytes,3,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Score         float32                `protobuf:"fixed32,4,opt,name=score,proto3" json:"score,omitempty"` // 相似度分数
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchResult) Reset() {
	*x = SearchResult{}
	mi := &file_protos_memory_service_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchResult) ProtoMessage() {}

func (x *SearchResult) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 SearchResult.ProtoReflect.Descriptor。
func (*SearchResult) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{13}
}

func (x *SearchResult) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *SearchResult) GetVector() []float32 {
	if x != nil {
		return x.Vector
	}
	return nil
}

func (x *SearchResult) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *SearchResult) GetScore() float32 {
	if x != nil {
		return x.Score
	}
	return 0
}

type DeleteVectorRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Collection    string                 `protobuf:"bytes,1,opt,name=collection,proto3" json:"collection,omitempty"`
	Ids           []int64                `protobuf:"varint,2,rep,packed,name=ids,proto3" json:"ids,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteVectorRequest) Reset() {
	*x = DeleteVectorRequest{}
	mi := &file_protos_memory_service_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteVectorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteVectorRequest) ProtoMessage() {}

func (x *DeleteVectorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 DeleteVectorRequest.ProtoReflect.Descriptor。
func (*DeleteVectorRequest) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{14}
}

func (x *DeleteVectorRequest) GetCollection() string {
	if x != nil {
		return x.Collection
	}
	return ""
}

func (x *DeleteVectorRequest) GetIds() []int64 {
	if x != nil {
		return x.Ids
	}
	return nil
}

type DeleteVectorResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteVectorResponse) Reset() {
	*x = DeleteVectorResponse{}
	mi := &file_protos_memory_service_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteVectorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteVectorResponse) ProtoMessage() {}

func (x *DeleteVectorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_memory_service_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请改用 DeleteVectorResponse.ProtoReflect.Descriptor。
func (*DeleteVectorResponse) Descriptor() ([]byte, []int) {
	return file_protos_memory_service_proto_rawDescGZIP(), []int{15}
}

func (x *DeleteVectorResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *DeleteVectorResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_protos_memory_service_proto protoreflect.FileDescriptor

const file_protos_memory_service_proto_rawDesc = "" +
	"\n" +
	"\x1bprotos/memory_service.proto\x12\x06memory\"\xbe\x01\n" +
	"\x12SaveContextRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12A\n" +
	"\acontext\x18\x02 \x03(\v2'.memory.SaveContextRequest.ContextEntryR\acontext\x12\x10\n" +
	"\x03ttl\x18\x03 \x01(\x03R\x03ttl\x1a:\n" +
	"\fContextEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"E\n" +
	"\x13SaveContextResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error\"@\n" +
	"\x11GetContextRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12\x12\n" +
	"\x04keys\x18\x02 \x03(\tR\x04keys\"\xc3\x01\n" +
	"\x12GetContextResponse\x12A\n" +
	"\acontext\x18\x01 \x03(\v2'.memory.GetContextResponse.ContextEntryR\acontext\x12\x18\n" +
	"\asuccess\x18\x02 \x01(\bR\asuccess\x12\x14\n" +
	"\x05error\x18\x03 \x01(\tR\x05error\x1a:\n" +
	"\fContextEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"C\n" +
	"\x14DeleteContextRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12\x12\n" +
	"\x04keys\x18\x02 \x03(\tR\x04keys\"G\n" +
	"\x15DeleteContextResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error\"\xc2\x01\n" +
	"\x14UpdateContextRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12C\n" +
	"\acontext\x18\x02 \x03(\v2).memory.UpdateContextRequest.ContextEntryR\acontext\x12\x10\n" +
	"\x03ttl\x18\x03 \x01(\x03R\x03ttl\x1a:\n" +
	"\fContextEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"G\n" +
	"\x15UpdateContextResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error\"a\n" +
	"\x11SaveVectorRequest\x12\x1e\n" +
	"\n" +
	"collection\x18\x01 \x01(\tR\n" +
	"collection\x12,\n" +
	"\avectors\x18\x02 \x03(\v2\x12.memory.VectorDataR\avectors\"\xaf\x01\n" +
	"\n" +
	"VectorData\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x16\n" +
	"\x06vector\x18\x02 \x03(\x02R\x06vector\x12<\n" +
	"\bmetadata\x18\x03 \x03(\v2 .memory.VectorData.MetadataEntryR\bmetadata\x1a;\n" +
	"\rMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"g\n" +
	"\x12SaveVectorResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error\x12!\n" +
	"\finserted_ids\x18\x03 \x03(\x03R\vinsertedIds\"\xeb\x01\n" +
	"\x14SearchSimilarRequest\x12\x1e\n" +
	"\n" +
	"collection\x18\x01 \x01(\tR\n" +
	"collection\x12!\n" +
	"\fquery_vector\x18\x02 \x03(\x02R\vqueryVector\x12\x13\n" +
	"\x05top_k\x18\x03 \x01(\x05R\x04topK\x12@\n" +
	"\x06filter\x18\x04 \x03(\v2(.memory.SearchSimilarRequest.FilterEntryR\x06filter\x1a9\n" +
	"\vFilterEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"w\n" +
	"\x15SearchSimilarResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error\x12.\n" +
	"\aresults\x18\x03 \x03(\v2\x14.memory.SearchResultR\aresults\"\xc9\x01\n" +
	"\fSearchResult\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x16\n" +
	"\x06vector\x18\x02 \x03(\x02R\x06vector\x12>\n" +
	"\bmetadata\x18\x03 \x03(\v2\".memory.SearchResult.MetadataEntryR\bmetadata\x12\x14\n" +
	"\x05score\x18\x04 \x01(\x02R\x05score\x1a;\n" +
	"\rMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"G\n" +
	"\x13DeleteVectorRequest\x12\x1e\n" +
	"\n" +
	"collection\x18\x01 \x01(\tR\n" +
	"collection\x12\x10\n" +
	"\x03ids\x18\x02 \x03(\x03R\x03ids\"F\n" +
	"\x14DeleteVectorResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error2\x96\x04\n" +
	"\rMemoryService\x12F\n" +
	"\vSaveContext\x12\x1a.memory.SaveContextRequest\x1a\x1b.memory.SaveContextResponse\x12C\n" +
	"\n" +
	"GetContext\x12\x19.memory.GetContextRequest\x1a\x1a.memory.GetContextResponse\x12L\n" +
	"\rDeleteContext\x12\x1c.memory.DeleteContextRequest\x1a\x1d.memory.DeleteContextResponse\x12L\n" +
	"\rUpdateContext\x12\x1c.memory.UpdateContextRequest\x1a\x1d.memory.UpdateContextResponse\x12C\n" +
	"\n" +
	"SaveVector\x12\x19.memory.SaveVectorRequest\x1a\x1a.memory.SaveVectorResponse\x12L\n" +
	"\rSearchSimilar\x12\x1c.memory.SearchSimilarRequest\x1a\x1d.memory.SearchSimilarResponse\x12I\n" +
	"\fDeleteVector\x12\x1b.memory.DeleteVectorRequest\x1a\x1c.memory.DeleteVectorResponseB\tZ\a./protob\x06proto3"

var (
	file_protos_memory_service_proto_rawDescOnce sync.Once
	file_protos_memory_service_proto_rawDescData []byte
)

func file_protos_memory_service_proto_rawDescGZIP() []byte {
	file_protos_memory_service_proto_rawDescOnce.Do(func() {
		file_protos_memory_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_memory_service_proto_rawDesc), len(file_protos_memory_service_proto_rawDesc)))
	})
	return file_protos_memory_service_proto_rawDescData
}

var file_protos_memory_service_proto_msgTypes = make([]protoimpl.MessageInfo, 22)
var file_protos_memory_service_proto_goTypes = []any{
	(*SaveContextRequest)(nil),    // 0: memory.SaveContextRequest
	(*SaveContextResponse)(nil),   // 1: memory.SaveContextResponse
	(*GetContextRequest)(nil),     // 2: memory.GetContextRequest
	(*GetContextResponse)(nil),    // 3: memory.GetContextResponse
	(*DeleteContextRequest)(nil),  // 4: memory.DeleteContextRequest
	(*DeleteContextResponse)(nil), // 5: memory.DeleteContextResponse
	(*UpdateContextRequest)(nil),  // 6: memory.UpdateContextRequest
	(*UpdateContextResponse)(nil), // 7: memory.UpdateContextResponse
	(*SaveVectorRequest)(nil),     // 8: memory.SaveVectorRequest
	(*VectorData)(nil),            // 9: memory.VectorData
	(*SaveVectorResponse)(nil),    // 10: memory.SaveVectorResponse
	(*SearchSimilarRequest)(nil),  // 11: memory.SearchSimilarRequest
	(*SearchSimilarResponse)(nil), // 12: memory.SearchSimilarResponse
	(*SearchResult)(nil),          // 13: memory.SearchResult
	(*DeleteVectorRequest)(nil),   // 14: memory.DeleteVectorRequest
	(*DeleteVectorResponse)(nil),  // 15: memory.DeleteVectorResponse
	nil,                           // 16: memory.SaveContextRequest.ContextEntry
	nil,                           // 17: memory.GetContextResponse.ContextEntry
	nil,                           // 18: memory.UpdateContextRequest.ContextEntry
	nil,                           // 19: memory.VectorData.MetadataEntry
	nil,                           // 20: memory.SearchSimilarRequest.FilterEntry
	nil,                           // 21: memory.SearchResult.MetadataEntry
}
var file_protos_memory_service_proto_depIdxs = []int32{
	16, // 0: memory.SaveContextRequest.context:type_name -> memory.SaveContextRequest.ContextEntry
	17, // 1: memory.GetContextResponse.context:type_name -> memory.GetContextResponse.ContextEntry
	18, // 2: memory.UpdateContextRequest.context:type_name -> memory.UpdateContextRequest.ContextEntry
	9,  // 3: memory.SaveVectorRequest.vectors:type_name -> memory.VectorData
	19, // 4: memory.VectorData.metadata:type_name -> memory.VectorData.MetadataEntry
	20, // 5: memory.SearchSimilarRequest.filter:type_name -> memory.SearchSimilarRequest.FilterEntry
	13, // 6: memory.SearchSimilarResponse.results:type_name -> memory.SearchResult
	21, // 7: memory.SearchResult.metadata:type_name -> memory.SearchResult.MetadataEntry
	0,  // 8: memory.MemoryService.SaveContext:input_type -> memory.SaveContextRequest
	2,  // 9: memory.MemoryService.GetContext:input_type -> memory.GetContextRequest
	4,  // 10: memory.MemoryService.DeleteContext:input_type -> memory.DeleteContextRequest
	6,  // 11: memory.MemoryService.UpdateContext:input_type -> memory.UpdateContextRequest
	8,  // 12: memory.MemoryService.SaveVector:input_type -> memory.SaveVectorRequest
	11, // 13: memory.MemoryService.SearchSimilar:input_type -> memory.SearchSimilarRequest
	14, // 14: memory.MemoryService.DeleteVector:input_type -> memory.DeleteVectorRequest
	1,  // 15: memory.MemoryService.SaveContext:output_type -> memory.SaveContextResponse
	3,  // 16: memory.MemoryService.GetContext:output_type -> memory.GetContextResponse
	5,  // 17: memory.MemoryService.DeleteContext:output_type -> memory.DeleteContextResponse
	7,  // 18: memory.MemoryService.UpdateContext:output_type -> memory.UpdateContextResponse
	10, // 19: memory.MemoryService.SaveVector:output_type -> memory.SaveVectorResponse
	12, // 20: memory.MemoryService.SearchSimilar:output_type -> memory.SearchSimilarResponse
	15, // 21: memory.MemoryService.DeleteVector:output_type -> memory.DeleteVectorResponse
	15, // [15:22] is the sub-list for method output_type
	8,  // [8:15] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_protos_memory_service_proto_init() }
func file_protos_memory_service_proto_init() {
	if File_protos_memory_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_memory_service_proto_rawDesc), len(file_protos_memory_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   22,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_protos_memory_service_proto_goTypes,
		DependencyIndexes: file_protos_memory_service_proto_depIdxs,
		MessageInfos:      file_protos_memory_service_proto_msgTypes,
	}.Build()
	File_protos_memory_service_proto = out.File
	file_protos_memory_service_proto_goTypes = nil
	file_protos_memory_service_proto_depIdxs = nil
}
