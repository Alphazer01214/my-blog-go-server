# 管理后台 API

> 仅 **超级管理员**（`takamatsu_tomori`）角色可访问。  
> 所有端点均需 JWT 认证 + 管理员角色校验。

## 目录

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/tomori/stats` | GET | 获取系统统计数据 |
| **用户管理** | | |
| `/api/tomori/users` | GET | 获取所有用户列表 |
| `/api/tomori/user/ban` | POST | 封禁/解封用户 |
| `/api/tomori/user/role` | POST | 设置用户角色 |
| `/api/tomori/user/reset_password` | POST | 重置用户密码 |
| `/api/tomori/user/:id` | DELETE | 删除用户 |
| **帖子管理** | | |
| `/api/tomori/posts` | GET | 获取所有帖子列表 |
| `/api/tomori/post/:id` | DELETE | 删除任意帖子 |
| **评论管理** | | |
| `/api/tomori/comments` | GET | 获取所有评论列表 |
| `/api/tomori/comment/:id` | DELETE | 删除任意评论 |
| **文件管理** | | |
| `/api/tomori/files` | GET | 获取所有文件列表 |
| `/api/tomori/file/:id` | DELETE | 删除任意文件 |
| **视频管理** | | |
| `/api/tomori/videos` | GET | 获取所有视频列表 |
| `/api/tomori/video/:id` | DELETE | 删除任意视频 |
| **AI 智能体管理** | | |
| `/api/tomori/agents` | GET | 获取所有 Agent 列表 |
| `/api/tomori/agent/:id` | DELETE | 删除任意 Agent |
| **聊天会话管理** | | |
| `/api/tomori/chats` | GET | 获取所有聊天记录 |
| `/api/tomori/chat/:chat_id` | DELETE | 删除聊天会话 |
| **系统维护** | | |
| `/api/tomori/blacklist` | GET | 获取 Token 黑名单 |
| `/api/tomori/blacklist/clear` | POST | 清空黑名单 |

---

## 端点详情

### GET /api/tomori/stats

获取系统统计数据面板。

**Response**：

```json
{
  "code": 0,
  "data": {
    "user_count": 100,
    "post_count": 500,
    "comment_count": 2000,
    "file_count": 300,
    "video_count": 150,
    "agent_count": 50,
    "chat_count": 150
  },
  "msg": "Success"
}
```

---

## 用户管理

### GET /api/tomori/users

获取所有用户列表。支持分页。

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
        "user_id": 1,
        "username": "testuser",
        "email": "test@example.com",
        "phone": "",
        "bio": "",
        "avatar": "",
        "admin": false,
        "role": "normal_user",
        "banned": false,
        "follower_count": 0,
        "following_count": 0,
        "post_count": 3,
        "comment_count": 12,
        "received_like_count": 42,
        "received_dislike_count": 2,
        "is_followed": false,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 100
  },
  "msg": "Success"
}
```

### POST /api/tomori/user/ban

封禁或解封用户。

**Request Body**：

```json
{
  "user_id": 2,
  "banned": true
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `user_id` | uint | 是 | 目标用户 ID |
| `banned` | bool | 是 | `true`=封禁，`false`=解封 |

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

### POST /api/tomori/user/role

设置用户角色。

**Request Body**：

```json
{
  "user_id": 2,
  "role": "moderator"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `user_id` | uint | 是 | 目标用户 ID |
| `role` | string | 是 | 角色值（见枚举） |

### POST /api/tomori/user/reset_password

重置用户密码。

**Request Body**：

```json
{
  "user_id": 2,
  "new_password": "newpassword123"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `user_id` | uint | 是 | 目标用户 ID |
| `new_password` | string | 是 | 新密码 |

### DELETE /api/tomori/user/:id

删除用户。

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

---

## 帖子管理

### GET /api/tomori/posts

获取所有帖子列表。支持分页。响应格式同帖子列表接口。

### DELETE /api/tomori/post/:id

删除指定帖子。超级管理员可以删除任意帖子。

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

---

## 评论管理

### GET /api/tomori/comments

获取所有评论列表。支持分页。响应格式同评论列表接口。

### DELETE /api/tomori/comment/:id

删除指定评论。

---

## 文件管理

### GET /api/tomori/files

获取所有文件列表。支持分页。响应格式同文件列表接口。

### DELETE /api/tomori/file/:id

删除指定文件。

---

## 视频管理

### GET /api/tomori/videos

获取所有视频列表。支持分页。

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
        "title": "视频标题",
        "description": "视频描述",
        "video_src_url": "/api/video/1/stream",
        "video_cover_url": "/api/video/1/cover",
        "duration": 120,
        "size": 10485760,
        "mime_type": "video/mp4",
        "user_id": 1,
        "author": {
          "user_id": 1,
          "username": "testuser",
          "avatar": ""
        },
        "category": "技术",
        "tags": ["Go", "Vue"],
        "view_count": 100,
        "like_count": 10,
        "dislike_count": 0,
        "comment_count": 5,
        "favorite_count": 3,
        "share_count": 2,
        "public": true,
        "forbid_comment": false,
        "forbid_share": false,
        "status": "published",
        "is_liked": false,
        "is_disliked": false,
        "is_favorited": false
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 150
  },
  "msg": "Success"
}
```

### DELETE /api/tomori/video/:id

删除指定视频。

---

## AI 智能体管理

### GET /api/tomori/agents

获取所有 Agent 列表。

**Response**：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "agent_id": 1,
        "user_id": 1,
        "username": "testuser",
        "name": "助手",
        "model_name": "gpt-4o-mini",
        "provider": "openai",
        "activate": false,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 50
  },
  "msg": "Success"
}
```

### DELETE /api/tomori/agent/:id

删除指定 Agent。

---

## 聊天会话管理

### GET /api/tomori/chats

获取所有聊天会话。支持分页。

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
        "title": "对话标题",
        "messages": []
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 150
  },
  "msg": "Success"
}
```

### DELETE /api/tomori/chat/:chat_id

删除指定聊天会话。

---

## 系统维护

### GET /api/tomori/blacklist

获取 Token 黑名单列表（从 Redis 读取）。支持分页。

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
        "id": "token:blacklist:eyJhbGciOi...",
        "token": "eyJhbGciOi...",
        "created_at": "5m30s"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 30
  },
  "msg": "Success"
}
```

> `created_at` 字段显示的是 Token 剩余过期时间（TTL）。

### POST /api/tomori/blacklist/clear

清空 Token 黑名单。

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```
