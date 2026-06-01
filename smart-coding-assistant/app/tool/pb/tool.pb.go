// 由 protoc-gen-go 生成。请勿编辑。
// 版本:
// 	protoc-gen-go v1.36.11
// 	protoc        v7.34.1
// 源文件: app/tool/tool.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// 验证生成的代码是最新版本。
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// 验证 runtime/protoimpl 是最新版本。
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AnalyzeCodeErrorRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Code          string                 `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`
	ErrorMessage  string                 `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	Language      string                 `protobuf:"bytes,3,opt,name=language,proto3" json:"language,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AnalyzeCodeErrorRequest) Reset() {
	*x = AnalyzeCodeErrorRequest{}
	mi := &file_app_tool_tool_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AnalyzeCodeErrorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AnalyzeCodeErrorRequest) ProtoMessage() {}

func (x *AnalyzeCodeErrorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_tool_tool_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请使用 AnalyzeCodeErrorRequest.ProtoReflect.Descriptor 代替。
func (*AnalyzeCodeErrorRequest) Descriptor() ([]byte, []int) {
	return file_app_tool_tool_proto_rawDescGZIP(), []int{0}
}

func (x *AnalyzeCodeErrorRequest) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *AnalyzeCodeErrorRequest) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *AnalyzeCodeErrorRequest) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

type AnalyzeCodeErrorResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Analysis      string                 `protobuf:"bytes,1,opt,name=analysis,proto3" json:"analysis,omitempty"`
	SuggestedFix  string                 `protobuf:"bytes,2,opt,name=suggested_fix,json=suggestedFix,proto3" json:"suggested_fix,omitempty"`
	Success       bool                   `protobuf:"varint,3,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AnalyzeCodeErrorResponse) Reset() {
	*x = AnalyzeCodeErrorResponse{}
	mi := &file_app_tool_tool_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AnalyzeCodeErrorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AnalyzeCodeErrorResponse) ProtoMessage() {}

func (x *AnalyzeCodeErrorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_tool_tool_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请使用 AnalyzeCodeErrorResponse.ProtoReflect.Descriptor 代替。
func (*AnalyzeCodeErrorResponse) Descriptor() ([]byte, []int) {
	return file_app_tool_tool_proto_rawDescGZIP(), []int{1}
}

func (x *AnalyzeCodeErrorResponse) GetAnalysis() string {
	if x != nil {
		return x.Analysis
	}
	return ""
}

func (x *AnalyzeCodeErrorResponse) GetSuggestedFix() string {
	if x != nil {
		return x.SuggestedFix
	}
	return ""
}

func (x *AnalyzeCodeErrorResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type QuerySyntaxRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Language      string                 `protobuf:"bytes,1,opt,name=language,proto3" json:"language,omitempty"`
	Query         string                 `protobuf:"bytes,2,opt,name=query,proto3" json:"query,omitempty"`
	Context       string                 `protobuf:"bytes,3,opt,name=context,proto3" json:"context,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QuerySyntaxRequest) Reset() {
	*x = QuerySyntaxRequest{}
	mi := &file_app_tool_tool_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QuerySyntaxRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QuerySyntaxRequest) ProtoMessage() {}

func (x *QuerySyntaxRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_tool_tool_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请使用 QuerySyntaxRequest.ProtoReflect.Descriptor 代替。
func (*QuerySyntaxRequest) Descriptor() ([]byte, []int) {
	return file_app_tool_tool_proto_rawDescGZIP(), []int{2}
}

func (x *QuerySyntaxRequest) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

func (x *QuerySyntaxRequest) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

func (x *QuerySyntaxRequest) GetContext() string {
	if x != nil {
		return x.Context
	}
	return ""
}

type QuerySyntaxResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Explanation   string                 `protobuf:"bytes,1,opt,name=explanation,proto3" json:"explanation,omitempty"`
	Example       string                 `protobuf:"bytes,2,opt,name=example,proto3" json:"example,omitempty"`
	Success       bool                   `protobuf:"varint,3,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QuerySyntaxResponse) Reset() {
	*x = QuerySyntaxResponse{}
	mi := &file_app_tool_tool_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QuerySyntaxResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QuerySyntaxResponse) ProtoMessage() {}

func (x *QuerySyntaxResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_tool_tool_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请使用 QuerySyntaxResponse.ProtoReflect.Descriptor 代替。
func (*QuerySyntaxResponse) Descriptor() ([]byte, []int) {
	return file_app_tool_tool_proto_rawDescGZIP(), []int{3}
}

func (x *QuerySyntaxResponse) GetExplanation() string {
	if x != nil {
		return x.Explanation
	}
	return ""
}

func (x *QuerySyntaxResponse) GetExample() string {
	if x != nil {
		return x.Example
	}
	return ""
}

func (x *QuerySyntaxResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type GenerateProblemSolutionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Problem       string                 `protobuf:"bytes,1,opt,name=problem,proto3" json:"problem,omitempty"`
	Difficulty    string                 `protobuf:"bytes,2,opt,name=difficulty,proto3" json:"difficulty,omitempty"`
	Language      string                 `protobuf:"bytes,3,opt,name=language,proto3" json:"language,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateProblemSolutionRequest) Reset() {
	*x = GenerateProblemSolutionRequest{}
	mi := &file_app_tool_tool_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateProblemSolutionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateProblemSolutionRequest) ProtoMessage() {}

