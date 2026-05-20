# AI 智能体 API

- [通用说明](./API_common.md)

---

## 概述

AI 模块支持用户创建自定义 LLM 智能体（Agent），配置模型参数、提示词和记忆，并进行单次调用或流式多轮对话。

**支持的提供商：** `openai`、`ollama`

---

## 1. 创建智能体

**POST** `/api/create_agent` `[认证]`

### 请求体

```json
{
  "agent_name": "my-assistant",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxx",
  "provider": "openai",
  "model_name": "gpt-4"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_name` | string | 是 | 智能体名称 |
| `base_url` | string | 是 | API 基础地址 |
| `api_key` | string | 是 | API 密钥 |
| `provider` | string | 是 | 模型提供商，仅支持 `openai` / `ollama` |
| `model_name` | string | 是 | 模型名称 |

### 响应 data

空对象

---

## 2. 更新智能体

**POST** `/api/update_agent` `[认证]`

### 请求体

```json
{
  "agent_id": 1,
  "agent_name": "my-assistant-v2",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxx",
  "provider": "openai",
  "model_name": "gpt-4-turbo",
  "prompts": { "system": "You are a helpful assistant" },
  "memories": { "user_name": "Alice" },
  "activate": true
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_id` | uint | 是 | 智能体 ID |
| `agent_name` | string | 否 | 新名称 |
| `base_url` | string | 否 | API 地址 |
| `api_key` | string | 否 | API 密钥 |
| `provider` | string | 否 | 提供商 |
| `model_name` | string | 否 | 模型名 |
| `prompts` | JSONMap | 否 | 提示词配置 |
| `memories` | JSONMap | 否 | 记忆键值对 |
| `activate` | bool | 否 | 是否激活 |

### 响应 data

空对象

---

## 3. 获取智能体列表

**GET** `/api/agents` `[认证]`

### 响应 data

```json
[
  {
    "agent_id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "user_id": 1,
    "name": "my-assistant",
    "model_name": "gpt-4",
    "provider": "openai",
    "activate": true,
    "temperature": 0.7,
    "thinking": false,
    "prompts": {},
    "memories": {}
  }
]
```

---

## 4. 查询单个智能体

**GET** `/api/agent/:id` `[认证]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 智能体 ID |

### 响应 data

> **安全设计**：`api_key` 和 `base_url` 不在响应中返回，防止敏感信息泄露。

同 [获取智能体列表](#3-获取智能体列表) 中的单个元素结构

---

## 5. 调用智能体（非流式）

**POST** `/api/invoke_agent` `[认证]`

> 单次调用，返回完整回复内容

### 请求体

```json
{
  "agent_id": 1,
  "ai_options": {
    "temperature": 0.7,
    "thinking": false,
    "search_internet": false
  },
  "sys_prompt": "override system prompt (可选)",
  "usr_prompt": "你好，请介绍一下自己"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent_id` | uint | 是 | 智能体 ID |
| `ai_options.temperature` | float64 | 否 | 温度参数 |
| `ai_options.thinking` | bool | 否 | 是否启用思维链 |
| `ai_options.search_internet` | bool | 否 | 是否联网搜索（未实现） |
| `sys_prompt` | string | 否 | 系统提示词 |
| `usr_prompt` | string | 是 | 用户消息 |

### 响应 data

```json
{
  "agent_id": 1,
  "chat_id": "",
  "reasoning_content": "",
  "content": "你好！我是 AI 助手，有什么可以帮你的？",
  "status": true,
  "message": "success"
}
```

---

## 6. 流式聊天

**POST** `/api/chat` `[认证]`

> 使用 **Server-Sent Events (SSE)** 推送流式响应，支持多轮对话

### Query 参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `chat_id` | string | 可选，传入已有会话 UUID 以继续对话；不传则自动生成 |

### 请求体

同 [调用智能体](#5-调用智能体非流式)

### SSE 响应格式

#### 第一帧：历史记录

```json
{
  "chat_id": "uuid-xxx",
  "agent_id": 1,
  "user_id": 1,
  "history": [
    { "role": "user", "content": "之前的问题" },
    { "role": "assistant", "content": "之前的回答" }
  ]
}
```

#### 后续帧：增量回复

```
data: {"agent_id":1,"content":"你好","status":true,"message":"","reasoning_content":"","chat_id":"uuid-xxx"}

data: {"agent_id":1,"content":"","status":true,"message":"done","reasoning_content":"","chat_id":"uuid-xxx"}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_id` | uint | 智能体 ID |
| `chat_id` | string | 会话 UUID |
| `content` | string | 增量文本 |
| `status` | bool | 正常为 `true`，出错为 `false` |
| `message` | string | `""` 为进行中，`"done"` 为流结束 |
| `reasoning_content` | string | 思维链内容（未实现） |

---

## 7. 查询聊天会话列表

**GET** `/api/chats` `[认证]`

### Query 参数

| 参数 | 类型 | 默认值 |
|------|------|--------|
| `page` | int | 1 |
| `page_size` | int | 20 |

### 响应 data

```json
{
  "items": [
    {
      "id": 1,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "deleted_at": null,
      "uuid": "chat-uuid-xxx",
      "user_id": 1,
      "agent_id": 1,
      "title": "你好，请介绍一下自己",
      "messages": [
        { "role": "user", "content": "你好", "time": 1704067200 },
        { "role": "model", "content": "你好！有什么可以帮你的？", "time": 1704067201 }
      ]
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 5
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `uuid` | string | 会话唯一标识，用于继续对话 |
| `title` | string | 会话标题（截取首条用户消息前 100 字符） |
| `agent_id` | uint | 关联的智能体 ID |
| `messages` | JSON | 历史消息数组 `[{role, content, time}]` |

---

## 8. 查询单个聊天会话

**GET** `/api/chat/:chat_id` `[认证]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `chat_id` | string | 会话 UUID |

### 响应 data

同 [查询聊天会话列表](#7-查询聊天会话列表) 中的单个元素结构
