# My Blog Go Server API 文档

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **认证方式**: JWT Token (Bearer Token)
- **数据格式**: JSON

## 通用响应格式

所有 API 响应遵循以下格式：

```json
{
  "code": 200,
  "data": {},
  "msg": "Success"
}
```

### 响应码说明

- `200`: 成功
- `400`: 请求参数错误
- `401`: 未授权/Token 无效
- `500`: 服务器内部错误

---

## 用户认证模块 (Auth)

### 1. 用户注册

**接口**: `POST /api/auth/register`

**认证**: 不需要

**请求体**:
```json
{
  "username": "string",
  "password": "string",
  "env": {
    // 环境信息（可选）
  }
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "env": {},
    "username": "testuser"
  },
  "msg": "register success"
}
```

---

### 2. 用户登录

**接口**: `POST /api/auth/login`

**认证**: 不需要

**请求体**:
```json
{
  "username": "string",
  "password": "string",
  "env": {
    // 环境信息（可选）
  }
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "env": {},
    "token": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "access_token_expire_time": 3600,
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token_expire_time": 86400
    },
    "user_info": {
      "id": 1,
      "username": "testuser",
      "email": "",
      "phone": "",
      "bio": ""
    }
  },
  "msg": "login success"
}
```

---

### 3. 用户登出

**接口**: `POST /api/auth/logout`

**认证**: 需要 JWT Token

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应示例**:
```json
{
  "code": 200,
  "data": {},
  "msg": "Logout successful"
}
```

---

## 用户管理模块 (User)

### 4. 查询用户信息

**接口**: `GET /api/user/:id`

**认证**: 不需要

**路径参数**:
- `id`: 用户 ID (uint)

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "user_info": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "phone": "1234567890",
      "bio": "This is a bio"
    }
  },
  "msg": "Query user by id successful"
}
```

---

### 5. 更新用户资料

**接口**: `POST /api/update_profile`

**认证**: 需要 JWT Token

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "id": 1,
  "new_username": "newusername",
  "new_email": "newemail@example.com",
  "new_phone": "0987654321",
  "new_bio": "Updated bio",
  "new_avatar": "https://example.com/avatar.jpg",
  "env": {
    // 环境信息（可选）
  }
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "env": {},
    "user_info": {
      "id": 1,
      "username": "newusername",
      "email": "newemail@example.com",
      "phone": "0987654321",
      "bio": "Updated bio"
    }
  },
  "msg": "UserUpdate profile successful"
}
```

---

### 6. 修改密码

**接口**: `POST /api/update_password`

**认证**: 需要 JWT Token

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "id": 1,
  "old_password": "oldpassword",
  "new_password": "newpassword",
  "env": {
    // 环境信息（可选）
  }
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "env": {},
    "user_info": {}
  },
  "msg": "UserUpdate password successful"
}
```

---

## 文章管理模块 (Post)

### 7. 创建文章

**接口**: `POST /api/create`

**认证**: 需要 JWT Token

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "user_id": 1,
  "title": "文章标题",
  "cover": "https://example.com/cover.jpg",
  "category": "技术",
  "keywords": "Go,后端,API",
  "content": "文章内容...",
  "public": true,
  "env": {
    // 环境信息（可选）
  }
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {},
  "msg": "post success"
}
```

---

### 8. 查询文章详情

**接口**: `GET /api/post/:id`

**认证**: 不需要

**路径参数**:
- `id`: 文章 ID (uint)

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "文章标题",
    "cover": "https://example.com/cover.jpg",
    "category": "技术",
    "keywords": "Go,后端,API",
    "content": "文章内容...",
    "public": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "msg": "query success"
}
```

**注意**: 仅返回公开的文章 (`public: true`)

---

### 9. 更新文章

**接口**: `POST /api/update` (待实现)

**认证**: 需要 JWT Token

**请求体**:
```json
{
  "post_id": 1,
  "title": "更新后的标题",
  "cover": "https://example.com/new-cover.jpg",
  "category": "新技术",
  "keywords": "Go,更新",
  "content": "更新后的内容...",
  "public": true,
  "env": {
    // 环境信息（可选）
  }
}
```

---

## AI 服务模块 (AI Agent)

### 10. 创建 AI Agent

**接口**: `POST /api/create_agent`

**认证**: 需要 JWT Token

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "agent_name": "My Assistant",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxxxxxxxxxxxx",
  "provider": "openai",
  "model_name": "gpt-4"
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {},
  "msg": "create agent success"
}
```