func (x *GenerateProblemSolutionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_tool_tool_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请使用 GenerateProblemSolutionRequest.ProtoReflect.Descriptor 代替。
func (*GenerateProblemSolutionRequest) Descriptor() ([]byte, []int) {
	return file_app_tool_tool_proto_rawDescGZIP(), []int{4}
}

func (x *GenerateProblemSolutionRequest) GetProblem() string {
	if x != nil {
		return x.Problem
	}
	return ""
}

func (x *GenerateProblemSolutionRequest) GetDifficulty() string {
	if x != nil {
		return x.Difficulty
	}
	return ""
}

func (x *GenerateProblemSolutionRequest) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

type GenerateProblemSolutionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Approach      string                 `protobuf:"bytes,1,opt,name=approach,proto3" json:"approach,omitempty"`
	Code          string                 `protobuf:"bytes,2,opt,name=code,proto3" json:"code,omitempty"`
	Explanation   string                 `protobuf:"bytes,3,opt,name=explanation,proto3" json:"explanation,omitempty"`
	Success       bool                   `protobuf:"varint,4,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateProblemSolutionResponse) Reset() {
	*x = GenerateProblemSolutionResponse{}
	mi := &file_app_tool_tool_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateProblemSolutionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateProblemSolutionResponse) ProtoMessage() {}

func (x *GenerateProblemSolutionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_tool_tool_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// 已弃用：请使用 GenerateProblemSolutionResponse.ProtoReflect.Descriptor 代替。
func (*GenerateProblemSolutionResponse) Descriptor() ([]byte, []int) {
	return file_app_tool_tool_proto_rawDescGZIP(), []int{5}
}

func (x *GenerateProblemSolutionResponse) GetApproach() string {
	if x != nil {
		return x.Approach
	}
	return ""
}

func (x *GenerateProblemSolutionResponse) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *GenerateProblemSolutionResponse) GetExplanation() string {
	if x != nil {
		return x.Explanation
	}
	return ""
}

func (x *GenerateProblemSolutionResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_app_tool_tool_proto protoreflect.FileDescriptor

const file_app_tool_tool_proto_rawDesc = "" +
	"\n" +
	"\x13app/tool/tool.proto\x12\x04tool\"n\n" +
	"\x17AnalyzeCodeErrorRequest\x12\x12\n" +
	"\x04code\x18\x01 \x01(\tR\x04code\x12#\n" +
	"\rerror_message\x18\x02 \x01(\tR\ferrorMessage\x12\x1a\n" +
	"\blanguage\x18\x03 \x01(\tR\blanguage\"u\n" +
	"\x18AnalyzeCodeErrorResponse\x12\x1a\n" +
	"\banalysis\x18\x01 \x01(\tR\banalysis\x12#\n" +
	"\rsuggested_fix\x18\x02 \x01(\tR\fsuggestedFix\x12\x18\n" +
	"\asuccess\x18\x03 \x01(\bR\asuccess\"`\n" +
	"\x12QuerySyntaxRequest\x12\x1a\n" +
	"\blanguage\x18\x01 \x01(\tR\blanguage\x12\x14\n" +
	"\x05query\x18\x02 \x01(\tR\x05query\x12\x18\n" +
	"\acontext\x18\x03 \x01(\tR\acontext\"k\n" +
	"\x13QuerySyntaxResponse\x12 \n" +
	"\vexplanation\x18\x01 \x01(\tR\vexplanation\x12\x18\n" +
	"\aexample\x18\x02 \x01(\tR\aexample\x12\x18\n" +
	"\asuccess\x18\x03 \x01(\bR\asuccess\"v\n" +
	"\x1eGenerateProblemSolutionRequest\x12\x18\n" +
	"\aproblem\x18\x01 \x01(\tR\aproblem\x12\x1e\n" +
	"\n" +
	"difficulty\x18\x02 \x01(\tR\n" +
	"difficulty\x12\x1a\n" +
	"\blanguage\x18\x03 \x01(\tR\blanguage\"\x8d\x01\n" +
	"\x1fGenerateProblemSolutionResponse\x12\x1a\n" +
	"\bapproach\x18\x01 \x01(\tR\bapproach\x12\x12\n" +
	"\x04code\x18\x02 \x01(\tR\x04code\x12 \n" +
	"\vexplanation\x18\x03 \x01(\tR\vexplanation\x12\x18\n" +
	"\asuccess\x18\x04 \x01(\bR\asuccess2\x8c\x02\n" +
	"\vToolService\x12Q\n" +
	"\x10AnalyzeCodeError\x12\x1d.tool.AnalyzeCodeErrorRequest\x1a\x1e.tool.AnalyzeCodeErrorResponse\x12B\n" +
	"\vQuerySyntax\x12\x18.tool.QuerySyntaxRequest\x1a\x19.tool.QuerySyntaxResponse\x12f\n" +
	"\x17GenerateProblemSolution\x12$.tool.GenerateProblemSolutionRequest\x1a%.tool.GenerateProblemSolutionResponseB$Z\"smart-coding-assistant/app/tool/pbb\x06proto3"

var (
	file_app_tool_tool_proto_rawDescOnce sync.Once
	file_app_tool_tool_proto_rawDescData []byte
)

func file_app_tool_tool_proto_rawDescGZIP() []byte {
	file_app_tool_tool_proto_rawDescOnce.Do(func() {
		file_app_tool_tool_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_tool_tool_proto_rawDesc), len(file_app_tool_tool_proto_rawDesc)))
	})
	return file_app_tool_tool_proto_rawDescData
}

