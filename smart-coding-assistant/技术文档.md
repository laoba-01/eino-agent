# 智能编程学习助手 Agent 技术文档

## 1. 系统架构

智能编程学习助手 Agent 采用分布式架构，由多个独立的服务组成，通过 gRPC 进行通信。系统架构图如下：

```
用户前端 (登录页 + 聊天页)
    ↓
API 网关 (HTTP 服务 + 认证中间件)
    ↓ ┌──────────────┐ ↓
    ↓ │              │ ↓
认证服务 (RPC)  核心学习服务 (RPC)
    ↓              ↓
Redis (用户数据)  ↓ ┌──────────────┐ ↓
                ↓ │              │ ↓
            工具调用服务 (RPC)  记忆管理服务 (RPC)  大模型API (字节云)
                ↓ │              │ ↓
                ↓ └──────────────┘ ↓
            代码报错解析工具  语法查询工具  刷题思路工具      Redis (上下文存储)
```

### 1.1 组件说明

| 组件 | 类型 | 端口 | 功能描述 |
|------|------|------|----------|
| API 网关 | HTTP 服务 | 8080 | 接收前端请求，处理认证，转发给核心学习服务 |
| 核心学习服务 | gRPC 服务 | 50051 | 处理学习请求，与大模型 API 交互 |
| 工具调用服务 | gRPC 服务 | 50052 | 提供代码报错解析、语法查询和刷题思路工具 |
| 记忆管理服务 | gRPC 服务 | 50053 | 与 Redis/Milvus 交互，管理学习上下文和向量检索 |
| 认证服务 | gRPC 服务 | 50054 | 处理用户注册、登录、Token 验证 |
| 前端 | 网页应用 | - | 提供用户交互界面（登录页 + 聊天页） |

## 2. 技术栈

| 类别 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 后端语言 | Go | 1.25.6 | 实现所有服务 |
| 通信协议 | gRPC | - | 服务间通信 |
| HTTP 框架 | 标准库 | - | API 网关 |
| 缓存 | Redis | - | 上下文存储 |
| 向量数据库 | Milvus | 2.x | 向量存储与语义搜索 |
| 前端 | HTML/CSS/JavaScript | - | 用户界面 |

## 3. 核心功能

### 3.1 API 网关

- **HTTP 接口**：
  - `/api/auth/register`：用户注册
  - `/api/auth/login`：用户登录
  - `/api/auth/logout`：用户登出
  - `/api/chat`：处理学习请求（需要认证）
  - `/health`：健康检查
- **请求处理**：解析 JSON 请求，处理认证，转发给核心学习服务
- **响应处理**：将 gRPC 响应转换为 JSON 格式返回给前端
- **认证中间件**：验证 JWT Token，保护需要认证的接口

### 3.2 核心学习服务

- **gRPC 接口**：
  - `Chat`：处理学习请求
  - `GetHistory`：获取学习历史
- **功能**：
  - 与大模型 API 交互
  - 调用工具服务执行特定学习任务
  - 通过记忆管理服务存储和获取上下文

### 3.3 工具调用服务

- **gRPC 接口**：
  - `AnalyzeCodeError`：分析代码错误
  - `QuerySyntax`：查询语法
  - `GenerateProblemSolution`：生成问题解决方案
- **功能**：
  - 提供代码报错解析工具
  - 提供语法查询工具
  - 提供刷题思路工具

### 3.4 认证服务

- **gRPC 接口**：
  - `Register`：用户注册
  - `Login`：用户登录，返回 JWT Token
  - `ValidateToken`：验证 Token 有效性
  - `Logout`：用户登出，撤销 Token
- **功能**：
  - 用户注册和登录管理
  - JWT Token 生成和验证
  - 与 Redis 存储用户凭据和 Token
  - 密码使用 bcrypt 加密存储

### 3.5 记忆管理服务

- **gRPC 接口**：
  - `SaveContext`：保存上下文（Redis）
  - `GetContext`：获取上下文（Redis）
  - `DeleteContext`：删除上下文（Redis）
  - `UpdateContext`：更新上下文（Redis）
  - `SaveVector`：保存向量数据（Milvus）
  - `SearchSimilar`：语义相似度搜索（Milvus）
  - `DeleteVector`：删除向量数据（Milvus）
