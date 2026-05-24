// 由 protoc-gen-go-grpc 自动生成。请勿手动编辑。
// 版本:
// - protoc-gen-go-grpc v1.6.1
// - protoc             v7.34.1
// 源文件: protos/tool_service.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// 这是一个编译时断言，用于确保此生成的文件
// 与其编译所针对的 gRPC 包兼容。
// 要求 gRPC-Go v1.64.0 或更高版本。
const _ = grpc.SupportPackageIsVersion9

const (
	ToolService_AnalyzeCodeError_FullMethodName        = "/tool.ToolService/AnalyzeCodeError"
	ToolService_QuerySyntax_FullMethodName             = "/tool.ToolService/QuerySyntax"
	ToolService_GenerateProblemSolution_FullMethodName = "/tool.ToolService/GenerateProblemSolution"
)

// ToolServiceClient 是 ToolService 服务的客户端 API。
//
// 关于 ctx 的使用语义以及流式 RPC 的关闭/结束，请参考 https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream。
type ToolServiceClient interface {
	AnalyzeCodeError(ctx context.Context, in *AnalyzeCodeErrorRequest, opts ...grpc.CallOption) (*AnalyzeCodeErrorResponse, error)
	QuerySyntax(ctx context.Context, in *QuerySyntaxRequest, opts ...grpc.CallOption) (*QuerySyntaxResponse, error)
	GenerateProblemSolution(ctx context.Context, in *GenerateProblemSolutionRequest, opts ...grpc.CallOption) (*GenerateProblemSolutionResponse, error)
}

type toolServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewToolServiceClient(cc grpc.ClientConnInterface) ToolServiceClient {
	return &toolServiceClient{cc}
}

func (c *toolServiceClient) AnalyzeCodeError(ctx context.Context, in *AnalyzeCodeErrorRequest, opts ...grpc.CallOption) (*AnalyzeCodeErrorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AnalyzeCodeErrorResponse)
	err := c.cc.Invoke(ctx, ToolService_AnalyzeCodeError_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *toolServiceClient) QuerySyntax(ctx context.Context, in *QuerySyntaxRequest, opts ...grpc.CallOption) (*QuerySyntaxResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QuerySyntaxResponse)
	err := c.cc.Invoke(ctx, ToolService_QuerySyntax_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *toolServiceClient) GenerateProblemSolution(ctx context.Context, in *GenerateProblemSolutionRequest, opts ...grpc.CallOption) (*GenerateProblemSolutionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GenerateProblemSolutionResponse)
	err := c.cc.Invoke(ctx, ToolService_GenerateProblemSolution_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ToolServiceServer 是 ToolService 服务的服务端 API。
// 所有实现都必须嵌入 UnimplementedToolServiceServer
// 以保证向前兼容性。
type ToolServiceServer interface {
	AnalyzeCodeError(context.Context, *AnalyzeCodeErrorRequest) (*AnalyzeCodeErrorResponse, error)
	QuerySyntax(context.Context, *QuerySyntaxRequest) (*QuerySyntaxResponse, error)
	GenerateProblemSolution(context.Context, *GenerateProblemSolutionRequest) (*GenerateProblemSolutionResponse, error)
	mustEmbedUnimplementedToolServiceServer()
}

// UnimplementedToolServiceServer 必须被嵌入以获得
// 向前兼容的实现。
//
// 注意：应该以值类型而不是指针类型嵌入，以避免方法调用时出现空指针解引用。
type UnimplementedToolServiceServer struct{}

func (UnimplementedToolServiceServer) AnalyzeCodeError(context.Context, *AnalyzeCodeErrorRequest) (*AnalyzeCodeErrorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method AnalyzeCodeError not implemented")
}
func (UnimplementedToolServiceServer) QuerySyntax(context.Context, *QuerySyntaxRequest) (*QuerySyntaxResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method QuerySyntax not implemented")
}
func (UnimplementedToolServiceServer) GenerateProblemSolution(context.Context, *GenerateProblemSolutionRequest) (*GenerateProblemSolutionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GenerateProblemSolution not implemented")
}
func (UnimplementedToolServiceServer) mustEmbedUnimplementedToolServiceServer() {}
func (UnimplementedToolServiceServer) testEmbeddedByValue()                     {}

// UnsafeToolServiceServer 可以嵌入以选择退出此服务的向前兼容性。
// 不推荐使用此接口，因为对 ToolServiceServer 添加的方法
// 将导致编译错误。
type UnsafeToolServiceServer interface {
	mustEmbedUnimplementedToolServiceServer()
}

func RegisterToolServiceServer(s grpc.ServiceRegistrar, srv ToolServiceServer) {
	// 如果以下调用发生 panic，表明 UnimplementedToolServiceServer
	// 是通过指针嵌入且为 nil。如果调用了未实现的方法，这将导致 panic，
	// 因此我们在初始化时进行测试，以防止后续运行时因 I/O 而触发。
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ToolService_ServiceDesc, srv)
}

func _ToolService_AnalyzeCodeError_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnalyzeCodeErrorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ToolServiceServer).AnalyzeCodeError(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ToolService_AnalyzeCodeError_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ToolServiceServer).AnalyzeCodeError(ctx, req.(*AnalyzeCodeErrorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ToolService_QuerySyntax_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySyntaxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ToolServiceServer).QuerySyntax(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ToolService_QuerySyntax_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ToolServiceServer).QuerySyntax(ctx, req.(*QuerySyntaxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ToolService_GenerateProblemSolution_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateProblemSolutionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ToolServiceServer).GenerateProblemSolution(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ToolService_GenerateProblemSolution_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ToolServiceServer).GenerateProblemSolution(ctx, req.(*GenerateProblemSolutionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ToolService_ServiceDesc 是 ToolService 服务的 grpc.ServiceDesc。
// 它仅供 grpc.RegisterService 直接使用，
// 不应被内省或修改（即使是作为副本）
var ToolService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tool.ToolService",
	HandlerType: (*ToolServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AnalyzeCodeError",
			Handler:    _ToolService_AnalyzeCodeError_Handler,
		},
		{
			MethodName: "QuerySyntax",
			Handler:    _ToolService_QuerySyntax_Handler,
		},
		{
			MethodName: "GenerateProblemSolution",
			Handler:    _ToolService_GenerateProblemSolution_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/tool_service.proto",
}
