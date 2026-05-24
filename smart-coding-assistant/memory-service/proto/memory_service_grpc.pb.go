// 由 protoc-gen-go-grpc 自动生成。请勿手动编辑。
// 版本：
// - protoc-gen-go-grpc v1.6.1
// - protoc             v7.34.1
// 源文件：protos/memory_service.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// 这是一个编译时断言，用于确保此生成的文件与正在编译的 grpc 包兼容。
// 要求 gRPC-Go v1.64.0 或更高版本。
const _ = grpc.SupportPackageIsVersion9

const (
	MemoryService_SaveContext_FullMethodName   = "/memory.MemoryService/SaveContext"
	MemoryService_GetContext_FullMethodName    = "/memory.MemoryService/GetContext"
	MemoryService_DeleteContext_FullMethodName = "/memory.MemoryService/DeleteContext"
	MemoryService_UpdateContext_FullMethodName = "/memory.MemoryService/UpdateContext"
	MemoryService_SaveVector_FullMethodName    = "/memory.MemoryService/SaveVector"
	MemoryService_SearchSimilar_FullMethodName = "/memory.MemoryService/SearchSimilar"
	MemoryService_DeleteVector_FullMethodName  = "/memory.MemoryService/DeleteVector"
)

// MemoryServiceClient 是 MemoryService 服务的客户端 API。
//
// 关于 ctx 的使用和关闭/结束流式 RPC 的语义，请参考 https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream。
type MemoryServiceClient interface {
	// 键值上下文（Redis）
	SaveContext(ctx context.Context, in *SaveContextRequest, opts ...grpc.CallOption) (*SaveContextResponse, error)
	GetContext(ctx context.Context, in *GetContextRequest, opts ...grpc.CallOption) (*GetContextResponse, error)
	DeleteContext(ctx context.Context, in *DeleteContextRequest, opts ...grpc.CallOption) (*DeleteContextResponse, error)
	UpdateContext(ctx context.Context, in *UpdateContextRequest, opts ...grpc.CallOption) (*UpdateContextResponse, error)
	// 向量操作（Milvus）
	SaveVector(ctx context.Context, in *SaveVectorRequest, opts ...grpc.CallOption) (*SaveVectorResponse, error)
	SearchSimilar(ctx context.Context, in *SearchSimilarRequest, opts ...grpc.CallOption) (*SearchSimilarResponse, error)
	DeleteVector(ctx context.Context, in *DeleteVectorRequest, opts ...grpc.CallOption) (*DeleteVectorResponse, error)
}

type memoryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMemoryServiceClient(cc grpc.ClientConnInterface) MemoryServiceClient {
	return &memoryServiceClient{cc}
}

