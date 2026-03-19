PROTO_DIR = protos
CORE_DIR = core-service
TOOL_DIR = tool-service
MEMORY_DIR = memory-service

all: generate

generate:
	@echo "Generating gRPC code..."
	@protoc --go_out=$(CORE_DIR) --go-grpc_out=$(CORE_DIR) $(PROTO_DIR)/core_service.proto
	@protoc --go_out=$(TOOL_DIR) --go-grpc_out=$(TOOL_DIR) $(PROTO_DIR)/tool_service.proto
	@protoc --go_out=$(MEMORY_DIR) --go-grpc_out=$(MEMORY_DIR) $(PROTO_DIR)/memory_service.proto
	@echo "Generated gRPC code successfully!"

clean:
	@echo "Cleaning generated files..."
	@rm -f $(CORE_DIR)/*_service.pb.go
	@rm -f $(CORE_DIR)/*_service_grpc.pb.go
	@rm -f $(TOOL_DIR)/*_service.pb.go
	@rm -f $(TOOL_DIR)/*_service_grpc.pb.go
	@rm -f $(MEMORY_DIR)/*_service.pb.go
	@rm -f $(MEMORY_DIR)/*_service_grpc.pb.go
	@echo "Cleaned successfully!"
