// 由 protoc-gen-go-grpc 自动生成。请勿手动编辑。
// 版本:
// - protoc-gen-go-grpc v1.6.1
// - protoc             v7.34.1
// 源文件: protos/core_service.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// 这是一个编译时断言，用于确保此生成文件与所编译的 grpc 包兼容。
// 要求 gRPC-Go v1.64.0 或更高版本。
const _ = grpc.SupportPackageIsVersion9

const (
	CoreService_Chat_FullMethodName       = "/core.CoreService/Chat"
	CoreService_GetHistory_FullMethodName = "/core.CoreService/GetHistory"
)

// CoreServiceClient 是 CoreService 服务的客户端 API。
//
// 关于 ctx 的使用以及流式 RPC 的关闭/结束语义，请参考 https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream。
type CoreServiceClient interface {
	Chat(ctx context.Context, in *ChatRequest, opts ...grpc.CallOption) (*ChatResponse, error)
	GetHistory(ctx context.Context, in *GetHistoryRequest, opts ...grpc.CallOption) (*GetHistoryResponse, error)
}

type coreServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCoreServiceClient(cc grpc.ClientConnInterface) CoreServiceClient {
	return &coreServiceClient{cc}
}

func (c *coreServiceClient) Chat(ctx context.Context, in *ChatRequest, opts ...grpc.CallOption) (*ChatResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChatResponse)
	err := c.cc.Invoke(ctx, CoreService_Chat_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreServiceClient) GetHistory(ctx context.Context, in *GetHistoryRequest, opts ...grpc.CallOption) (*GetHistoryResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetHistoryResponse)
	err := c.cc.Invoke(ctx, CoreService_GetHistory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CoreServiceServer 是 CoreService 服务的服务端 API。
// 所有实现都必须嵌入 UnimplementedCoreServiceServer
// 以保证向前兼容性。
type CoreServiceServer interface {
	Chat(context.Context, *ChatRequest) (*ChatResponse, error)
	GetHistory(context.Context, *GetHistoryRequest) (*GetHistoryResponse, error)
	mustEmbedUnimplementedCoreServiceServer()
}

// UnimplementedCoreServiceServer 必须被嵌入以保证向前兼容的实现。
//
// 注意：应该以值类型而不是指针类型嵌入，以避免方法调用时出现空指针解引用。
type UnimplementedCoreServiceServer struct{}

func (UnimplementedCoreServiceServer) Chat(context.Context, *ChatRequest) (*ChatResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Chat not implemented")
}
func (UnimplementedCoreServiceServer) GetHistory(context.Context, *GetHistoryRequest) (*GetHistoryResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetHistory not implemented")
}
func (UnimplementedCoreServiceServer) mustEmbedUnimplementedCoreServiceServer() {}
func (UnimplementedCoreServiceServer) testEmbeddedByValue()                     {}

// UnsafeCoreServiceServer 可以嵌入以选择退出此服务的向前兼容性。
// 不推荐使用此接口，因为向 CoreServiceServer 添加新方法将导致编译错误。
type UnsafeCoreServiceServer interface {
	mustEmbedUnimplementedCoreServiceServer()
}

func RegisterCoreServiceServer(s grpc.ServiceRegistrar, srv CoreServiceServer) {
	// 如果以下调用发生 panic，说明 UnimplementedCoreServiceServer 是以指针类型嵌入的且为 nil。
	// 这将在未实现的方法被调用时引发 panic，因此我们在初始化时检测此问题，
	// 以避免后续运行时因 I/O 操作而触发。
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CoreService_ServiceDesc, srv)
}

func _CoreService_Chat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServiceServer).Chat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CoreService_Chat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServiceServer).Chat(ctx, req.(*ChatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CoreService_GetHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServiceServer).GetHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CoreService_GetHistory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServiceServer).GetHistory(ctx, req.(*GetHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CoreService_ServiceDesc 是 CoreService 服务的 grpc.ServiceDesc。
// 它仅供 grpc.RegisterService 直接使用，
// 不应被内省或修改（即使是复制也不行）。
var CoreService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "core.CoreService",
	HandlerType: (*CoreServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Chat",
			Handler:    _CoreService_Chat_Handler,
		},
		{
			MethodName: "GetHistory",
			Handler:    _CoreService_GetHistory_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/core_service.proto",
}