func (c *memoryServiceClient) SaveContext(ctx context.Context, in *SaveContextRequest, opts ...grpc.CallOption) (*SaveContextResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SaveContextResponse)
	err := c.cc.Invoke(ctx, MemoryService_SaveContext_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoryServiceClient) GetContext(ctx context.Context, in *GetContextRequest, opts ...grpc.CallOption) (*GetContextResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetContextResponse)
	err := c.cc.Invoke(ctx, MemoryService_GetContext_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoryServiceClient) DeleteContext(ctx context.Context, in *DeleteContextRequest, opts ...grpc.CallOption) (*DeleteContextResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteContextResponse)
	err := c.cc.Invoke(ctx, MemoryService_DeleteContext_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoryServiceClient) UpdateContext(ctx context.Context, in *UpdateContextRequest, opts ...grpc.CallOption) (*UpdateContextResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateContextResponse)
	err := c.cc.Invoke(ctx, MemoryService_UpdateContext_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoryServiceClient) SaveVector(ctx context.Context, in *SaveVectorRequest, opts ...grpc.CallOption) (*SaveVectorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SaveVectorResponse)
	err := c.cc.Invoke(ctx, MemoryService_SaveVector_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoryServiceClient) SearchSimilar(ctx context.Context, in *SearchSimilarRequest, opts ...grpc.CallOption) (*SearchSimilarResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SearchSimilarResponse)
	err := c.cc.Invoke(ctx, MemoryService_SearchSimilar_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoryServiceClient) DeleteVector(ctx context.Context, in *DeleteVectorRequest, opts ...grpc.CallOption) (*DeleteVectorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteVectorResponse)
	err := c.cc.Invoke(ctx, MemoryService_DeleteVector_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MemoryServiceServer 是 MemoryService 服务的服务端 API。
// 所有实现都必须嵌入 UnimplementedMemoryServiceServer
// 以保证向前兼容性。
type MemoryServiceServer interface {
	// 键值上下文（Redis）
	SaveContext(context.Context, *SaveContextRequest) (*SaveContextResponse, error)
	GetContext(context.Context, *GetContextRequest) (*GetContextResponse, error)
	DeleteContext(context.Context, *DeleteContextRequest) (*DeleteContextResponse, error)
	UpdateContext(context.Context, *UpdateContextRequest) (*UpdateContextResponse, error)
	// 向量操作（Milvus）
	SaveVector(context.Context, *SaveVectorRequest) (*SaveVectorResponse, error)
	SearchSimilar(context.Context, *SearchSimilarRequest) (*SearchSimilarResponse, error)
	DeleteVector(context.Context, *DeleteVectorRequest) (*DeleteVectorResponse, error)
	mustEmbedUnimplementedMemoryServiceServer()
}

// UnimplementedMemoryServiceServer 必须被嵌入以保证向前兼容的实现。
//
// 注意：应该以值类型而不是指针类型嵌入，以避免方法调用时出现空指针解引用。
type UnimplementedMemoryServiceServer struct{}

func (UnimplementedMemoryServiceServer) SaveContext(context.Context, *SaveContextRequest) (*SaveContextResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method SaveContext not implemented")
}
func (UnimplementedMemoryServiceServer) GetContext(context.Context, *GetContextRequest) (*GetContextResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetContext not implemented")
}
func (UnimplementedMemoryServiceServer) DeleteContext(context.Context, *DeleteContextRequest) (*DeleteContextResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteContext not implemented")
}
func (UnimplementedMemoryServiceServer) UpdateContext(context.Context, *UpdateContextRequest) (*UpdateContextResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateContext not implemented")
}
func (UnimplementedMemoryServiceServer) SaveVector(context.Context, *SaveVectorRequest) (*SaveVectorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method SaveVector not implemented")
}
func (UnimplementedMemoryServiceServer) SearchSimilar(context.Context, *SearchSimilarRequest) (*SearchSimilarResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method SearchSimilar not implemented")
}
func (UnimplementedMemoryServiceServer) DeleteVector(context.Context, *DeleteVectorRequest) (*DeleteVectorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteVector not implemented")
}
func (UnimplementedMemoryServiceServer) mustEmbedUnimplementedMemoryServiceServer() {}
func (UnimplementedMemoryServiceServer) testEmbeddedByValue()                       {}

// UnsafeMemoryServiceServer 可以嵌入以选择退出此服务的向前兼容性。
// 不推荐使用此接口，因为向 MemoryServiceServer 添加的方法将导致编译错误。
type UnsafeMemoryServiceServer interface {
	mustEmbedUnimplementedMemoryServiceServer()
}

func RegisterMemoryServiceServer(s grpc.ServiceRegistrar, srv MemoryServiceServer) {
	// 如果以下调用发生 panic，则表示 UnimplementedMemoryServiceServer 是通过指针嵌入的且为 nil。
	// 当未实现的方法被调用时将导致 panic，因此我们在初始化时进行测试，
	// 以防止之后因 I/O 在运行时发生此情况。
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MemoryService_ServiceDesc, srv)
}

func _MemoryService_SaveContext_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SaveContextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoryServiceServer).SaveContext(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoryService_SaveContext_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoryServiceServer).SaveContext(ctx, req.(*SaveContextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoryService_GetContext_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetContextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoryServiceServer).GetContext(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoryService_GetContext_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoryServiceServer).GetContext(ctx, req.(*GetContextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoryService_DeleteContext_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteContextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoryServiceServer).DeleteContext(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoryService_DeleteContext_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoryServiceServer).DeleteContext(ctx, req.(*DeleteContextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoryService_UpdateContext_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateContextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoryServiceServer).UpdateContext(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoryService_UpdateContext_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoryServiceServer).UpdateContext(ctx, req.(*UpdateContextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoryService_SaveVector_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SaveVectorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoryServiceServer).SaveVector(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoryService_SaveVector_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoryServiceServer).SaveVector(ctx, req.(*SaveVectorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoryService_SearchSimilar_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchSimilarRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoryServiceServer).SearchSimilar(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoryService_SearchSimilar_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoryServiceServer).SearchSimilar(ctx, req.(*SearchSimilarRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoryService_DeleteVector_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteVectorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoryServiceServer).DeleteVector(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoryService_DeleteVector_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoryServiceServer).DeleteVector(ctx, req.(*DeleteVectorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MemoryService_ServiceDesc 是 MemoryService 服务的 grpc.ServiceDesc。
// 它仅供 grpc.RegisterService 直接使用，不应被内省或修改（即使是作为副本）。
var MemoryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "memory.MemoryService",
	HandlerType: (*MemoryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SaveContext",
			Handler:    _MemoryService_SaveContext_Handler,
		},
		{
			MethodName: "GetContext",
			Handler:    _MemoryService_GetContext_Handler,
		},
		{
			MethodName: "DeleteContext",
			Handler:    _MemoryService_DeleteContext_Handler,
		},
		{
			MethodName: "UpdateContext",
			Handler:    _MemoryService_UpdateContext_Handler,
		},
		{
			MethodName: "SaveVector",
			Handler:    _MemoryService_SaveVector_Handler,
		},
		{
			MethodName: "SearchSimilar",
			Handler:    _MemoryService_SearchSimilar_Handler,
		},
		{
			MethodName: "DeleteVector",
			Handler:    _MemoryService_DeleteVector_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/memory_service.proto",
}