- **功能**：
  - 与 Redis 交互，存储学习上下文
  - 与 Milvus 交互，存储和检索向量数据
  - 管理用户学习会话信息
  - 支持基于语义相似度的上下文检索

### 3.5 前端

- **功能**：
  - 提供学习界面
  - 发送学习请求到 API 网关
  - 显示学习响应

## 4. 代码结构

```
smart-coding-assistant/
├── api-gateway/       # API 网关
│   ├── main.go        # 主文件
│   └── go.mod         # Go 模块文件
├── auth-service/      # 认证服务
│   ├── main.go        # 主文件
│   ├── go.mod         # Go 模块文件
│   └── generate_proto.sh  # Proto 生成脚本
├── core-service/      # 核心学习服务
│   ├── main.go        # 主文件
│   └── go.mod         # Go 模块文件
├── tool-service/      # 工具调用服务
│   ├── main.go        # 主文件
│   └── go.mod         # Go 模块文件
├── memory-service/    # 记忆管理服务
│   ├── main.go        # 主文件
│   └── go.mod         # Go 模块文件
├── frontend/          # 前端
│   ├── index.html     # 聊天页面
│   └── login.html     # 登录/注册页面
├── protos/            # gRPC 协议文件
│   ├── core_service.proto     # 核心服务协议
│   ├── tool_service.proto     # 工具服务协议
│   ├── memory_service.proto   # 记忆服务协议
│   └── auth_service.proto     # 认证服务协议
└── Makefile           # 构建脚本
```

## 5. 关键 API

### 5.1 API 网关 HTTP 接口

#### POST /api/auth/register

**请求体**：
```json
{
  "username": "string",
  "password": "string",
  "email": "string"
}
```

**响应体**：
```json
{
  "success": true,
  "message": "Registration successful",
  "user_id": "string"
}
```

#### POST /api/auth/login

**请求体**：
```json
{
  "username": "string",
  "password": "string"
}
```

**响应体**：
```json
{
  "success": true,
  "message": "Login successful",
  "token": "jwt_token_string",
  "user_id": "string"
}
```

#### POST /api/auth/logout

**请求头**：
```
Authorization: Bearer <token>
```

**响应体**：
```json
{
  "success": true,
  "message": "Logout successful"
}
```

#### POST /api/chat

**请求头**：
```
Authorization: Bearer <token>
```

**请求体**：
```json
{
  "message": "string",
  "context": {"key": "value"}
}
```

**响应体**：
```json
{
  "response": "string",
  "is_finished": true,
  "context": {"key": "value"}
}
```

> 注意：此接口需要有效的 JWT Token 进行认证

### 5.2 认证服务 gRPC 接口

#### Register

**请求**：
```protobuf
message RegisterRequest {
  string username = 1;
  string password = 2;
  string email = 3;
}
```

**响应**：
```protobuf
message RegisterResponse {
  bool success = 1;
  string message = 2;
  string user_id = 3;
}
```

#### Login

**请求**：
```protobuf
message LoginRequest {
  string username = 1;
  string password = 2;
}
```

**响应**：
```protobuf
message LoginResponse {
  bool success = 1;
  string message = 2;
  string token = 3;
  string user_id = 4;
}
```

### 5.3 核心学习服务 gRPC 接口

#### Chat

**请求**：
```protobuf
message ChatRequest {
  string user_id = 1;
  string message = 2;
  map<string, string> context = 3;
}
```

**响应**：
```protobuf
message ChatResponse {
  string response = 1;
  bool is_finished = 2;
  map<string, string> context = 3;
}
```

### 5.3 工具调用服务 gRPC 接口

#### AnalyzeCodeError

**请求**：
```protobuf
message AnalyzeCodeErrorRequest {
  string code = 1;
  string error_message = 2;
  string language = 3;
}
```

**响应**：
```protobuf
message AnalyzeCodeErrorResponse {
  string analysis = 1;
  string suggested_fix = 2;
  bool success = 3;
}
```

### 5.4 记忆管理服务 gRPC 接口

#### SaveContext

**请求**：
```protobuf
message SaveContextRequest {
  string user_id = 1;
  map<string, string> context = 2;
  int64 ttl = 3; // Time to live in seconds
}
```

