# AI 智能体模块 API 文档

## 基础信息

- Base URL: `http://localhost:2333/api`
- 数据格式: `application/json`
- 通用响应格式同用户模块（`code: 0` 成功，`code: 7` 失败）
- **所有接口均需 Cookie 认证**（`access-token` + `refresh-token`）

## 数据模型

### Agent 智能体

```json
{
  "id": 1,
  "created_at": "2026-05-12T10:00:00Z",
  "updated_at": "2026-05-12T10:00:00Z",
  "user_id": 1,
  "name": "我的助手",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxx",
  "model_name": "gpt-4",
  "provider": "openai",
  "activate": true,
  "temperature": 0.7,
  "thinking": false,
  "prompts": {},
  "memories": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 智能体 ID |
| `name` | string | 名称 |
| `provider` | string | 提供商：`openai` / `ollama` |
| `base_url` | string | API 地址 |
| `api_key` | string | API 密钥 |
| `model_name` | string | 模型名（如 `gpt-4`, `qwen2.5`） |
| `activate` | bool | 是否激活 |
| `temperature` | float | 温度参数（实际作用于 LLM 请求） |
| `thinking` | bool | 是否启用思考（ollama） |
| `prompts` | object | 提示词配置（JSON） |
| `memories` | object | 记忆配置（JSON） |

### StandardAiResponse

```json
{
  "agent_id": 1,
  "chat_id": "uuid-string",
  "content": "AI 返回的内容",
  "reasoning_content": "思考链内容",
  "status": true,
  "message": "success"
}
```

---

### ChatSession 会话（PostgreSQL 持久化）

```json
{
  "id": 1,
  "uuid": "a1b2c3d4-...",
  "user_id": 1,
  "agent_id": 1,
  "title": "你好",
  "messages": [
    {"role": "user", "content": "你好", "time": 1715500000},
    {"role": "model", "content": "你好！", "time": 1715500001}
  ],
  "created_at": "2026-05-12T10:00:00Z",
  "updated_at": "2026-05-12T10:00:01Z"
}
```

> 会话存储在 PostgreSQL，刷新页面后历史记录不丢失。`uuid` 即 `chat_id`，唯一且按用户隔离。

---

## 1) 创建智能体

`POST /api/create_agent`

**Request Body:**

```json
{
  "agent_name": "我的助手",
  "provider": "openai",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxxxxxxx",
  "model_name": "gpt-4"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_name` | string | 是 | 智能体名称 |
| `provider` | string | 是 | 提供商（`openai` / `ollama`） |
| `base_url` | string | 是 | API 地址 |
| `api_key` | string | 是 | API 密钥 |
| `model_name` | string | 是 | 模型名称 |

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "create agent success"
}
```

---

## 2) 更新智能体

`POST /api/update_agent`

**Request Body:**

```json
{
  "agent_id": 1,
  "agent_name": "新名称",
  "provider": "ollama",
  "base_url": "http://localhost:11434",
  "api_key": "",
  "model_name": "qwen2.5",
  "activate": true,
  "prompts": {},
  "memories": {}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_id` | uint | 是 | 智能体 ID |
| `agent_name` | string | 否 | 新名称 |
| `provider` | string | 否 | 新提供商 |
| `base_url` | string | 否 | 新 API 地址 |
| `api_key` | string | 否 | 新 API 密钥 |
| `model_name` | string | 否 | 新模型名 |
| `activate` | bool | 否 | 是否激活 |
| `prompts` | object | 否 | 提示词配置 |
| `memories` | object | 否 | 记忆配置 |

> 权限：仅智能体创建者可更新。

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "update agent success"
}
```

---

## 3) 获取智能体列表

`GET /api/agents` 或 `GET /api/agent/list`

**Query Params:** 无

> 返回当前用户创建的所有智能体，按创建时间倒序。

**Response:**

```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "created_at": "2026-05-12T10:00:00Z",
      "updated_at": "2026-05-12T10:00:00Z",
      "user_id": 1,
      "name": "我的助手",
      "base_url": "https://api.openai.com/v1",
      "api_key": "sk-xxx",
      "model_name": "gpt-4",
      "provider": "openai",
      "activate": true,
      "temperature": 0,
      "thinking": false,
      "prompts": {},
      "memories": {}
    }
  ],
  "msg": "query agents success"
}
```

---

## 4) 获取智能体详情

`GET /api/agent/:id`

**Path Params:**

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 智能体 ID |

> 权限：仅创建者可查看。

**Response:** 同智能体数据模型，包在 data 中。

---

## 5) 调用智能体（非流式）

`POST /api/invoke_agent`

**Request Body:**

```json
{
  "agent_id": 1,
  "sys_prompt": "你是助手的角色设定",
  "usr_prompt": "用户的问题",
  "ai_options": {
    "temperature": 0.7,
    "thinking": false,
    "search_internet": false
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_id` | uint | 是 | 智能体 ID |
| `usr_prompt` | string | 是 | 用户输入 |
| `sys_prompt` | string | 否 | 系统提示词 |
| `ai_options` | object | 否 | 推理选项 |

**Response:**

```json
{
  "code": 0,
  "data": {
    "agent_id": 1,
    "content": "AI 回答的内容",
    "status": true,
    "message": "success"
  },
  "msg": "success"
}
```

---

## 6) 流式对话（SSE）

`POST /api/chat` — 使用 **Server-Sent Events (SSE)** 协议

**Query Params:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `chat_id` | string | 否 | 会话 ID。不传则服务端自动生成。传已有 ID 可继续历史对话。 |

**Request Body:**

```json
{
  "agent_id": 1,
  "usr_prompt": "用户的问题",
  "sys_prompt": "你是助手的角色设定",
  "ai_options": {}
}
```

字段含义同 `/api/invoke_agent`。

### SSE 消息格式

每个事件为 `data: <json>\n\n` 格式：

**① 会话开始消息（第一条）：**

```json
{
  "chat_id": "a1b2c3d4-...",
  "agent_id": 1,
  "user_id": 1,
  "history": [
    {
      "role": "user",
      "content": "你好"
    },
    {
      "role": "assistant",
      "content": "你好！有什么可以帮你的？"
    }
  ]
}
```

- 如果 `chat_id` 对应已有会话，`history` 会返回该会话的完整历史消息
- 如果是新会话，`history` 为空数组
- 前端应保存 `chat_id`，用于后续继续对话

**② 流式内容消息（中间多条）：**

```json
{
  "agent_id": 1,
  "chat_id": "a1b2c3d4-...",
  "content": "逐字返回的文本片段",
  "status": true,
  "message": ""
}
```

将每条消息的 `content` 拼接即得完整回复。

**③ 结束消息（最后一条）：**

```json
{
  "agent_id": 1,
  "chat_id": "a1b2c3d4-...",
  "content": "完整回复内容",
  "status": true,
  "message": "done"
}
```

失败时：

```json
{
  "agent_id": 1,
  "chat_id": "a1b2c3d4-...",
  "content": "",
  "status": false,
  "message": "错误信息"
}
```

### 前端接入示例

```javascript
// 流式对话
const chatId = localStorage.getItem('chatId') || '';

const response = await fetch('/api/chat?chat_id=' + chatId, {
  method: 'POST',
  credentials: 'include',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ agent_id: 1, usr_prompt: '你好' })
});

