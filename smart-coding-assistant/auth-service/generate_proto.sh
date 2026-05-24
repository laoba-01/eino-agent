#!/bin/bash

# 生成 auth 服务的 protobuf 代码
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    ../protos/auth_service.proto