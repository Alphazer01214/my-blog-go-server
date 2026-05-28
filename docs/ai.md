# AI API

| Method | Endpoint              | Description               |
| ------ | --------------------- | ------------------------- |
| POST   | /api/agents           | Create agent              |
| PUT    | /api/agents/:id       | Update agent              |
| GET    | /api/agents           | List agents               |
| GET    | /api/agent/:id        | Get agent detail          |
| POST   | /api/invoke_agent     | Invoke agent (sync)       |
| POST   | /api/chat             | Chat with agent (SSE)     |
| GET    | /api/chats            | List chat sessions        |
| GET    | /api/chat/:session_id | Get chat session          |

---

## POST /api/agents

创建新 Agent。需要登录。

**Request Body:**

```json
{
  "agent_name": "助手",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxx",
  "model_name": "gpt-4o-mini",
  "provider": "openai"
}
```

- `provider`：`openai` 或 `ollama`

**Response:**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "user_id": 1,
    "name": "助手",
    "base_url": "https://api.openai.com/v1",
    "api_key": "sk-***",
    "model_name": "gpt-4o-mini",
    "provider": "openai",
    "activate": true,
    "temperature": 0.7,
    "thinking": false,
    "tools": {},
    "prompts": {},
    "memories": {}
  },
  "msg": "success"
}
```

---

## PUT /api/agents/:id

更新 Agent。需要登录（必须是创建者）。

**Path Parameters:**
- `id` (integer) - Agent ID

**Request Body:** 同 `POST /api/agents`。

**Response:** 同 `POST /api/agents`。

---

## GET /api/agents

获取当前用户的 Agent 列表。需要登录。

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...AgentInfo..." }
  ],
  "msg": "success"
}
```

---

## GET /api/agent/:id

获取指定 Agent 详情。需要登录。

**Path Parameters:**
- `id` (integer) - Agent ID

**Response:**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "助手",
    "model_name": "gpt-4o-mini",
    "provider": "openai",
    "temperature": 0.7,
    "thinking": false,
    "tools": {},
    "prompts": {},
    "memories": {}
  },
  "msg": "success"
}
```

---

## POST /api/invoke_agent

非流式调用 Agent（同步返回完整结果）。需要登录。

**Request Body:**

```json
{
  "agent_id": 1,
  "usr_prompt": "你好",
  "sys_prompt": "你是一个交易助手"
}
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "agent_id": 1,
    "session_id": "uuid",
    "content": "AI 回复内容",
    "status": true,
    "message": ""
  },
  "msg": "success"
}
```

---

## POST /api/chat

流式 AI 聊天（SSE）。需要登录。

`session_id` 通过 **Query 参数** 传递，不是 Request Body。

- 新对话：前端生成 `session_id`（`crypto.randomUUID()`）
- 追问：传入相同的 `session_id` 以延续会话

**Query Parameters:**
- `session_id` (string) - 会话 UUID。不传则后端自动生成。

**Request Body:**

```json
{
  "agent_id": 1,
  "usr_prompt": "什么是止损？",
  "sys_prompt": "你是一个交易助手"
}
```

**SSE 响应：**

首帧（含历史）：

```
data: {"session_id":"uuid","agent_id":1,"user_id":1,"history":[{"role":"user","content":"你好"},{"role":"assistant","content":"你好！"}]}
```

流式块：

```
data: {"agent_id":1,"session_id":"uuid","content":"止损","status":true,"message":""}
data: {"agent_id":1,"session_id":"uuid","content":"是一种","status":true,"message":""}
```

完成帧：

```
data: {"agent_id":1,"session_id":"uuid","content":"","status":true,"message":"done"}
```

错误帧：

```
data: {"agent_id":1,"session_id":"uuid","content":"","status":false,"message":"错误描述"}
```

---

## GET /api/chats

获取当前用户的聊天会话列表。需要登录。

**Query Parameters:**
- `page` (integer, default: 1)
- `page_size` (integer, default: 20)

**Response:**

```json
{
  "code": 0,
  "data": {
    "sessions": [
      {
        "id": 1,
        "uuid": "session-uuid",
        "agent_id": 1,
        "user_id": 1,
        "title": "交易讨论",
        "messages": [],
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 25,
    "page": 1,
    "page_size": 20
  },
  "msg": "success"
}
```

---

## GET /api/chat/:session_id

获取指定聊天会话详情（含消息）。需要登录。

**Path Parameters:**
- `session_id` (string) - 会话 UUID

**Response:**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "uuid": "session-uuid",
    "agent_id": 1,
    "user_id": 1,
    "title": "交易讨论",
    "messages": [
      {
        "role": "user",
        "content": "什么是止损？",
        "is_done": false,
        "is_error": false,
        "use_tool": false,
        "time": 1716624000
      },
      {
        "role": "assistant",
        "content": "止损是一种...",
        "is_done": true,
        "is_error": false,
        "use_tool": false,
        "time": 1716624001
      }
    ],
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  },
  "msg": "success"
}
```
