# 多语言支持 + 代码高亮 实现计划

> **For agentic workers:** 使用 checkboxes 追踪进度。

**目标:** 前端 zh/en 双语切换 + 代码高亮；后端 System Prompt 按请求语言选择。

**架构:** 前端嵌入 i18n 字典 + 语言切换按钮 + highlight.js CDN；Proto ChatRequest 增加 language 字段；后端按语言选择 System Prompt。

**技术栈:** Go 1.25.6, highlight.js, protobuf, Eino ReAct Agent

## 全局约束

- 不改变 gRPC 服务接口签名（只增加 optional 字段）
- 不改变 auth / tool / memory 服务
- 代码高亮为纯前端改动
- 语言默认值 "zh"，向后兼容

---

### Task 1: Proto 增加 language 字段

**文件:**
- 修改: `app/core/core.proto`
- 重新生成: `app/core/pb/core.pb.go`, `app/core/pb/core_grpc.pb.go`

- [ ] **Step 1: proto 增加 language 字段**

`app/core/core.proto` 中 ChatRequest 增加：

```proto
message ChatRequest {
  string user_id = 1;
  string message = 2;
  map<string, string> context = 3;
  string language = 4;   // "zh" | "en", 默认 "zh"
}
```

- [ ] **Step 2: 重新生成 pb 文件**

```bash
cd d:/agent/smart-coding-assistant
protoc --go_out=. --go-grpc_out=. app/core/core.proto
```

注意 protoc 生成路径，确保 pb 文件写入 `app/core/pb/` 目录。

- [ ] **Step 3: 验证编译**

```bash
cd d:/agent/smart-coding-assistant && go build ./app/core/...
```

- [ ] **Step 4: Commit**

```bash
git add app/core/core.proto app/core/pb/core.pb.go app/core/pb/core_grpc.pb.go
git commit -m "feat(core): ChatRequest 增加 language 字段，支持 zh/en"
```

---

### Task 2: Config 支持多语言 SystemPrompt

**文件:**
- 修改: `app/core/internal/config/config.go`
- 修改: `app/core/etc/core.yaml`

- [ ] **Step 1: Config 改为 map 结构**

`app/core/internal/config/config.go` 中 LLM 结构体：

```go
LLM struct {
	Endpoint      string
	APIKey        string
	Model         string
	SystemPrompt map[string]string  // 改为 map: "zh" / "en"
}
```

- [ ] **Step 2: YAML 配置**

`app/core/etc/core.yaml`：

```yaml
LLM:
  Endpoint: "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
  APIKey: ""
  Model: "deepseek-v4-pro"
  SystemPrompt:
    zh: "你是「智能编程学习助手」，一名友好的 AI 编程导师。对于问候、闲聊、自我介绍等问题，自然地回复，不需要调用工具。对于代码报错分析、语法查询、算法解题等编程问题，调用对应工具后给出答案。回复风格：专业但不生硬，适当使用 emoji。"
    en: "You are 'AI Coding Tutor', a friendly programming mentor. For greetings, chitchat, or self-introductions, reply naturally without calling tools. For code error analysis, syntax queries, algorithm problems, call the appropriate tools and provide answers. Be professional yet warm, use emoji appropriately."
```

- [ ] **Step 3: 验证编译**

```bash
cd d:/agent/smart-coding-assistant && go build ./app/core/...
```

- [ ] **Step 4: Commit**

```bash
git add app/core/internal/config/config.go app/core/etc/core.yaml
git commit -m "feat(core): SystemPrompt 改为 zh/en 多语言 map 配置"
```

---

### Task 3: ServiceContext + ChatLogic 按语言选择 SystemPrompt

**文件:**
- 修改: `app/core/internal/svc/servicecontext.go`
- 修改: `app/core/internal/logic/chatlogic.go`

**接口:**
- 消费: `Config.LLM.SystemPrompt` (map)，`ChatRequest.Language` (新增)

- [ ] **Step 1: ServiceContext 存储 SystemPrompt map**

`app/core/internal/svc/servicecontext.go` — `NewServiceContext()` 中，不再在创建 Agent 时传入固定的 SystemPrompt。改为在 ServiceContext 中直接存储 SystemPrompt map，由 ChatLogic 在运行时按请求语言选择。

```go
type ServiceContext struct {
	Config       config.Config
	MCPClient    *mcp.ClientManager
	Embedder     embedding.Embedder
	MemoryRpc    memorypb.MemoryServiceClient
	Agent        *agent.Agent
	SystemPrompt map[string]string  // 新增: 多语言 system prompt
}
```

Agent 创建时不再传入 SystemPrompt（或传入空字符串，待 Task 4 改为运行时注入）：

