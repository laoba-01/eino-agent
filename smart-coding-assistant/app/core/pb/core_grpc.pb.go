// 由 protoc-gen-go-grpc 生成。请勿编辑。
// 版本:
// - protoc-gen-go-grpc v1.6.1
// - protoc             v7.34.1
// 源文件: app/core/core.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// 这是编译时断言，用于确保此生成的文件
// 与其编译依赖的 grpc 包兼容。
// 需要 gRPC-Go v1.64.0 或更高版本。
const _ = grpc.SupportPackageIsVersion9

const (
	CoreService_Chat_FullMethodName       = "/core.CoreService/Chat"
	CoreService_GetHistory_FullMethodName = "/core.CoreService/GetHistory"
)

// CoreServiceClient 是 CoreService 服务的客户端 API。
// 关于 ctx 的使用语义和流式 RPC 的关闭/结束，请参阅 https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream。
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
// 所有实现必须嵌入 UnimplementedCoreServiceServer
// 以保证向前兼容性。
type CoreServiceServer interface {
	Chat(context.Context, *ChatRequest) (*ChatResponse, error)
	GetHistory(context.Context, *GetHistoryRequest) (*GetHistoryResponse, error)
	mustEmbedUnimplementedCoreServiceServer()
}

// 必须嵌入 UnimplementedCoreServiceServer 以获得
// 向前兼容的实现。
// 注意：应通过值而非指针嵌入，以避免
// 调用方法时发生空指针解引用。
type UnimplementedCoreServiceServer struct{}

func (UnimplementedCoreServiceServer) Chat(context.Context, *ChatRequest) (*ChatResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Chat not implemented")
}
func (UnimplementedCoreServiceServer) GetHistory(context.Context, *GetHistoryRequest) (*GetHistoryResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetHistory not implemented")
}
func (UnimplementedCoreServiceServer) mustEmbedUnimplementedCoreServiceServer() {}
func (UnimplementedCoreServiceServer) testEmbeddedByValue()                     {}

// 可嵌入 UnsafeCoreServiceServer 以选择退出此服务的向前兼容性。
// 不建议使用此接口，因为向 CoreServiceServer 添加的方法将
// 导致编译错误。
type UnsafeCoreServiceServer interface {
	mustEmbedUnimplementedCoreServiceServer()
}

func RegisterCoreServiceServer(s grpc.ServiceRegistrar, srv CoreServiceServer) {
	// 如果以下调用引发 panic，表示 UnimplementedCoreServiceServer 是通过
	// 指针嵌入且为 nil。这会在调用未实现的方法时引发 panic，
	// 因此我们在初始化时检测以防止后续因 I/O 在运行时发生。
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
// 仅用于 grpc.RegisterService 直接使用，
// 不应内省或修改（即使是复制）
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
	Metadata: "app/core/core.proto",
}
