# AI API

| Method | Endpoint            | Description               |
| ------ | ------------------- | ------------------------- |
| POST   | /api/agents         | Create agent              |
| PUT    | /api/agents/:id     | Update agent              |
| GET    | /api/agents         | List agents               |
| GET    | /api/agent/:id      | Get agent detail          |
| POST   | /api/invoke_agent   | Invoke agent (sync)       |
| POST   | /api/chat           | Chat with agent (SSE)     |
| GET    | /api/chats          | List chat sessions        |
| GET    | /api/chat/:chatId   | Get chat session          |

---

## POST /api/agents

Create a new AI agent. Requires authentication.

**Request Body:**

```json
{
  "name": "string (required)",
  "base_url": "string (required)",
  "api_key": "string (required)",
  "model_name": "string (required)",
  "provider": "string (required)",
  "temperature": 0.7,
  "thinking": false,
  "tools": [],
  "prompts": [],
  "memories": []
}
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "user_id": 1,
    "name": "Trading Assistant",
    "base_url": "https://api.openai.com/v1",
    "api_key": "sk-***",
    "model_name": "gpt-4",
    "provider": "openai",
    "activate": true,
    "temperature": 0.7,
    "thinking": false,
    "tools": [],
    "prompts": [],
    "memories": []
  },
  "msg": "success"
}
```

---

## PUT /api/agents/:id

Update an existing agent. Requires authentication (must be owner).

**Path Parameters:**
- `id` (integer) - Agent ID

**Request Body:** Same as `POST /api/agents`.

**Response:** Same as `POST /api/agents`.

---

## GET /api/agents

List all agents belonging to the current user. Requires authentication.

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

Get a specific agent by ID.

**Path Parameters:**
- `id` (integer) - Agent ID

**Response:**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "user_id": 1,
    "name": "Trading Assistant",
    "base_url": "https://api.openai.com/v1",
    "api_key": "sk-***",
    "model_name": "gpt-4",
    "provider": "openai",
    "activate": true,
    "temperature": 0.7,
    "thinking": false,
    "tools": [],
    "prompts": [],
    "memories": []
  },
  "msg": "success"
}
```

---

## POST /api/invoke_agent

Invoke an agent synchronously. Requires authentication.

**Request Body:**

```json
{
  "agent_id": 1,
  "usr_prompt": "string (required)",
  "sys_prompt": "string (optional)"
}
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "agent_id": 1,
    "content": "The agent's response text...",
    "status": "success",
    "message": ""
  },
  "msg": "success"
}
```

---

## POST /api/chat

Chat with an agent via Server-Sent Events (SSE). Requires authentication.

**Request Body:**

```json
{
  "agent_id": 1,
  "chat_id": "string (optional, for continuing a conversation)",
  "prompt": "string (required)"
}
```

**Response (SSE stream):**

```
data: {"content": "chunk1"}
data: {"content": " chunk2"}
data: {"content": " chunk3"}
data: [DONE]
```

---

## GET /api/chats

List chat sessions with pagination. Requires authentication.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": {
    "sessions": [
      {
        "id": "chat-uuid",
        "agent_id": 1,
        "user_id": 1,
        "title": "Trading discussion",
        "messages": [],
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 25,
    "page": 1,
    "page_size": 10
  },
  "msg": "success"
}
```

---

## GET /api/chat/:chatId

Get a specific chat session with all messages. Requires authentication.

**Path Parameters:**
- `chatId` (string) - Chat session ID

**Response:**

```json
{
  "code": 0,
  "data": {
    "id": "chat-uuid",
    "agent_id": 1,
    "user_id": 1,
    "title": "Trading discussion",
    "messages": [
      {
        "role": "user",
        "content": "What is a stop loss?"
      },
      {
        "role": "assistant",
        "content": "A stop loss is..."
      }
    ],
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  },
  "msg": "success"
}
```