```go
einoAgent, err := agent.New(context.Background(), agent.Config{
	ChatModel: chatModel,
	MaxSteps:  12,
	// SystemPrompt 不再此处固定，由 ChatLogic 运行时按语言选择
}, mcpClient)
```

存储 SystemPrompt map：

```go
return &ServiceContext{
	Config:       c,
	MCPClient:    mcpClient,
	Embedder:     embedder,
	MemoryRpc:    memoryRpc,
	Agent:        einoAgent,
	SystemPrompt: c.LLM.SystemPrompt,  // map[string]string
}
```

- [ ] **Step 2: ChatLogic 按 language 选择 SystemPrompt 并传给 Agent**

`app/core/internal/logic/chatlogic.go` — `Chat()` 方法中：

```go
func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
	message := in.GetMessage()

	// 按请求语言选择 system prompt
	lang := in.GetLanguage()
	if lang == "" {
		lang = "zh"  // 默认中文
	}
	sysPrompt := l.svcCtx.SystemPrompt[lang]
	if sysPrompt == "" {
		sysPrompt = l.svcCtx.SystemPrompt["zh"]  // 降级
	}
	l.svcCtx.Agent.SetSystemPrompt(sysPrompt)

	// ... 后续流程不变: RAG 召回 → Agent.Run() → 异步存储
}
```

同时降级错误消息按语言返回：

```go
	response, err := l.svcCtx.Agent.Run(l.ctx, agentInput)
	if err != nil {
		if lang == "en" {
			response = fmt.Sprintf("Sorry, an error occurred: %v\n\nPlease try again or describe your problem more specifically.", err)
		} else {
			response = fmt.Sprintf("抱歉，执行过程中出现错误: %v\n\n请稍后重试或更具体地描述你的问题。", err)
		}
	}
```

**注意:** 需要 Agent 增加 `SetSystemPrompt()` 方法（或者改为在 `Run()` 时传入 system prompt）。

实际上更简洁的做法：为 Agent 增加 `SetSystemPrompt(string)` 方法：

`pkg/agent/agent.go`:
```go
func (a *Agent) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
}
```

- [ ] **Step 3: 网关层传递 language**

`app/gateway/internal/logic/chatlogic.go` — `Handle()` 方法中，将前端传来的 `language` 字段传给 CoreRpc：

在请求解析结构体中增加 `Language` 字段：

```go
var req struct {
	Message  string            `json:"message"`
	Context  map[string]string `json:"context"`
	Language string            `json:"language"`  // 新增
}
```

`CoreRpc.Chat()` 调用增加 Language：

```go
resp, err := l.svcCtx.CoreRpc.Chat(r.Context(), &corepb.ChatRequest{
	UserId:   userID,
	Message:  req.Message,
	Context:  req.Context,
	Language: req.Language,  // 新增
})
```

- [ ] **Step 4: 验证编译**

```bash
cd d:/agent/smart-coding-assistant && go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add app/core/internal/svc/servicecontext.go app/core/internal/logic/chatlogic.go app/gateway/internal/logic/chatlogic.go pkg/agent/agent.go
git commit -m "feat(core): 按请求语言选择 SystemPrompt，网关传递 language 字段"
```

---

### Task 4: 前端 i18n + 代码高亮

**文件:**
- 修改: `frontend/index.html`

**说明:** 这是一个较大的前端改动，所有改动集中在一个 HTML 文件中。

- [ ] **Step 1: 添加 i18n 字典、语言切换器、highlight.js、XSS 安全的代码渲染**

改动涉及以下几个部分：

**1) 语言切换按钮** — 在导航栏 `.user-info` 区域添加：

```html
<button class="btn btn-lang" id="langZh" onclick="switchLang('zh')">中文</button>
<button class="btn btn-lang active" id="langEn" onclick="switchLang('en')">English</button>
```

**2) i18n 字典** — 在 `<script>` 顶部定义：

```js
const I18N = {
    zh: {
        title: "智能编程学习助手",
        featureLabels: ["代码调试", "语法查询", "刷题辅助"],
        featureDescs: ["分析错误，提供修复建议", "查询语法，提供示例", "提供思路和代码实现"],
        chatHeader: "💬 开始对话",
        welcomeMsg: "你好！我是智能编程学习助手，有什么可以帮助你的吗？",
        placeholder: "输入你的编程问题...",
        send: "发送",
        logout: "退出",
        errorMsg: "抱歉，系统暂时无法响应，请稍后再试。",
        unauthorized: "未登录，正在跳转...",
    },
    en: {
        title: "AI Coding Tutor",
        featureLabels: ["Debug", "Syntax", "Practice"],
        featureDescs: ["Analyze errors, get fixes", "Query syntax, see examples", "Get solutions and code"],
        chatHeader: "💬 Start Chat",
        welcomeMsg: "Hello! I'm your AI coding tutor. How can I help?",
        placeholder: "Ask a coding question...",
        send: "Send",
        logout: "Logout",
        errorMsg: "Sorry, the system is temporarily unavailable. Please try again later.",
        unauthorized: "Not logged in, redirecting...",
    }
};
```

