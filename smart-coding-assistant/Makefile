PROTO_DIR = protos
CORE_DIR = core-service
TOOL_DIR = tool-service
MEMORY_DIR = memory-service
AUTH_DIR = auth-service

all: generate

generate:
	@echo "Generating gRPC code..."
	@protoc --go_out=$(CORE_DIR) --go-grpc_out=$(CORE_DIR) $(PROTO_DIR)/core_service.proto
	@protoc --go_out=$(TOOL_DIR) --go-grpc_out=$(TOOL_DIR) $(PROTO_DIR)/tool_service.proto
	@protoc --go_out=$(MEMORY_DIR) --go-grpc_out=$(MEMORY_DIR) $(PROTO_DIR)/memory_service.proto
	@protoc --go_out=$(AUTH_DIR) --go-grpc_out=$(AUTH_DIR) $(PROTO_DIR)/auth_service.proto
	@echo "Generated gRPC code successfully!"

# 运行认证服务
run-auth:
	@echo "启动认证服务..."
	@cd $(AUTH_DIR) && go run main.go

# 运行记忆服务
run-memory:
	@echo "启动记忆管理服务..."
	@cd $(MEMORY_DIR) && go run main.go

# 运行工具服务
run-tool:
	@echo "启动工具调用服务..."
	@cd $(TOOL_DIR) && go run main.go

# 运行核心服务
run-core:
	@echo "启动核心学习服务..."
	@cd $(CORE_DIR) && go run main.go

# 运行 API 网关
run-gateway:
	@echo "启动 API 网关..."
	@cd api-gateway && go run main.go

# 运行所有服务
run-all:
	@echo "请在多个终端中分别运行以下命令："
	@echo "  make run-auth    # 启动认证服务"
	@echo "  make run-memory  # 启动记忆管理服务"
	@echo "  make run-tool    # 启动工具调用服务"
	@echo "  make run-core    # 启动核心学习服务"
	@echo "  make run-gateway # 启动 API 网关"

clean:
	@echo "Cleaning generated files..."
	@rm -f $(CORE_DIR)/*_service.pb.go
	@rm -f $(CORE_DIR)/*_service_grpc.pb.go
	@rm -f $(TOOL_DIR)/*_service.pb.go
	@rm -f $(TOOL_DIR)/*_service_grpc.pb.go
	@rm -f $(MEMORY_DIR)/*_service.pb.go
	@rm -f $(MEMORY_DIR)/*_service_grpc.pb.go
	@rm -f $(AUTH_DIR)/*_service.pb.go
	@rm -f $(AUTH_DIR)/*_service_grpc.pb.go
	@echo "Cleaned successfully!"
