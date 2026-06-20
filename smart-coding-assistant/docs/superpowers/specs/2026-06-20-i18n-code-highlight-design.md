# 多语言支持 + 代码高亮 设计文档

**日期:** 2026-06-20
**状态:** 设计中

## 1. 背景

当前系统前端全中文、System Prompt 固定中文、回复无代码语法着色。需要支持中英文切换和代码高亮渲染。

## 2. 目标

- 前端 zh/en 双语切换，浏览器语言自动检测
- 后端 System Prompt 按请求语言选择
- 回复中代码块语法高亮（highlight.js）

## 3. 设计方案

### 3.1 前端 i18n (`frontend/index.html`)

嵌入 i18n 文本字典，按 `zh`/`en` 组织所有 UI 字符串：

```js
const I18N = {
  zh: {
    title: "智能编程学习助手",
    features: ["代码调试", "语法查询", "刷题辅助"],
    featuresDesc: ["分析错误，提供修复建议", "查询语法，提供示例", "提供思路和代码实现"],
    chatHeader: "💬 开始对话",
    welcomeMsg: "你好！我是智能编程学习助手，有什么可以帮助你的吗？",
    placeholder: "输入你的编程问题...",
    send: "发送",
    logout: "退出",
    errorMsg: "抱歉，系统暂时无法响应，请稍后再试。",
  },
  en: { /* 对应英文 */ }
}
```

- 顶部导航栏加语言切换按钮（`中文`|`English`）
- 默认从 `navigator.language` 检测，存入 `localStorage('lang')`
- 切换时刷新所有 UI 文本，更新 `<html lang>`
- 发送消息时附带 `language` 字段

### 3.2 Proto 扩展 (`app/core/core.proto`)

`ChatRequest` 增加字段：
```proto
message ChatRequest {
  string user_id = 1;
  string message = 2;
  map<string, string> context = 3;
  string language = 4;   // "zh" | "en"，默认 "zh"
}
```

### 3.3 后端 System Prompt 按语言选择

**配置 (`core.yaml`):**

```yaml
LLM:
  SystemPrompt:
    zh: "你是「智能编程学习助手」..."
    en: "You are 'AI Coding Tutor', a friendly programming mentor..."
```

**ServiceContext:** 按 `ChatRequest.language` 选择对应 prompt 注入 Agent。

**降级错误消息**也按语言返回。

### 3.4 代码高亮

- 引入 highlight.js CDN + `atom-one-dark` 主题
- 消息渲染改用 `innerHTML`：
  - 先 HTML-转义纯文本
  - 正则匹配 ` ``` ` 代码块，替换为 `<pre><code>` 标签
  - 调用 `hljs.highlightElement()` 着色
- XSS 防护：所有非代码块内容先 `escapeHtml()`

## 4. 受影响文件

| 文件 | 改动类型 | 改动描述 |
|------|----------|----------|
| `frontend/index.html` | 修改 | i18n 字典、语言切换器、highlight.js、消息渲染 |
| `app/core/core.proto` | 修改 | ChatRequest 增加 language 字段 |
| `app/core/pb/core.pb.go` | 重新生成 | proto 生成 |
| `app/core/pb/core_grpc.pb.go` | 重新生成 | proto 生成 |
| `app/core/etc/core.yaml` | 修改 | SystemPrompt 改为 zh/en 键值 |
| `app/core/internal/config/config.go` | 修改 | LLM.SystemPrompt 改为 map |
| `app/core/internal/svc/servicecontext.go` | 修改 | 按语言选择 SystemPrompt |
| `app/core/internal/logic/chatlogic.go` | 修改 | Chat() 读取 language 字段 |
| `app/gateway/internal/logic/chatlogic.go` | 修改 | 传递 language 字段 |

## 5. 不变部分

- Agent 推理流程不变
- MCP 工具调用不变
- RAG 语义记忆不变
- 其他服务 (auth/tool/memory) 完全不受影响
