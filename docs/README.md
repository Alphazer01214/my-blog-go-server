# Trading Forum 前端开发指南

> 版本：2.0.0
> 基础 URL：`http://<host>:2333/api`
> 内容类型：`application/json`（除非另有说明）

---

## 目录

1. [设计风格要求](#设计风格要求)
2. [认证机制](#认证机制)
3. [API 通用规范](#api-通用规范)
4. [前端集成指南](#前端集成指南)
5. [API 端点总览](#api-端点总览)
6. [数据模型参考](#数据模型参考)
7. [详细 API 文档](#详细-api-文档)

---

## 设计风格要求

### 网页风格

带有交易论坛的风格，同时增加年轻化元素。

### 文章内容渲染

Post 文章内容由前端 Markdown 渲染，需支持以下特性：

1. **图像插入**：通过标准 Markdown 图片语法 `![alt](url)`
2. **话题链接**：通过 `#` 添加话题，例如 `#上证指数`
   - 点击跳转至搜索页：`/search?keyword=上证指数`
3. **关键词识别**：识别股票名称并高亮
   - 指数名如 `上证指数` → 跳转至东方财富网搜索页
   - 个股名如 `中际旭创` → 跳转至东方财富网个股页
   - 建议使用正则匹配 + 自定义渲染器实现

---

## 认证机制

使用 **JWT 双 Token** 机制，存储在 **HttpOnly Cookie** 中，前端**无需手动管理 Token**。

| Cookie 名 | 类型 | HttpOnly | 说明 |
|-----------|------|----------|------|
| `access-token` | JWT | 是 | 短期（15min），认证用 |
| `refresh-token` | JWT | 是 | 长期（7天），刷新用 |

**工作流程**：

1. 调用 `POST /api/auth/login`，Token 自动注入到 `Set-Cookie` 响应头
2. 后续所有请求自动携带 Cookie
3. `access-token` 过期后，中间件自动用 `refresh-token` 刷新
4. 两者都过期时，返回 `{ code: 7, data: { reload: true }, msg: "token expired" }`

---

## API 通用规范

### 统一响应格式

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | `0`=成功，`7`=错误 |
| `data` | object/array | 业务数据 |
| `msg` | string | 提示消息 |

**Token 失效响应**（前端需跳转登录页）：

```json
{ "code": 7, "data": { "reload": true }, "msg": "token expired" }
```

### 分页规范

**请求参数**（Query String）：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | 页码 |
| `page_size` | int | 否 | 20 | 每页数量（最大 100） |

**响应格式**：

```json
{
  "code": 0,
  "data": {
    "items": [],
    "page": 1,
    "page_size": 20,
    "total": 42
  },
  "msg": "Success"
}
```

### SSE 流式格式

AI 相关接口使用 Server-Sent Events。

**响应头**：

```
Content-Type: text/event-stream
Connection: keep-alive
Cache-Control: no-cache
```

**事件格式**（每行以 `data: ` 开头，`\n\n` 结尾）：

```
data: {"chat_id":"uuid","content":"思考中","status":true}
data: {"chat_id":"uuid","content":"文本块","status":true}
data: {"chat_id":"uuid","content":"","status":true,"message":"done"}
```

| 事件 | 说明 |
|------|------|
| 首帧（仅 `/api/chat`） | 携带 `history` 数组 |
| 流式块 | `content` 为增量文本 |
| 完成帧 | `message` = `"done"` |
| 错误帧 | `status` = `false` |

### EnvInfo 结构

部分请求包含可选的环境信息字段：

```json
{
  "ipv4": "192.168.1.1",
  "ipv6": "",
  "os": "Windows 11",
  "device_info": "Chrome 120"
}
```

### 枚举常量

#### 用户角色

| 值 | 说明 |
|----|------|
| `normal_user` | 普通用户（默认） |
| `vip` | VIP 用户 |
| `moderator` | 版主 |
| `takamatsu_tomori` | 超级管理员 |
| `evil` | 恶意用户（封禁） |
| `guest` | 访客（未注册） |

#### 交互行为（action_type）

| 值 | 说明 |
|----|------|
| `like` | 点赞 |
| `dislike` | 点踩 |
| `favorite` | 收藏 |
| `share` | 分享 |

#### 目标类型（target_type）

| 值 | 说明 |
|----|------|
| `post` | 帖子 |
| `comment` | 评论 |
| `video` | 视频 |
| `user` | 用户 |
| `tag` | 标签 |

#### AI Provider

| 值 | 说明 |
|----|------|
| `openai` | OpenAI 兼容接口 |
| `ollama` | Ollama 本地模型 |

#### Post Ask Mode

| 值 | 说明 |
|----|------|
| `summarize` | 总结文章 |
| `ask` | 对文章提问 |
| `selected` | 划词提问 |

---

## 前端集成指南

### 请求配置

所有 API 请求必须携带 Cookie：

```javascript
// fetch
fetch('/api/xxx', {
  credentials: 'include',
  headers: { 'Content-Type': 'application/json' }
})

// axios
axios.defaults.withCredentials = true
```

### 全局错误处理

```javascript
async function request(url, options = {}) {
  const res = await fetch(url, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    ...options
  })
  const data = await res.json()

  if (data.code !== 0) {
    // Token 失效 → 跳转登录页
    if (data.data?.reload) {
      window.location.href = '/login'
      return
    }
    throw new Error(data.msg)
  }
  return data.data
}
```

### 登录流程

```
1. POST /api/auth/login { username, password }
2. 响应 Set-Cookie 自动保存 access-token 和 refresh-token
3. 后续请求自动携带 Cookie
4. 退出时 POST /api/auth/logout，Cookie 被清除
```

### SSE 流式请求

由于 `EventSource` 不支持 POST + credentials，需用 `fetch` + `ReadableStream`：

```javascript
async function streamChat(payload, onChunk, onDone) {
  const response = await fetch('/api/chat', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() // 保留不完整的行

    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const json = JSON.parse(line.slice(6))
        if (json.message === 'done') {
          onDone()
          return
        }
        if (json.status === false) {
          throw new Error(json.message || json.content)
        }
        onChunk(json.content)
      }
    }
  }
}
```

### 文件分块上传

三步流程：

```javascript
// 1. 初始化
const { upload_id, chunk_size, total_chunks } = await request('/api/upload/init', {
  method: 'POST',
  body: JSON.stringify({ filename: file.name, file_size: file.size, mime_type: file.type })
})

// 2. 逐块上传
for (let i = 0; i < total_chunks; i++) {
  const start = i * chunk_size
  const end = Math.min(start + chunk_size, file.size)
  const chunk = file.slice(start, end)

  await request('/api/upload/chunk', {
    method: 'POST',
    body: JSON.stringify({ upload_id, chunk_index: i, chunk_hash: '...' })
  })
  // 进度 = (i + 1) / total_chunks * 100%
}

// 3. 合并
const fileDetail = await request('/api/upload/complete', {
  method: 'POST',
  body: JSON.stringify({ upload_id, title: '文件标题', public: true })
})
```

**断点续传**：上传前查询 `GET /api/upload/:uploadId/status` 获取已上传块列表，跳过已上传的块。

### 视频上传

与文件上传类似，额外 `merge` 步骤：

```javascript
// 1-2. 同文件上传（init → chunk）
// 3. 合并 + 转码
const video = await request('/api/video/merge', {
  method: 'POST',
  body: JSON.stringify({ upload_id })
})
```

### 点赞/点踩/收藏

独立端点，Toggle 模式（再次调用取消）：

```javascript
// 点赞/取消点赞
await request('/api/post/like', { method: 'POST', body: JSON.stringify({ post_id: 1 }) })

// 点踩/取消点踩
await request('/api/post/dislike', { method: 'POST', body: JSON.stringify({ post_id: 1 }) })

// 收藏/取消收藏
await request('/api/post/favorite', { method: 'POST', body: JSON.stringify({ post_id: 1 }) })
```

评论和视频的交互端点类似：`/api/comment/like`、`/api/video/:id/like` 等。

---

## API 端点总览

### 认证

| 端点 | 方法 | 说明 | 登录 |
|------|------|------|------|
| `/api/auth/register` | POST | 注册 | 否 |
| `/api/auth/login` | POST | 登录 | 否 |
| `/api/auth/logout` | POST | 退出 | 是 |

### 用户

| 端点 | 方法 | 说明 | 登录 |
|------|------|------|------|
| `/api/me` | GET | 当前用户信息 | 是 |
| `/api/user/:id` | GET | 用户信息 | 否 |
| `/api/user/:id/posts` | GET | 用户的帖子 | 否 |
| `/api/user/:id/comments` | GET | 用户的评论 | 否 |
| `/api/user/:id/followers` | GET | 粉丝列表 | 否 |
| `/api/user/:id/following` | GET | 关注列表 | 否 |
| `/api/all_users` | GET | 所有用户 | 否 |
| `/api/update_password` | POST | 修改密码 | 是 |
| `/api/update_profile` | POST | 修改资料 | 是 |
| `/api/user/follow` | POST | 关注/取消关注 | 是 |
| `/api/settings` | GET | 获取设置 | 是 |
| `/api/settings` | POST | 更新设置 | 是 |

### 帖子

| 端点 | 方法 | 说明 | 登录 |
|------|------|------|------|
| `/api/create` | POST | 创建帖子 | 是 |
| `/api/post/:id` | GET | 帖子详情 | 是 |
| `/api/post` | GET | 帖子列表 | 是 |
| `/api/update` | POST | 更新帖子 | 是 |
| `/api/post/:id` | DELETE | 删除帖子 | 是 |
| `/api/post/like` | POST | 点赞 | 是 |
| `/api/post/dislike` | POST | 点踩 | 是 |
| `/api/post/favorite` | POST | 收藏 | 是 |
| `/api/post/share` | POST | 分享 | 是 |
| `/api/post/search` | GET | 搜索 | 是 |
| `/api/post/ask` | POST | AI 问答 (SSE) | 是 |

### 评论

| 端点 | 方法 | 说明 | 登录 |
|------|------|------|------|
| `/api/comment` | POST | 创建评论 | 是 |
| `/api/comment/:id` | DELETE | 删除评论 | 是 |
| `/api/comment/like` | POST | 点赞 | 是 |
| `/api/comment/dislike` | POST | 点踩 | 是 |
| `/api/post/:id/comments` | GET | 帖子评论 | 否 |
| `/api/video/:id/comments` | GET | 视频评论 | 否 |
| `/api/user/:id/comments` | GET | 用户评论 | 否 |

### 视频

| 端点 | 方法 | 说明 | 登录 |
|------|------|------|------|
| `/api/video/:id/stream` | GET | 视频流 | 否 |
| `/api/video/:id/cover` | GET | 封面 | 否 |
| `/api/video/init` | POST | 初始化上传 | 是 |
| `/api/video/chunk` | POST | 上传分块 | 是 |
| `/api/video/merge` | POST | 合并 | 是 |
| `/api/video/create` | POST | 创建视频 | 是 |
| `/api/video/:id` | GET | 视频详情 | 是 |
| `/api/video` | GET | 视频列表 | 是 |
| `/api/video/user/:userId` | GET | 用户视频 | 是 |
| `/api/video/:id` | PUT | 更新 | 是 |
| `/api/video/:id` | DELETE | 删除 | 是 |
| `/api/video/:id/like` | POST | 点赞 | 是 |
| `/api/video/:id/dislike` | POST | 点踩 | 是 |
| `/api/video/:id/favorite` | POST | 收藏 | 是 |
| `/api/video/:id/share` | POST | 分享 | 是 |

### 文件上传

| 端点 | 方法 | 说明 | 登录 |
|------|------|------|------|
| `/api/upload/init` | POST | 初始化 | 是 |
| `/api/upload/chunk` | POST | 上传分块 | 是 |
| `/api/upload/complete` | POST | 合并 | 是 |
| `/api/upload/:uploadId/status` | GET | 上传状态 | 是 |
| `/api/upload/:uploadId/abort` | POST | 取消 | 是 |
| `/api/file/:id` | GET | 文件信息 | 是 |
| `/api/file/:id/download` | GET | 下载 | 是 |
| `/api/files` | GET | 文件列表 | 是 |
| `/api/user/:id/files` | GET | 用户文件 | 是 |
| `/api/file/:id` | DELETE | 删除 | 是 |

### AI / Agent

| 端点 | 方法 | 说明 | 登录 |
|------|------|------|------|
| `/api/agents` | POST | 创建 Agent | 是 |
| `/api/agents/:id` | PUT | 更新 Agent | 是 |
| `/api/agents` | GET | Agent 列表 | 是 |
| `/api/agent/:id` | GET | Agent 详情 | 是 |
| `/api/invoke_agent` | POST | 非流式调用 | 是 |
| `/api/chat` | POST | 流式聊天 (SSE) | 是 |
| `/api/chats` | GET | 聊天会话列表 | 是 |
| `/api/chat/:chatId` | GET | 聊天会话详情 | 是 |

### 行情

| 端点 | 方法 | 说明 | 登录 |
|------|------|------|------|
| `/api/market` | GET | 最新行情 | 否 |
| `/api/market/history` | GET | 历史行情 | 否 |

### 管理后台

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/tomori/stats` | GET | 统计面板 |
| `/api/tomori/users` | GET | 用户列表 |
| `/api/tomori/user/ban` | POST | 封禁/解封 |
| `/api/tomori/user/role` | POST | 修改角色 |
| `/api/tomori/user/reset_password` | POST | 重置密码 |
| `/api/tomori/user/:id` | DELETE | 删除用户 |
| `/api/tomori/posts` | GET | 帖子列表 |
| `/api/tomori/post/:id` | DELETE | 删除帖子 |
| `/api/tomori/comments` | GET | 评论列表 |
| `/api/tomori/comment/:id` | DELETE | 删除评论 |
| `/api/tomori/files` | GET | 文件列表 |
| `/api/tomori/file/:id` | DELETE | 删除文件 |
| `/api/tomori/agents` | GET | Agent 列表 |
| `/api/tomori/agent/:id` | DELETE | 删除 Agent |
| `/api/tomori/chats` | GET | 聊天列表 |
| `/api/tomori/chat/:id` | DELETE | 删除聊天 |
| `/api/tomori/blacklist` | GET | 黑名单 |
| `/api/tomori/blacklist/clear` | POST | 清空黑名单 |

---

## 数据模型参考

### UserInfo

```json
{
  "user_id": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "username": "testuser",
  "email": "test@example.com",
  "phone": "",
  "bio": "个人签名",
  "avatar": "/uploads/avatars/xxx.png",
  "admin": false,
  "role": "normal_user",
  "banned": false,
  "follower_count": 10,
  "following_count": 5,
  "post_count": 20,
  "comment_count": 30,
  "received_like_count": 100,
  "received_dislike_count": 2,
  "is_followed": false
}
```

### PostDetail

```json
{
  "id": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "title": "帖子标题",
  "cover": "/uploads/covers/xxx.png",
  "user_id": 1,
  "author": { "user_id": 1, "username": "作者", "avatar": "..." },
  "tags": ["tag1", "tag2"],
  "category": "分类",
  "keywords": ["关键词"],
  "content": "Markdown 内容",
  "view_count": 100,
  "comment_count": 10,
  "like_count": 50,
  "dislike_count": 2,
  "favorite_count": 20,
  "share_count": 5,
  "public": true,
  "forbid_comment": false,
  "forbid_share": false,
  "is_liked": false,
  "is_disliked": false,
  "is_favorited": false,
  "env": { "ipv4": "...", "os": "..." }
}
```

### Comment（树形结构）

```json
{
  "comment_id": 1,
  "user_id": 1,
  "target_id": 1,
  "target_type": "post",
  "root_comment_id": 0,
  "parent_comment_id": 0,
  "content": "评论内容",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "author": { "user_id": 1, "username": "用户", "avatar": "..." },
  "like_count": 5,
  "dislike_count": 0,
  "reply_count": 3,
  "is_liked": false,
  "is_disliked": false,
  "reply_comments": [
    {
      "comment_id": 2,
      "parent_comment_id": 1,
      "content": "回复内容",
      "reply_comments": []
    }
  ]
}
```

- `root_comment_id = 0`：顶级评论
- `root_comment_id != 0`：子回复
- `reply_comments`：递归嵌套的子回复

### VideoDetail

```json
{
  "id": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "title": "视频标题",
  "description": "描述",
  "video_src_url": "/api/video/1/stream",
  "video_cover_url": "/api/video/1/cover",
  "duration": 120,
  "size": 104857600,
  "mime_type": "video/mp4",
  "user_id": 1,
  "author": { "user_id": 1, "username": "作者", "avatar": "..." },
  "category": "分类",
  "tags": ["tag1"],
  "view_count": 1000,
  "like_count": 50,
  "dislike_count": 2,
  "comment_count": 20,
  "favorite_count": 30,
  "share_count": 10,
  "public": true,
  "forbid_comment": false,
  "forbid_share": false,
  "status": "published",
  "is_liked": false,
  "is_disliked": false,
  "is_favorited": false
}
```

### FileDetail

```json
{
  "id": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "user_id": 1,
  "name": "文件名",
  "path": "/uploads/files/xxx.mp4",
  "size": 104857600,
  "mime_type": "video/mp4",
  "public": true,
  "duration": 120,
  "cover_url": "/api/video/1/cover",
  "view_count": 100,
  "like_count": 10,
  "dislike_count": 0,
  "favorite_count": 5,
  "share_count": 2
}
```

### AgentInfo

```json
{
  "id": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "user_id": 1,
  "name": "助手",
  "base_url": "https://api.openai.com/v1",
  "api_key": "***",
  "model_name": "gpt-4o-mini",
  "provider": "openai",
  "activate": true,
  "temperature": 0.7,
  "thinking": false,
  "tools": {},
  "prompts": {},
  "memories": {}
}
```

---

## 详细 API 文档

| 文档 | 说明 |
|------|------|
| [认证](auth.md) | 注册、登录、退出 |
| [用户](user.md) | 用户信息、资料修改、关注、设置 |
| [帖子](post.md) | 帖子 CRUD、搜索、AI 问答 |
| [评论](comment.md) | 评论 CRUD、树形结构、点赞/点踩 |
| [视频](video.md) | 视频上传、CRUD、流播放 |
| [文件上传](upload.md) | 通用文件分块上传 |
| [AI / Agent](ai.md) | Agent 管理、流式聊天 |
| [行情](market.md) | 市场指数实时/历史数据 |
| [管理后台](admin.md) | 超级管理员运营管理接口 |