var file_app_tool_tool_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_app_tool_tool_proto_goTypes = []any{
	(*AnalyzeCodeErrorRequest)(nil),         // 0: tool.AnalyzeCodeErrorRequest
	(*AnalyzeCodeErrorResponse)(nil),        // 1: tool.AnalyzeCodeErrorResponse
	(*QuerySyntaxRequest)(nil),              // 2: tool.QuerySyntaxRequest
	(*QuerySyntaxResponse)(nil),             // 3: tool.QuerySyntaxResponse
	(*GenerateProblemSolutionRequest)(nil),  // 4: tool.GenerateProblemSolutionRequest
	(*GenerateProblemSolutionResponse)(nil), // 5: tool.GenerateProblemSolutionResponse
}
var file_app_tool_tool_proto_depIdxs = []int32{
	0, // 0: tool.ToolService.AnalyzeCodeError:input_type -> tool.AnalyzeCodeErrorRequest
	2, // 1: tool.ToolService.QuerySyntax:input_type -> tool.QuerySyntaxRequest
	4, // 2: tool.ToolService.GenerateProblemSolution:input_type -> tool.GenerateProblemSolutionRequest
	1, // 3: tool.ToolService.AnalyzeCodeError:output_type -> tool.AnalyzeCodeErrorResponse
	3, // 4: tool.ToolService.QuerySyntax:output_type -> tool.QuerySyntaxResponse
	5, // 5: tool.ToolService.GenerateProblemSolution:output_type -> tool.GenerateProblemSolutionResponse
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_app_tool_tool_proto_init() }
func file_app_tool_tool_proto_init() {
	if File_app_tool_tool_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_tool_tool_proto_rawDesc), len(file_app_tool_tool_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_app_tool_tool_proto_goTypes,
		DependencyIndexes: file_app_tool_tool_proto_depIdxs,
		MessageInfos:      file_app_tool_tool_proto_msgTypes,
	}.Build()
	File_app_tool_tool_proto = out.File
	file_app_tool_tool_proto_goTypes = nil
	file_app_tool_tool_proto_depIdxs = nil
}
