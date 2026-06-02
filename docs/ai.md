# AI / Agent API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/agents` | GET | 获取当前用户的 Agent 列表 | 是 |
| `/api/agent/:id` | GET | 获取单个 Agent | 是 |
| `/api/create_agent` | POST | 创建 Agent | 是 |
| `/api/update_agent` | POST | 更新 Agent | 是 |
| `/api/agent/:id` | DELETE | 删除 Agent | 是 |
| `/api/invoke_agent` | POST | 非流式调用 AI | 是 |
| `/api/chat` | POST | 流式 AI 聊天（SSE） | 是 |
| `/api/chats` | GET | 获取聊天记录列表 | 是 |
| `/api/chat/:chat_id` | GET | 获取聊天会话详情 | 是 |
| `/api/chat/:chat_id` | DELETE | 删除聊天会话 | 是 |
| `/api/chat/update` | POST | 更新聊天会话标题 | 是 |

---

## Agent 管理

### GET /api/agents

获取当前用户的所有 Agent 列表。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "user_id": 1,
    "agents": [
      {
        "agent_id": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "user_id": 1,
        "name": "助手",
        "model_name": "gpt-4o-mini",
        "provider": "openai",
        "activate": false,
        "temperature": 0.7,
        "thinking": false,
        "prompts": {},
        "memories": {}
      }
    ]
  },
  "msg": "query agents success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_id` | uint | Agent ID |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |
| `user_id` | uint | 所属用户 ID |
| `name` | string | Agent 名称 |
| `model_name` | string | 模型名称 |
| `provider` | string | AI 提供商（`openai`/`ollama`） |
| `activate` | bool | 是否激活 |
| `temperature` | float | 温度参数 |
| `thinking` | bool | 是否启用思考模式（Ollama） |
| `prompts` | object | 自定义提示词配置 |
| `memories` | object | 记忆配置 |

### GET /api/agent/:id

获取单个 Agent 详情。仅创建者可访问。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "agent_id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "user_id": 1,
    "name": "助手",
    "model_name": "gpt-4o-mini",
    "provider": "openai",
    "activate": false,
    "temperature": 0.7,
    "thinking": false,
    "prompts": {},
    "memories": {}
  },
  "msg": "query agent success"
}
```

### POST /api/create_agent

创建新的 Agent。需要登录。

**Request Body**：

```json
{
  "agent_name": "助手",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxx",
  "provider": "openai",
  "model_name": "gpt-4o-mini"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_name` | string | 是 | Agent 名称 |
| `base_url` | string | 是 | AI 服务地址 |
| `api_key` | string | 是 | API Key |
| `provider` | string | 是 | `openai` 或 `ollama` |
| `model_name` | string | 是 | 模型名称 |

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "create agent success"
}
```

### POST /api/update_agent

更新 Agent。仅创建者可操作。全部为可选字段。需要登录。

**Request Body**：

```json
{
  "agent_id": 1,
  "agent_name": "新名称",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-new-xxx",
  "provider": "openai",
  "model_name": "gpt-4o",
  "prompts": {
    "system": "你是一个有帮助的助手"
  },
  "memories": {},
  "activate": true
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_id` | uint | Agent ID（必填） |
| `agent_name` | string | 新名称 |
| `base_url` | string | AI 服务地址 |
| `api_key` | string | API Key |
| `provider` | string | 提供商 |
| `model_name` | string | 模型名称 |
| `prompts` | object | 提示词配置 |
| `memories` | object | 记忆配置 |
| `activate` | bool | 是否激活 |

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "update agent success"
}
```

### DELETE /api/agent/:id

删除 Agent。仅创建者可操作。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "delete agent success"
}
```

---

## AI 调用

### POST /api/invoke_agent

非流式调用 AI（同步返回完整结果）。需要登录。

**Request Body**：

```json
{
  "agent_id": 1,
  "ai_options": {
    "temperature": 0.7,
    "thinking": false,
    "search_internet": false
  },
  "sys_prompt": "你是一个有帮助的助手",
  "usr_prompt": "你好，请介绍一下自己"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_id` | uint | 是 | 使用的 Agent ID |
| `ai_options` | object | 否 | AI 选项 |
| `ai_options.temperature` | float | 否 | 温度参数 |
| `ai_options.thinking` | bool | 否 | 是否启用思考模式 |
| `ai_options.search_internet` | bool | 否 | 是否启用联网搜索 |
| `sys_prompt` | string | 否 | 系统提示词（覆盖 Agent 配置） |
| `usr_prompt` | string | 是 | 用户提示词 |

**Response**：

```json
{
  "code": 0,
  "data": {
    "agent_id": 1,
    "reasoning_content": "",
    "content": "AI 回复内容",
    "status": true,
    "message": "success"
  },
  "msg": "success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_id` | uint | Agent ID |
| `reasoning_content` | string | 思考过程（Thinking 模式） |
| `content` | string | AI 回复内容 |
| `status` | bool | 是否成功 |
| `message` | string | 状态消息 |

---

## 流式聊天

### POST /api/chat