---

### 11. 更新 AI Agent

**接口**: `POST /api/update_agent`

**认证**: 需要 JWT Token

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `agent_id`: Agent ID (从上下文获取)

**请求体**:
```json
{
  "agent_name": "Updated Assistant",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-newkey",
  "provider": "openai",
  "model_name": "gpt-4-turbo",
  "prompts": {
    "system": "You are a helpful assistant"
  },
  "memories": {
    "key1": "value1"
  },
  "activate": true
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {},
  "msg": "update agent success"
}
```

---

### 12. 调用 AI Agent

**接口**: `POST /api/invoke_agent`

**认证**: 需要 JWT Token

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `agent_id`: Agent ID (从上下文获取)

**请求体**:
```json
{
  "ai_options": {
    "temperature": 0.7,
    "thinking": false,
    "search_internet": false
  },
  "sys_prompt": "系统提示词",
  "usr_prompt": "用户问题"
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "agent_id": 1,
    "reasoning_content": "",
    "content": "AI 回答内容",
    "status": true,
    "message": ""
  },
  "msg": "success"
}
```

---

### 13. 在线流式对话

**接口**: `POST /api/chat?chat_id={chat_id}`

**认证**: 需要 JWT Token

**请求头**:
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**查询参数**:
- `chat_id`: 会话 ID (可选，不提供则自动生成)

**路径参数**:
- `agent_id`: Agent ID (从上下文获取)

**请求体**:
```json
{
  "ai_options": {
    "temperature": 0.7,
    "thinking": false,
    "search_internet": false
  },
  "sys_prompt": "系统提示词",
  "usr_prompt": "用户问题"
}
```

**响应格式**: Server-Sent Events (SSE)

**流式响应示例**:
```
data: {"agent_id":1,"reasoning_content":"","content":"你","status":true,"message":""}

data: {"agent_id":1,"reasoning_content":"","content":"你好","status":true,"message":""}

data: {"agent_id":1,"reasoning_content":"","content":"你好！","status":true,"message":""}
```

**最终响应**:
```json
{
  "code": 200,
  "data": {
    "chat_id": "uuid-string",
    "agent_id": 1,
    "user_id": 1
  },
  "msg": "success"
}
```

---

## 认证说明

### JWT Token 使用

需要在请求头中添加 Authorization 字段：

```
Authorization: Bearer <your_access_token>
```

### Token 获取

通过登录接口 `/api/auth/login` 获取 access_token 和 refresh_token。

### Token 失效处理

当收到 `401` 错误或响应中包含 `"reload": true` 时，需要重新登录获取新的 Token。

---

## 错误处理

### 常见错误响应

**参数错误**:
```json
{
  "code": 400,
  "data": {},
  "msg": "Invalid request"
}
```

**认证失败**:
```json
{
  "code": 401,
  "data": {
    "reload": true
  },
  "msg": "Unauthorized"
}
```

**服务器错误**:
```json
{
  "code": 500,
  "data": {},
  "msg": "Failed to create post"
}
```

---

## 注意事项

1. **密码安全**: 当前版本密码以明文传输，建议在生产环境使用 HTTPS
2. **环境变量**: 部分接口支持 `env` 参数传递环境信息
3. **文章可见性**: 查询文章接口只返回 `public: true` 的文章
4. **流式对话**: `/api/chat` 接口使用 SSE 协议，需要客户端支持事件流处理
5. **Agent 激活**: 更新 Agent 时可设置 `activate` 字段控制是否激活该 Agent

---

## 更新日志

- **v1.0**: 初始版本，包含用户认证、文章管理、AI Agent 功能
