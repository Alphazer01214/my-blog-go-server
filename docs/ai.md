# AI / Agent API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/ai/agent` | GET | 获取当前用户的 Agent 列表 | 是 |
| `/api/ai/agent` | POST | 创建 Agent | 是 |
| `/api/ai/agent/:agent_id` | PUT | 更新 Agent | 是 |
| `/api/ai/agent/:agent_id` | DELETE | 删除 Agent | 是 |
| `/api/ai/invoke` | POST | 非流式调用 AI | 是 |
| `/api/chat` | POST | 流式 AI 聊天（SSE） | 否 |

---

## 端点详情

### GET /api/ai/agent

获取当前用户的 Agent 列表。

**Response**：

```json
{
  "data": [
    {
      "id": "agent-uuid",
      "name": "助手",
      "avatar_path": "",
      "description": "描述",
      "sub_description": "子描述",
      "system_prompt": "你是...",
      "provider": "openai",
      "model_id": "gpt-4o-mini",
      "temperature": 0.7,
      "max_tokens": 2048,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### POST /api/ai/agent

创建新的 Agent。

**Request Body**：

```json
{
  "name": "助手",
  "avatar_path": "可选头像路径",
  "description": "描述",
  "sub_description": "子描述",
  "system_prompt": "系统提示词",
  "provider": "openai",
  "model_id": "gpt-4o-mini",
  "temperature": 0.7,
  "max_tokens": 2048
}
```

- `provider`：支持 `openai` 和 `ollama`
- `model_id`：模型名称，由对应 provider 支持

**Response**：

```json
{
  "data": {
    "id": "agent-uuid"
  }
}
```

### PUT /api/ai/agent/:agent_id

更新 Agent。仅创建者可操作。全部为可选字段。

**Request Body**：同创建，所有字段可选。

**Response**：标准成功响应。

### DELETE /api/ai/agent/:agent_id

删除 Agent。仅创建者可操作。

**Response**：标准成功响应。

### POST /api/ai/invoke

非流式调用 AI（同步返回完整结果）。

**Request Body**：

```json
{
  "agent_id": "agent-uuid",
  "messages": [
    {
      "role": "user",
      "content": "你好"
    }
  ],
  "use_search": true
}
```

- `use_search`：是否启用联网搜索
- `agent_id`：指定使用的 Agent 配置

**Response**：

```json
{
  "data": {
    "content": "AI 回复内容"
  }
}
```

### POST /api/chat

流式 AI 聊天（SSE 流式输出）。**需要登录**（中间件强制）。

> 与帖子 AI 问答不同，该接口支持自定义 System Prompt（通过 Agent 配置）和联网搜索。

**Request Body**：

```json
{
  "agent_id": "agent-uuid",
  "messages": [
    { "role": "user", "content": "你好" }
  ],
  "use_search": true,
  "chat_id": "（可选，前端的 UUID，用于多轮对话）",
  "env_info": {
    "ipv4": "...",
    "os": "...",
    "device_info": "..."
  }
}
```

- `chat_id`：可选。传入则复用历史，不传则每次新建会话
- `use_search`：是否启用联网搜索
- 每次请求**必须**通过 `messages` 传入完整的对话历史（包括之前的问答）

**SSE 首帧**（包含历史）：

```json
data: {"chat_id":"uuid","history":[{"role":"user","content":"你好"}],"status":true}
```

**SSE 流式块**：

```json
data: {"chat_id":"uuid","content":"增量文本","status":true}
```

**SSE 完成帧**：

```json
data: {"chat_id":"uuid","content":"","status":true,"message":"done"}
```

**错误帧**：

```json
data: {"chat_id":"uuid","content":"错误描述","status":false}
```

---

## 前端实现说明

### 流式聊天流程

```
1. 首次：生成 chat_id（crypto.randomUUID()）
2. POST /api/chat 发送请求（保持连接打开）
3. 按行读取 data: ... 事件
4. 首帧：记录 chat_id，保存 history（用于后续请求）
5. 流式块：逐个 append 到消息气泡
6. 完成帧：关闭连接
7. 下次提问：将历史 messages + 新问题 一起发送，复用同一 chat_id
```

### 联网搜索

- `use_search: true` 时，AI 会调用百度千帆搜索 API 获取实时信息
- 搜索工具触发时，SSE 流中会先返回一条 `"content":"正在搜索..."` 的思考状态