流式 AI 聊天（SSE 流式输出）。需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `chat_id` | string | 否 | 会话 ID。首次不传自动生成，后续传入相同值以维持上下文 |

**Request Body**：

```json
{
  "agent_id": 1,
  "ai_options": {
    "temperature": 0.7,
    "thinking": false,
    "search_internet": true
  },
  "sys_prompt": "你是一个有帮助的助手",
  "usr_prompt": "你好"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_id` | uint | 是 | Agent ID |
| `ai_options` | object | 否 | AI 选项 |
| `sys_prompt` | string | 否 | 系统提示词 |
| `usr_prompt` | string | 是 | 用户提示词 |

**SSE 响应格式**：

首帧（包含历史）：
```json
data: {"chat_id":"uuid","agent_id":1,"user_id":1,"history":[{"role":"user","content":"你好"}]}
```

流式块：
```json
data: {"agent_id":1,"chat_id":"uuid","content":"增量文本","status":true,"reasoning_content":""}
```

完成帧：
```json
data: {"agent_id":1,"chat_id":"uuid","content":"","status":true,"message":"done"}
```

错误帧：
```json
data: {"agent_id":1,"chat_id":"uuid","content":"","status":false,"message":"错误描述"}
```

---

## 会话管理

### GET /api/chats

获取当前用户的聊天记录列表。需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

**Response**：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "uuid": "chat-uuid-string",
        "user_id": 1,
        "agent_id": 1,
        "title": "用户的第一条消息（截断100字）",
        "messages": []
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 15
  },
  "msg": "query chats success"
}
```

### GET /api/chat/:chat_id

获取指定聊天会话详情。仅可访问自己的会话。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "uuid": "chat-uuid-string",
    "user_id": 1,
    "agent_id": 1,
    "title": "对话标题",
    "messages": [
      {
        "role": "user",
        "content": "你好",
        "is_done": true,
        "is_error": false,
        "use_tool": false,
        "time": 1716624000
      },
      {
        "role": "model",
        "content": "你好！有什么可以帮助你的吗？",
        "is_done": true,
        "is_error": false,
        "use_tool": false,
        "time": 1716624001
      }
    ]
  },
  "msg": "query chat session success"
}
```

### DELETE /api/chat/:chat_id

删除指定聊天会话。仅可删除自己的会话。需要登录。

**Request Body**：

```json
{
  "chat_id": "chat-uuid-string"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `chat_id` | string | 是 | 要删除的聊天会话 ID |

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "delete chat session success"
}
```

### POST /api/chat/update

更新聊天会话标题。仅可更新自己的会话。需要登录。

**Request Body**：

```json
{
  "chat_id": "chat-uuid-string",
  "title": "新的对话标题"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `chat_id` | string | 是 | 要更新的聊天会话 ID |
| `title` | string | 是 | 新的会话标题 |

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "update chat session success"
}
```

---

## 前端实现说明

### 流式聊天流程

```
1. 首次：生成 chat_id（crypto.randomUUID()）
2. POST /api/chat?chat_id=xxx 发送请求（保持连接打开）
3. 按行读取 data: ... 事件
4. 首帧：记录 chat_id，保存 history（用于后续请求）
5. 流式块：逐个 append 到消息气泡
6. 完成帧：关闭连接
7. 下次提问：将新的 usr_prompt 一起发送，复用同一 chat_id
```

### 断线重连机制

后端使用 Redis 缓存流式响应内容，支持前端断线重连后继续接收：

1. **流式状态**：`stream:{chat_id}:status` - 值为 `streaming` / `done` / `error`
2. **流式内容**：`stream:{chat_id}:chunks` - List 结构，存储所有已生成的内容块
3. **错误信息**：`stream:{chat_id}:error` - 仅在 status=error 时存在

**重连流程**：
```
1. 前端重新连接 POST /api/chat?chat_id=xxx
2. 后端检查 stream:{chat_id}:status
3. 如果 status=streaming：
   - 从 stream:{chat_id}:chunks 读取已有内容并推送
   - 轮询等待新的 chunks 直到 status 变为 done 或 error
4. 如果 status=error：
   - 从 stream:{chat_id}:error 读取错误信息
   - 推送错误帧到前端
5. 如果 status 不存在：
   - 正常开始新的流式响应
```

**SSE 响应格式**：

首帧（包含历史）：
```json
data: {"chat_id":"uuid","agent_id":1,"user_id":1,"history":[{"role":"user","content":"你好"}]}
```

流式块：
```json
data: {"agent_id":1,"chat_id":"uuid","content":"增量文本","status":true,"reasoning_content":""}
```

完成帧：
```json
data: {"agent_id":1,"chat_id":"uuid","content":"","status":true,"message":"done"}
```

错误帧：
```json
data: {"agent_id":1,"chat_id":"uuid","content":"","status":false,"message":"错误描述"}
```

### 联网搜索

- `ai_options.search_internet: true` 时，AI 会调用搜索 API 获取实时信息
- 搜索工具触发时，SSE 流中会先返回一条 `"content":"正在搜索..."` 的思考状态
