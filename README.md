# Eino Agent — 智能编程学习助手

基于 **Eino ReAct Agent** + **go-zero 微服务** 的 AI 编程助手，具备自主规划、MCP 工具调用和语义记忆（RAG）能力。

## 架构

```
用户前端 (Nginx :80)
    │
    ▼
API Gateway (:8080, go-zero REST)
    │  JWT 中间件
    ├──────────────┬──────────────┬──────────────┐
    ▼              ▼              ▼              ▼
Auth (:50054)  Core (:50051)  Tool (:50052)  Memory (:50053)
gRPC           gRPC           gRPC+MCP       gRPC
    │              │              │              │
  Redis         ┌─Eino Agent    LLM API      Milvus
                │  ├ ChatModel  (DeepSeek)    Redis
                │  ├ MCP Tools
                │  └ Embedder
                │
                └─RAG Pipeline
                    Embedding → Milvus Search → Context Merge
```

## 核心技术

| 类别 | 技术 | 用途 |
|------|------|------|
| Agent 框架 | [Eino](https://github.com/cloudwego/eino) (字节跳动) | ReAct Agent 编排、ChatModel、Embedder |
| 微服务 | [go-zero](https://github.com/zeromicro/go-zero) | RPC 框架、服务发现(etcd)、REST 网关 |
| 协议 | MCP (Model Context Protocol) | AI 工具标准化接入 |
| LLM | DeepSeek (deepseek-chat) | 对话推理 + Tool Calling |
| Embedding | 智谱 GLM (embedding-2, 1024维) | 文本向量化 |
| 向量库 | Milvus 2.4 | ANN 语义检索、AUTOINDEX + L2 |
| 缓存 | Redis 7 | 用户信息、Token、会话上下文 |
| 容器化 | Docker Compose | 全栈一键部署（11 容器） |

## 快速开始

```bash
# 1. 设置环境变量
export DEEPSEEK_API_KEY=sk-xxx
export EMBEDDING_API_KEY=xxx

# 2. 一键启动
docker-compose up --build

# 3. 打开前端
open http://localhost
```

## 手动运行

```bash
# 编译所有服务
make build

# 分别在 5 个终端中启动
make run-auth      # 认证服务 :50054
make run-memory    # 记忆服务 :50053
make run-tool      # 工具服务 :50052 + MCP :8081
make run-core      # 核心 Agent :50051
make run-gateway   # API 网关 :8080
```

## 代码结构

```
smart-coding-assistant/
├── app/
│   ├── gateway/         # API 网关 (REST + JWT)
│   │   ├── internal/
│   │   │   ├── handler/     # 路由注册
│   │   │   ├── logic/       # 业务处理
│   │   │   └── middleware/  # CORS + Auth
│   │   └── etc/gateway.yaml
│   ├── auth/            # 认证服务 (gRPC)
│   │   ├── internal/logic/  # 注册/登录/Token
│   │   └── etc/auth.yaml
│   ├── core/            # 核心 Agent 服务 (gRPC)
│   │   ├── internal/
│   │   │   ├── logic/       # ChatLogic (Agent + RAG)
│   │   │   └── svc/         # ServiceContext (Eino注入)
│   │   └── etc/core.yaml
│   ├── tool/            # 工具服务 (gRPC + MCP Server)
│   │   ├── internal/
│   │   │   ├── mcp/         # MCP Server + 工具注册
│   │   │   └── logic/       # LLM 工具实现
│   │   └── etc/tool.yaml
│   └── memory/          # 记忆服务 (gRPC)
│       ├── internal/logic/  # Milvus + Redis 操作
│       └── etc/memory.yaml
├── pkg/
│   ├── agent/           # Eino Agent 封装
│   │   ├── agent.go         # ReAct Agent 创建与运行
│   │   └── tools.go         # MCP → Eino Tool 适配器
│   ├── llm/             # LLM HTTP 客户端 (Tool 服务用)
│   └── mcp/             # MCP ClientManager (共享)
├── protos/              # Protobuf 协议定义
│   ├── auth_service.proto
│   ├── core_service.proto
│   ├── tool_service.proto
│   └── memory_service.proto
├── frontend/            # 前端页面
├── docker-compose.yml   # 容器编排
├── Dockerfile           # 多阶段构建
└── Makefile             # 构建脚本
```

## Agent 工作流

```
用户输入
  ├─ 闲聊快速路径 (hi/hello/谢谢) → 直接回复
  └─ Agent 主流程:
       │
       ├─ ① RAG 语义召回
       │    Embedder.EmbedStrings(msg) → MemoryRpc.SearchSimilar(TopK=3)
       │    → 格式化历史 → 合并到 Agent 输入
       │
       ├─ ② Eino ReAct Agent 循环
       │    ChatModel 推理 → ToolCall 决策
       │    → MCP 工具调用 → 结果回传 → 再推理
       │    → ... → 最终答案
       │
       └─ ③ 异步回存
            Embedder.EmbedStrings(msg) → MemoryRpc.SaveVector()
```

## MCP 工具

| 工具 | 描述 |
|------|------|
| `analyze_code_error` | 分析代码错误并给出修复建议（传入源码+错误信息+语言） |
| `query_syntax` | 查询编程语言语法概念（传入语言+查询词+上下文） |
| `generate_problem_solution` | 为编程问题生成解题方案（传入问题+难度+语言） |

工具通过 **Streamable HTTP Transport** 在 `:8081/mcp` 暴露，Agent 通过 `ClientManager` 动态发现并适配为 Eino Tool。

## API

### POST /api/chat

```json
// 请求
{ "message": "我的 Go 代码报 panic: nil pointer...", "context": {} }

// 响应 (Agent 自主规划多步执行)
{
  "response": "## 分析结果\n\n✅ **步骤1**: 定位错误...\n✅ **步骤2**: 修复建议...",
  "is_finished": true
}
```

### 完整接口

| 方法 | 路径 | 认证 |
|------|------|------|
| POST | `/api/auth/register` | 无 |
| POST | `/api/auth/login` | 无 |
| POST | `/api/auth/logout` | JWT |
| POST | `/api/chat` | JWT |
| GET | `/health` | 无 |

## 与旧架构对比

| | 旧 (main 分支) | 新 (Eino + go-zero) |
|---|---|---|
| Agent | 手写 Planner 200行 + Executor 160行 | Eino ReAct Agent 65行 |
| LLM | 手写 HTTP Client 100行 | Eino ChatModel |
| Embedding | 手写 HTTP 150行 | Eino Embedder |
| MCP | 手写 SDK 封装 | Eino Tool 适配器 130行 |
| 总代码量 | ~1320行 | ~720行 (-43%) |

## License

MIT
