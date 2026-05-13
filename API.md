# My Blog Go Server API 文档

## 基础信息

- Base URL（默认配置）: `http://localhost:2333/api`
- 服务端口来自 `config/config.yaml` 的 `server.port`（当前是 `2333`）
- 数据格式: `application/json`

## 认证方式（当前实现）

- 受保护接口统一使用 `JWTAuthMiddleware`
- 中间件当前从 Cookie 读取令牌，而不是从 `Authorization: Bearer` 读取
  - `access-token`
  - `refresh-token`
- 登录成功后服务端会通过 `Set-Cookie` 写入上述两个 Cookie

> 注意: 文档中如果只带 Bearer Token、不带 Cookie，当前代码大概率无法通过鉴权。

---

## 通用响应格式

大多数接口返回:

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

### 业务码（`code`）

- `0`: 成功（`constant.SUCCESS`）
- `7`: 失败（`constant.ERROR`）

### HTTP 状态码

- 绝大多数接口即使失败也返回 HTTP `200`，通过 `code/msg` 区分成功失败
- 例外：`POST /api/create` 的部分参数错误/创建失败分支会直接返回 HTTP `400/500` 且格式为 `{ "message": "..." }`

---

## 接口可用性总览

### 已挂载且可调用

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/logout`（需 Cookie）
- `GET /api/user/:id`
- `GET /api/user/:id/posts`
- `GET /api/user/:id/comments`
- `GET /api/me`（需 Cookie）
- `POST /api/update_profile`（需 Cookie）
- `POST /api/update_password`（需 Cookie）
- `POST /api/create`（需 Cookie）
- `GET /api/post`
- `POST /api/comment`（需 Cookie）
- `DELETE /api/comment/:id`（需 Cookie）
- `POST /api/comment/like`（需 Cookie）
- `POST /api/comment/dislike`（需 Cookie）
- `GET /api/post/:id/comments`
- `GET /api/comments/:id/replies`
- `POST /api/update`（需 Cookie）
- `POST /api/post/update`（需 Cookie，`/api/update` 的兼容别名）
- `DELETE /api/post/:id`（需 Cookie）
- `POST /api/post/like`（需 Cookie）
- `POST /api/post/dislike`（需 Cookie）
- `POST /api/create_agent`（需 Cookie）
- `POST /api/update_agent`（需 Cookie）
- `GET /api/agents`（需 Cookie）
- `GET /api/agent/list`（需 Cookie，`/api/agents` 的兼容别名）
- `GET /api/agent/:id`（需 Cookie）
- `POST /api/invoke_agent`（需 Cookie）
- `POST /api/chat`（需 Cookie，SSE）
- `GET /api/chats`（需 Cookie）
- `GET /api/chat/:chat_id`（需 Cookie）

---

## 用户认证模块

### 1) 用户注册

- 接口: `POST /api/auth/register`
- 认证: 不需要

请求体:

```json
{
  "username": "string",
  "password": "string",
  "env": {
    "ipv4": "",
    "ipv6": "",
    "os": "Windows",
    "device_info": "Chrome"
  }
}
```

成功响应示例:

```json
{
  "code": 0,
  "data": {
    "env": {},
    "username": "testuser"
  },
  "msg": "register success"
}
```

---

### 2) 用户登录

- 接口: `POST /api/auth/login`
- 认证: 不需要

请求体:

```json
{
  "username": "string",
  "password": "string",
  "env": {}
}
```

成功响应示例:

```json
{
  "code": 0,
  "data": {
    "env": {},
    "token": {
      "access_token": "...",
      "access_token_expire_time": 114514,
      "refresh_token": "...",
      "refresh_token_expire_time": 114514
    },
    "user_info": {
      "user_id": 1,
      "created_at": "2026-05-12T10:00:00Z",
      "updated_at": "2026-05-12T10:00:00Z",
      "username": "testuser",
      "email": "",
      "phone": "",
      "bio": "",
      "avatar": "",
      "admin": false,
      "role": 0,
      "post_count": 0,
      "comment_count": 0
    }
  },
  "msg": "login success"
}
```

并且响应头会下发 Cookie：

- `Set-Cookie: access-token=...`
- `Set-Cookie: refresh-token=...`

---

### 3) 用户登出

- 接口: `POST /api/auth/logout`
- 认证: 需要 Cookie 认证

请求头示例:

```http
Cookie: access-token=<access_token>; refresh-token=<refresh_token>
```

成功响应示例:

```json
{
  "code": 0,
  "data": {},
  "msg": "Logout successful"
}
```

---

## 用户模块

### 4) 查询用户信息

- 接口: `GET /api/user/:id`
- 认证: 不需要

路径参数:

- `id`: 用户 ID

成功响应示例（直接返回 UserInfo 结构体）:

```json
{
  "code": 0,
  "data": {
    "user_id": 1,
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z",
    "username": "testuser",
    "email": "test@example.com",
    "phone": "1234567890",
    "bio": "This is a bio",
    "avatar": "",
    "admin": false,
    "role": 0,
    "post_count": 15,
    "comment_count": 42
  },
  "msg": "Query user by id successful"
}
```

---

### 5) 更新用户资料

- 接口: `POST /api/update_profile`
- 认证: 需要 Cookie 认证

请求头示例:

```http
Cookie: access-token=<access_token>; refresh-token=<refresh_token>
```

请求体:

```json
{
  "new_username": "newusername",
  "new_email": "newemail@example.com",
  "new_phone": "0987654321",
  "new_bio": "Updated bio",
  "new_avatar": "https://example.com/avatar.jpg",
  "env": {}
}
```

成功响应示例（接口返回值）:

```json
{
  "code": 0,
  "data": {
    "env": {},
    "user_info": {
      "user_id": 1,
      "created_at": "2026-05-12T10:00:00Z",
      "updated_at": "2026-05-12T10:00:00Z",
      "username": "newusername",
      "email": "newemail@example.com",
      "phone": "0987654321",
      "bio": "Updated bio",
      "avatar": "https://example.com/avatar.jpg",
      "admin": false,
      "role": 0,
      "post_count": 15,
      "comment_count": 42
    }
  },
  "msg": "UserUpdate profile successful"
}
```

> 当前状态: 该接口已修复，会从 JWT claims 读取当前用户 ID 后再更新。

---

### 6) 修改密码

- 接口: `POST /api/update_password`
- 认证: 需要 Cookie 认证

请求头示例:

```http
Cookie: access-token=<access_token>; refresh-token=<refresh_token>
```

请求体:

```json
{
  "old_password": "oldpassword",
  "new_password": "newpassword",
  "env": {}
}
```

成功响应示例:

```json
{
  "code": 0,
  "data": {
    "env": {},
    "user_info": {}
  },
  "msg": "UserUpdate password successful"
}
```

---

### 7) 获取当前登录用户

- 接口: `GET /api/me`
- 认证: 需要 Cookie 认证

请求头示例:

```http
Cookie: access-token=<access_token>; refresh-token=<refresh_token>
```

成功响应示例（直接返回 UserInfo 结构体）:

```json
{
  "code": 0,
  "data": {
    "user_id": 1,
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z",
    "username": "testuser",
    "email": "",
    "phone": "",
    "bio": "",
    "avatar": "",
    "admin": false,
    "role": 0,
    "post_count": 10,
    "comment_count": 5
  },
  "msg": "Query current user successful"
}
```

---

### 8) 查询用户文章列表

- 接口: `GET /api/user/:id/posts`
- 认证: 不需要

查询参数: `page`（默认1）、`page_size`（默认20，最大100）

> 返回该用户的所有文章，按时间倒序分页，结构与 `GET /api/post` 一致（含 author 信息）。

---

### 9) 查询用户评论列表

- 接口: `GET /api/user/:id/comments`
- 认证: 不需要

查询参数: `page`（默认1）、`page_size`（默认20，最大100）

> 返回该用户所有评论，按时间倒序分页（扁平列表）。

---

## 文章模块

> 响应使用 PostDetail 模型，包含 `author`（UserInfo 结构体，含用户全部公开信息）、`is_liked`/`is_disliked`（当前登录用户的点赞状态）。

### 10) 创建文章

- 接口: `POST /api/create`
- 认证: 需要 Cookie 认证

请求头示例:

```http
Cookie: access-token=<access_token>; refresh-token=<refresh_token>
```

请求体:

```json
{
  "title": "文章标题",
  "cover": "https://example.com/cover.jpg",
  "category": "技术",
  "keywords": ["Go", "后端", "API"],
  "content": "文章内容...",
  "public": true,
  "env": {}
}
```

说明:

- `user_id` 即使传入也会被服务端忽略，最终以 JWT 中用户 ID 为准

成功响应示例:

```json
{
  "code": 0,
  "data": {},
  "msg": "post success"
}
```

失败分支（该接口特有）可能返回:

```json
{
  "message": "Invalid request"
}
```

或

```json
{
  "message": "Failed to create post"
}
```

---

### 11) 查询文章（详情 / 列表）

- 接口: `GET /api/post`
- 认证: 不需要（详情私有可见性分支见下方说明）

#### 详情

`GET /api/post/:id`

> 返回 PostDetail（含 author 信息 + 当前用户 is_liked/is_disliked 状态）。未登录也可查看公开文章。

#### 列表

`GET /api/post?page=1&page_size=20`

> 分页返回 PostDetail 列表，按时间倒序。非公开文章内容隐藏。

---

### 12) 更新文章

- 接口: `POST /api/update`
- 兼容别名: `POST /api/post/update`
- 认证: 需要 Cookie 认证

请求体:

```json
{
  "post_id": 1,
  "title": "更新后的标题",
  "cover": "https://example.com/cover.jpg",
  "category": "技术",
  "keywords": "Go,后端",
  "content": "更新后的内容",
  "public": true,
  "env": {}
}
```

---

### 13) 删除文章

- 接口: `DELETE /api/post/:id`
- 认证: 需要 Cookie 认证
- 权限: 仅文章作者可删除

示例:

`DELETE /api/post/1`

---

### 14) 点赞文章（切换）

- 接口: `POST /api/post/like`
- 认证: 需要 Cookie 认证

请求体:

```json
{
  "post_id": 1
}
```

> 行为: 已点赞则取消点赞，已点踩则取消点踩并点赞，未操作则点赞。点赞与点踩互斥。

---

### 15) 点踩文章（切换）

- 接口: `POST /api/post/dislike`
- 认证: 需要 Cookie 认证

请求体:

```json
{
  "post_id": 1
}
```

> 行为: 已点踩则取消点踩，已点赞则取消点赞并点踩，未操作则点踩。点赞与点踩互斥。

---

## 评论模块

响应中的 Comment 结构（`comment_id` 而非 `id`）：

```json
{
  "comment_id": 1, "user_id": 1, "post_id": 1,
  "root_comment_id": 0, "parent_comment_id": 0,
  "content": "...", "created_at": "...", "updated_at": "...",
  "author": { "user_id": 1, "username": "...", "avatar": "...", "..." : "..." },
  "likes": 0, "dislikes": 0, "replies": 0,
  "is_liked": false, "is_disliked": false,
  "reply_comments": [ /* 递归嵌套子回复 */ ]
}
```

> `author` 字段使用统一 UserInfo 结构体（与用户模块一致）。
> `is_liked`/`is_disliked` 表示当前登录用户对该评论的点赞/点踩状态，未登录时为 `false`。

### 16) 创建评论

- 接口: `POST /api/comment`
- 认证: 需要 Cookie 认证

请求体:

```json
{
  "post_id": 1,
  "content": "评论内容",
  "root_comment_id": 0,
  "parent_comment_id": 0,
  "env": {}
}
```

> `root_comment_id = 0` 表示根评论；`parent_comment_id = 0` 表示直接回复根评论。

---

### 17) 删除评论

- 接口: `DELETE /api/comment/:id`
- 认证: 需要 Cookie（仅评论作者可删除根评论）

---

### 18) 点赞评论

- 接口: `POST /api/comment/like`
- 认证: 需要 Cookie

请求体:

```json
{ "comment_id": 1 }
```

> 切换逻辑：已点赞则取消，已点踩则转点赞，未操作则点赞。

---

### 19) 点踩评论

- 接口: `POST /api/comment/dislike`
- 认证: 需要 Cookie

请求体:

```json
{ "comment_id": 1 }
```

> 切换逻辑与点赞对称。

---

### 20) 获取文章根评论列表

- 接口: `GET /api/post/:id/comments`
- 认证: 不需要

查询参数: `page`（默认1）、`page_size`（默认20，最大100）

> 返回根评论列表，每条根评论的 `reply_comments` 递归包含完整楼中楼树。

---

### 21) 获取评论回复列表

- 接口: `GET /api/comments/:id/replies`
- 认证: 不需要

查询参数: `page`（默认1）、`page_size`（默认20，最大100）

> 返回指定根评论下的直接回复树，逐层递归嵌套。

---

## AI Agent 模块

### 22) 创建 Agent

- 接口: `POST /api/create_agent`
- 认证: 需要 Cookie 认证

请求头示例:

```http
Cookie: access-token=<access_token>; refresh-token=<refresh_token>
```

请求体:

```json
{
  "agent_name": "My Assistant",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxxxxxxxxxxxx",
  "provider": "openai",
  "model_name": "gpt-4"
}
```

成功响应:

```json
{
  "code": 0,
  "data": {},
  "msg": "create agent success"
}
```

> 当前状态: `create_agent` 已支持在创建时直接持久化 `api_key`。

---

### 23) 更新 Agent

- 接口: `POST /api/update_agent`
- 认证: 需要 Cookie 认证

请求体（`agent_id` 在 body 中，不是路径参数）:

```json
{
  "agent_id": 1,
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

成功响应:

```json
{
  "code": 0,
  "data": {},
  "msg": "update agent success"
}
```

---

### 24) 查询 Agent 列表

- 接口: `GET /api/agents`
- 兼容别名: `GET /api/agent/list`
- 认证: 需要 Cookie 认证

说明:

- 返回当前登录用户自己的 Agent 列表

---

### 25) 查询 Agent 详情

- 接口: `GET /api/agent/:id`
- 认证: 需要 Cookie 认证
- 权限: 仅 Agent 所有者可查询

---

### 26) 调用 Agent

- 接口: `POST /api/invoke_agent`
- 认证: 需要 Cookie 认证

请求体:

```json
{
  "agent_id": 1,
  "ai_options": {
    "temperature": 0.7,
    "thinking": false,
    "search_internet": false
  },
  "sys_prompt": "系统提示词",
  "usr_prompt": "用户问题"
}
```

当前状态:

- 该接口已改为基于 JWT claims 获取 `user_id`，可正常进行用户维度鉴权。

---

### 27) 在线流式对话（SSE）

- 接口: `POST /api/chat?chat_id={chat_id}`
- 认证: 需要 Cookie 认证
- 返回: `text/event-stream`

查询参数:

- `chat_id`: 会话 ID（可选）

行为说明:

- 首次不传 `chat_id` 时，服务端会先生成会话 ID，并在**第一个 SSE 事件**返回
- 前端拿到该 `chat_id` 后，后续请求通过 `POST /api/chat?chat_id=...` 继续同一会话

请求体:

```json
{
  "agent_id": 1,
  "ai_options": {
    "temperature": 0.7,
    "thinking": false,
    "search_internet": false
  },
  "sys_prompt": "系统提示词",
  "usr_prompt": "用户问题"
}
```

流式事件示例（首包先返回会话信息）:

```text
data: {"chat_id":"uuid-string","agent_id":1,"user_id":1,"history":null}

data: {"agent_id":1,"chat_id":"uuid-string","reasoning_content":"","content":"你","status":true,"message":""}

data: {"agent_id":1,"chat_id":"uuid-string","reasoning_content":"","content":"你好","status":true,"message":"done"}
```

> 说明: 该接口全程只返回 SSE 事件，不再在结束时额外追加标准 JSON。

---

### 28) 查询会话列表

- 接口: `GET /api/chats`
- 认证: 需要 Cookie 认证

查询参数: `page`（默认1）、`page_size`（默认20，最大100）

> 返回当前用户的所有聊天会话，按更新时间倒序。`title` 由首条用户消息自动截取前100字符。

---

### 29) 查询会话详情

- 接口: `GET /api/chat/:chat_id`
- 认证: 需要 Cookie 认证

> 返回单个会话的完整信息，包括所有历史消息。

---

## 调用建议（按当前实现）

1. 先 `register` / `login`
2. 从登录响应拿到 token，同时保留 `Set-Cookie`
3. 调用受保护接口时优先带 Cookie
4. AI 场景可以直接 `create_agent`（已支持保存 `api_key`）
5. 如需普通请求/响应模式可用 `invoke_agent`，流式场景优先使用 `chat`
6. 会话历史存储在 PostgreSQL，刷新页面不丢失；可通过 `GET /api/chats` 获取列表

---