const reader = response.body.getReader();
const decoder = new TextDecoder();
let buffer = '';

while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  buffer += decoder.decode(value, { stream: true });
  const lines = buffer.split('\n\n');
  buffer = lines.pop() || '';
  for (const line of lines) {
    if (line.startsWith('data: ')) {
      const data = JSON.parse(line.slice(6));
      if (data.message === 'done') {
        // 会话结束，保存 chat_id
        localStorage.setItem('chatId', data.chat_id);
      } else if (data.status) {
        // 拼接 content 显示
        console.log('收到:', data.content);
      }
    }
  }
}
```

## 7) 获取会话列表

`GET /api/chats` — 需要 Cookie

**Query Params:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码（默认 `1`） |
| `page_size` | int | 否 | 每页数量（默认 `20`，最大 `100`） |

> 返回当前用户的所有会话，按更新时间倒序。`title` 由首条用户消息自动生成（截取前100字符）。

**Response:**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "uuid": "a1b2c3d4-...",
        "user_id": 1,
        "agent_id": 1,
        "title": "你好",
        "messages": [{"role":"user","content":"你好","time":1715500000}, ...],
        "created_at": "2026-05-12T10:00:00Z",
        "updated_at": "2026-05-12T10:00:01Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 1
  },
  "msg": "query chats success"
}
```

---

## 8) 获取单个会话详情

`GET /api/chat/:chat_id` — 需要 Cookie

**Path Params:**

| 参数 | 类型 | 说明 |
|------|------|------|
| `chat_id` | string | 会话 UUID |

> 返回单个会话的完整信息，包括所有历史消息。

---

### 注意事项

- 会话存储在 PostgreSQL 中，刷新页面后历史记录不丢失
- 连接中断时，用之前保存的 `chat_id` 重新请求可继续对话（服务端保留历史消息）
- `chat_id` 由前端生成 UUID 或由服务端自动生成均可
- 不支持跨智能体共享会话
- `temperature` 和 `thinking` 在创建/更新 Agent 时配置，调用时从 Agent 配置读取
