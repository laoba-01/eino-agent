PROTO_DIR = protos

.PHONY: all generate tidy build run-%

all: tidy build

# 依赖管理
tidy:
	@echo "整理 Go 依赖..."
	GONOSUMCHECK=* GONOSUMDB=* go mod tidy
	@echo "依赖整理完成!"

# 代码生成
generate:
	@echo "重新生成 protobuf 代码..."
	@protoc --go_out=. --go-grpc_out=. app/auth/auth.proto app/memory/memory.proto app/tool/tool.proto app/core/core.proto
	@rm -rf smart-coding-assistant
	@echo "代码生成完成!"

# 编译所有服务
build:
	@echo "编译所有服务..."
	go build -o bin/auth-rpc ./app/auth
	go build -o bin/memory-rpc ./app/memory
	go build -o bin/tool-rpc ./app/tool
	go build -o bin/core-rpc ./app/core
	go build -o bin/api-gateway ./app/gateway
	@echo "编译完成! 二进制文件在 bin/ 目录"

# 运行认证服务
run-auth:
	@echo "启动认证 RPC 服务..."
	@cd app/auth && go run . -f etc/auth.yaml

# 运行记忆服务
run-memory:
	@echo "启动记忆 RPC 服务..."
	@cd app/memory && go run . -f etc/memory.yaml

# 运行工具服务
run-tool:
	@echo "启动工具 RPC 服务..."
	@cd app/tool && go run . -f etc/tool.yaml

# 运行核心服务
run-core:
	@echo "启动核心 RPC 服务..."
	@cd app/core && go run . -f etc/core.yaml

# 运行 API 网关
run-gateway:
	@echo "启动 API 网关..."
	@cd app/gateway && go run . -f etc/gateway.yaml

# 运行所有服务
run-all:
	@echo "请在多个终端中分别运行以下命令："
	@echo "  make run-auth     # 启动认证 RPC 服务"
	@echo "  make run-memory   # 启动记忆 RPC 服务"
	@echo "  make run-tool     # 启动工具 RPC 服务"
	@echo "  make run-core     # 启动核心 RPC 服务"
	@echo "  make run-gateway  # 启动 API 网关"
	@echo ""
	@echo "或使用 Docker Compose："
	@echo "  docker-compose up --build"

clean:
	@echo "清理编译产物和生成代码..."
	@rm -rf bin/
	@rm -f app/auth/pb/*.pb.go app/memory/pb/*.pb.go app/tool/pb/*.pb.go app/core/pb/*.pb.go
	@echo "清理完成!"
