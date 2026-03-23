# 智能编程学习助手 Agent

智能编程学习助手 Agent 是一个基于大模型的智能编程学习平台，采用分布式架构设计，为用户提供自然、智能的编程学习体验，同时集成专业的编程工具支持。

## 系统架构

智能编程学习助手 Agent 由以下几个核心组件组成：

```
用户前端 (CLI/简易网页)
    ↓
API 网关 (HTTP 服务)
    ↓
核心学习服务 (RPC)
    ↓ ┌──────────────┐ ↓
    ↓ │              │ ↓
工具调用服务 (RPC)  记忆管理服务 (RPC)  大模型API (字节云)
    ↓ │              │ ↓
    ↓ └──────────────┘ ↓
代码报错解析工具  语法查询工具  刷题思路工具      Redis (上下文存储)
```

### 组件说明

| 组件     | 类型      | 端口    | 功能描述                 |
| ------ | ------- | ----- | -------------------- |
| API 网关 | HTTP 服务 | 8080  | 接收前端请求，转发给核心学习服务     |
| 核心学习服务 | gRPC 服务 | 50051 | 处理学习请求，与大模型 API 交互   |
| 工具调用服务 | gRPC 服务 | 50052 | 提供代码报错解析、语法查询和刷题思路工具 |
| 记忆管理服务 | gRPC 服务 | 50053 | 与 Redis 交互，管理学习上下文   |
| 前端     | 网页应用    | -     | 提供用户交互界面             |

## 技术栈

| 类别      | 技术                  | 版本     | 用途     |
| ------- | ------------------- | ------ | ------ |
| 后端语言    | Go                  | 1.25.6 | 实现所有服务 |
| 通信协议    | gRPC                | -      | 服务间通信  |
| HTTP 框架 | 标准库                 | -      | API 网关 |
| 缓存      | Redis               | -      | 上下文存储  |
| 前端      | HTML/CSS/JavaScript | -      | 用户界面   |

## 核心功能

- **智能对话**：与大模型 API 交互，提供自然、智能的学习对话体验
- **代码报错解析**：分析代码错误，提供错误原因和修复建议
- **语法查询**：提供编程语言语法查询功能，包括解释和示例
- **刷题思路**：为编程问题提供解题思路和代码实现
- **上下文管理**：通过 Redis 存储和管理学习上下文，提供连续学习能力

## 快速开始

### 前置条件

- Go 1.25.6 或更高版本
- Redis 服务（用于记忆管理）
- protoc 编译器（用于生成 gRPC 代码）

### 安装与运行

1. **克隆项目**
2. **安装依赖**
   ```bash
   # 为各个服务安装依赖
   cd core-service && go mod tidy
   cd ../tool-service && go mod tidy
   cd ../memory-service && go mod tidy
   cd ../api-gateway && go mod tidy
   ```
3. **生成 gRPC 代码**
   ```bash
   # 使用 Makefile 生成 gRPC 代码
   make generate
   ```
4. **启动服务**
   - 启动 Redis 服务（如果使用记忆管理功能）
   ```bash
   redis-server
   ```
   - 启动核心学习服务
   ```bash
   cd core-service
   go run main.go
   ```
   - 启动工具调用服务
   ```bash
   cd tool-service
   go run main.go
   ```
   - 启动记忆管理服务
   ```bash
   cd memory-service
   go run main.go
   ```
   - 启动 API 网关
   ```bash
   cd api-gateway
   go run main.go
   ```
5. **打开前端页面**
   在浏览器中打开 `frontend/index.html`

## API 文档

### API 网关 HTTP 接口

#### POST /api/chat

**请求体**：

```json
{
  "user_id": "string",
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

#### GET /health

**响应体**：

```json
{
  "status": "ok",
  "timestamp": "2026-03-14T12:00:00Z"
}
```

## 目录结构

```
smart-coding-assistant/
├── api-gateway/       # API 网关
│   ├── main.go        # 主文件
│   └── go.mod         # Go 模块文件
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
│   └── index.html     # 前端页面
├── protos/            # gRPC 协议文件
│   ├── core_service.proto     # 核心服务协议
│   ├── tool_service.proto     # 工具服务协议
│   └── memory_service.proto   # 记忆服务协议
├── README.md          # 项目说明
├── 技术文档.md         # 技术文档
├── 需求文档.md         # 需求文档
└── Makefile           # 构建脚本
```

## 示例使用

### 代码调试

1. 在前端输入代码和错误信息
2. 系统分析代码错误并提供修复建议

### 语法查询

1. 在前端输入语法查询请求
2. 系统返回语法解释和示例代码

### 刷题辅助

1. 在前端输入编程问题描述
2. 系统提供解题思路和代码实现

### 学习对话

1. 在前端输入学习相关问题
2. 系统与大模型交互并返回响应

## 配置说明

### Redis 配置

记忆管理服务默认连接到 `localhost:6379`，无密码，使用默认 DB。

### 端口配置

| 服务     | 端口    | 配置文件                   |
| ------ | ----- | ---------------------- |
| API 网关 | 8080  | api-gateway/main.go    |
| 核心学习服务 | 50051 | core-service/main.go   |
| 工具调用服务 | 50052 | tool-service/main.go   |
| 记忆管理服务 | 50053 | memory-service/main.go |

## 扩展与维护

### 扩展新工具

1. 在 `tool_service.proto` 中添加新的 RPC 方法
2. 生成新的 gRPC 代码
3. 在 `tool-service/main.go` 中实现新的方法

### 集成新的大模型 API

在 `core-service/main.go` 中修改 `Chat` 方法，添加对新大模型 API 的调用。

## 安全考虑

- 服务间通信使用 gRPC，可考虑添加 TLS 加密
- API 网关可添加认证和授权机制
- 大模型 API 密钥应妥善保管，避免硬编码在代码中

## 未来规划

- 增加用户认证系统
- 实现更复杂的学习管理功能
- 支持多语言和多模态交互
- 优化性能和可扩展性
- 添加更多编程学习工具

## 许可证

MIT