**3) 语言切换逻辑：**

```js
let currentLang = localStorage.getItem('lang') || (navigator.language.startsWith('zh') ? 'zh' : 'en');

function switchLang(lang) {
    currentLang = lang;
    localStorage.setItem('lang', lang);
    applyI18n();
}

function applyI18n() {
    const t = I18N[currentLang];
    document.documentElement.lang = currentLang === 'zh' ? 'zh-CN' : 'en';
    document.title = t.title;
    document.querySelector('.navbar-brand h1').textContent = t.title;
    document.querySelector('#chatHeader').textContent = t.chatHeader;
    document.querySelector('#messageInput').placeholder = t.placeholder;
    document.querySelector('.send-btn').textContent = t.send;
    document.querySelector('.btn-logout').textContent = t.logout;
    // 功能面板
    const features = document.querySelectorAll('.feature');
    features.forEach((el, i) => {
        el.querySelector('h3').textContent = t.featureLabels[i];
        el.querySelector('p').textContent = t.featureDescs[i];
    });
    // 语言按钮状态
    document.querySelector('#langZh').classList.toggle('active', currentLang === 'zh');
    document.querySelector('#langEn').classList.toggle('active', currentLang === 'en');
    // 更新欢迎消息（如果聊天记录中第一条是欢迎消息）
    const firstMsg = document.querySelector('#chatHistory .message.assistant .message-bubble');
    if (firstMsg && !firstMsg.dataset.userContent) {
        firstMsg.textContent = t.welcomeMsg;
    }
}

window.addEventListener('load', () => { applyI18n(); /* 现有登录检查不变 */ });
```

**4) highlight.js CDN — 在 `<head>` 中添加：**

```html
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-dark.min.css">
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
```

**5) 消息渲染改用 innerHTML + XSS 安全：**

修改 `addMessage()` 函数：

```js
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function renderMarkdown(text) {
    // 先转义 HTML
    let html = escapeHtml(text);
    // 匹配 ```lang\ncode\n``` 代码块
    html = html.replace(/```(\w*)\n([\s\S]*?)```/g, function(match, lang, code) {
        return '<pre><code class="language-' + (lang || 'plaintext') + '">' + escapeHtml(code) + '</code></pre>';
    });
    // 匹配行内代码 `code`
    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');
    return html;
}

function addMessage(role, content) {
    const chatHistory = document.getElementById('chatHistory');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message ' + role;
    
    const bubbleDiv = document.createElement('div');
    bubbleDiv.className = 'message-bubble';
    bubbleDiv.innerHTML = renderMarkdown(content);
    
    messageDiv.appendChild(bubbleDiv);
    chatHistory.appendChild(messageDiv);
    chatHistory.scrollTop = chatHistory.scrollHeight;
    
    // 代码高亮
    bubbleDiv.querySelectorAll('pre code').forEach(block => {
        hljs.highlightElement(block);
    });
}
```

**6) 发送消息时附带 language：**

```js
async function sendMessage() {
    // ... 现有逻辑 ...
    const response = await fetch(API_BASE + '/api/chat', {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({
            message: message,
            context: {},
            language: currentLang  // 新增
        })
    });
    // ... 后续不变 ...
}
```

- [ ] **Step 2: 验证 — 用浏览器打开 frontend/index.html 检查：**
  - 语言切换按钮可见，切换后 UI 文字变化
  - 发送消息后 API 请求包含 `language` 字段
  - 代码块正确渲染高亮

- [ ] **Step 3: Commit**

```bash
git add frontend/index.html
git commit -m "feat(frontend): i18n 中英文切换 + highlight.js 代码高亮"
```

---

### Task 5: 全量编译 + 端到端验证

- [ ] **Step 1: 编译全量**

```bash
cd d:/agent/smart-coding-assistant && go build ./... && go vet ./...
```

- [ ] **Step 2: 确认无残留**

```bash
cd d:/agent/smart-coding-assistant && grep -r "textContent" frontend/ --include="*.html"  # 验证消息渲染已改用 innerHTML
```

- [ ] **Step 3: Commit（如有微调）**

```bash
git add -A && git commit -m "chore: 多语言+代码高亮最终验证"
```