**响应**：
```protobuf
message SaveContextResponse {
  bool success = 1;
  string error = 2;
}
```

#### SaveVector（Milvus 向量存储）

**请求**：
```protobuf
message SaveVectorRequest {
  string collection = 1;           // Milvus 集合名称
  repeated VectorData vectors = 2; // 向量数据列表
}

message VectorData {
  int64 id = 1;                    // 唯一向量ID
  repeated float vector = 2;       // 嵌入向量
  map<string, string> metadata = 3; // 关联元数据
}
```

**响应**：
```protobuf
message SaveVectorResponse {
  bool success = 1;
  string error = 2;
  repeated int64 inserted_ids = 3; // 已插入的向量ID列表
}
```

#### SearchSimilar（语义相似度搜索）

**请求**：
```protobuf
message SearchSimilarRequest {
  string collection = 1;           // Milvus 集合名称
  repeated float query_vector = 2; // 查询向量
  int32 top_k = 3;                 // 返回结果数量
  map<string, string> filter = 4;  // 元数据过滤条件
}
```

**响应**：
```protobuf
message SearchSimilarResponse {
  bool success = 1;
  string error = 2;
  repeated SearchResult results = 3;
}

message SearchResult {
  int64 id = 1;
  repeated float vector = 2;
  map<string, string> metadata = 3;
  float score = 4;                 // 相似度分数
}
```

#### DeleteVector（删除向量）

**请求**：
```protobuf
message DeleteVectorRequest {
  string collection = 1;    // Milvus 集合名称
  repeated int64 ids = 2;   // 要删除的向量ID列表
}
```

**响应**：
```protobuf
message DeleteVectorResponse {
  bool success = 1;
  string error = 2;
}
```

## 6. 部署与运行

### 6.1 前置条件

- Go 1.25.6 或更高版本
- Redis 服务（用于用户存储和记忆管理）
- Milvus 服务（用于向量存储与语义搜索）
- protoc 编译器（用于生成 gRPC 代码）

### 6.2 运行步骤

1. **启动 Redis 服务**：
   ```bash
   redis-server
   ```

2. **启动 Milvus 服务**：
   ```bash
   # 使用 Docker 启动 Milvus Standalone
   docker run -d --name milvus-standalone \
     -p 19530:19530 \
     -p 9091:9091 \
     -v milvus_data:/var/lib/milvus \
     milvusdb/milvus:v2.4-latest \
     milvus run standalone
   ```

3. **启动认证服务**：
   ```bash
   cd auth-service
   go run main.go
   ```

4. **启动核心学习服务**：
   ```bash
   cd core-service
   go run main.go
   ```

5. **启动工具调用服务**：
   ```bash
   cd tool-service
   go run main.go
   ```

6. **启动记忆管理服务**：
   ```bash
   cd memory-service
   go run main.go
   ```

7. **启动 API 网关**：
   ```bash
   cd api-gateway
   go run main.go
   ```

8. **打开前端页面**：
   在浏览器中打开 `frontend/login.html` 进行登录/注册，登录后跳转到聊天页面

## 7. 扩展与维护

### 7.1 扩展新工具

1. 在 `tool_service.proto` 中添加新的 RPC 方法
2. 生成新的 gRPC 代码
3. 在 `tool-service/main.go` 中实现新的方法

### 7.2 集成新的大模型 API

在 `core-service/main.go` 中修改 `Chat` 方法，添加对新大模型 API 的调用。

### 7.3 监控与日志

- 各服务均使用标准库的 `log` 包记录日志
- 可根据需要集成第三方监控工具

## 8. 安全考虑

- 服务间通信使用 gRPC，可考虑添加 TLS 加密
- API 网关使用 JWT Token 进行认证，保护受保护接口
- 用户密码使用 bcrypt 加密存储
- 大模型 API 密钥应妥善保管，避免硬编码在代码中
- JWT Secret 应在生产环境中使用环境变量配置，不应硬编码

## 9. 未来规划

- 实现更复杂的学习管理功能
- 支持多语言和多模态交互
- 优化性能和可扩展性
- 添加更多编程学习工具
- 增加用户学习进度统计和历史记录
