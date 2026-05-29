# 帖子 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/create` | POST | 创建帖子 | 是 |
| `/api/post/:id` | GET | 获取帖子详情 | 是 |
| `/api/post` | GET | 获取帖子列表 | 是 |
| `/api/post/search` | GET | 搜索帖子 | 是 |
| `/api/update` | POST | 更新帖子 | 是 |
| `/api/post/:id` | DELETE | 删除帖子 | 是 |
| `/api/post/like` | POST | 点赞/取消点赞 | 是 |
| `/api/post/dislike` | POST | 点踩/取消点踩 | 是 |
| `/api/post/favorite` | POST | 收藏/取消收藏 | 是 |
| `/api/post/share` | POST | 分享 | 是 |
| `/api/post/ask` | POST | AI 问答（SSE 流式） | 是 |
| `/api/favorites` | GET | 获取收藏列表 | 是 |

---

## 端点详情

### POST /api/create

创建新帖子。

**Request Body**：

```json
{
  "env": {
    "ipv4": "192.168.1.1",
    "ipv6": "",
    "os": "Windows 11",
    "device_info": "Chrome 120"
  },
  "title": "帖子标题",
  "cover": "/uploads/covers/xxx.png",
  "category": "tech",
  "tags": ["Go", "Web"],
  "keywords": ["golang", "gin"],
  "content": "# 帖子内容\n\nMarkdown 格式",
  "public": true,
  "forbid_comment": false,
  "forbid_share": false
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `env` | object | 环境信息（可选） |
| `title` | string | 帖子标题 |
| `cover` | string | 封面路径（可选） |
| `category` | string | 分类标识（可选） |
| `tags` | array | 标签数组（可选） |
| `keywords` | array | 关键词数组（可选） |
| `content` | string | 帖子内容（Markdown） |
| `public` | bool | 是否公开（默认 false） |
| `forbid_comment` | bool | 是否禁止评论（默认 false） |
| `forbid_share` | bool | 是否禁止分享（默认 false） |

**Response**：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "title": "帖子标题",
    "cover": "/uploads/covers/xxx.png",
    "user_id": 1,
    "tags": ["Go", "Web"],
    "category": "tech",
    "keywords": ["golang", "gin"],
    "content": "# 帖子内容\n\nMarkdown 格式",
    "view_count": 0,
    "comment_count": 0,
    "like_count": 0,
    "dislike_count": 0,
    "favorite_count": 0,
    "share_count": 0,
    "public": true,
    "forbid_comment": false,
    "forbid_share": false,
    "is_liked": false,
    "is_disliked": false,
    "is_favorited": false,
    "author": {
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
      "post_count": 1,
      "comment_count": 0,
      "received_like_count": 0,
      "received_dislike_count": 0,
      "is_followed": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    "env": { ... }
  },
  "msg": "Success"
}
```

### GET /api/post/:id

获取帖子详情。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "title": "帖子标题",
    "cover": "/uploads/covers/xxx.png",
    "user_id": 1,
    "tags": ["Go", "Web"],
    "category": "tech",
    "keywords": ["golang", "gin"],
    "content": "# 帖子内容\n\nMarkdown 格式",
    "view_count": 100,
    "comment_count": 5,
    "like_count": 20,
    "dislike_count": 1,
    "favorite_count": 8,
    "share_count": 3,
    "public": true,
    "forbid_comment": false,
    "forbid_share": false,
    "is_liked": false,
    "is_disliked": false,
    "is_favorited": false,
    "author": {
      "user_id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "phone": "",
      "bio": "",
      "avatar": "",
      "admin": false,
      "role": "normal_user",
      "banned": false,
      "follower_count": 10,
      "following_count": 5,
      "post_count": 3,
      "comment_count": 12,
      "received_like_count": 42,
      "received_dislike_count": 2,
      "is_followed": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    "env": { ... }
  },
  "msg": "Success"
}
```

### GET /api/post

获取帖子列表。需要登录。

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
        "title": "帖子标题",
        "cover": "",
        "user_id": 1,
        "tags": ["Go"],
        "category": "tech",
        "keywords": [],
        "content": "内容摘要...",
        "view_count": 100,
        "comment_count": 5,
        "like_count": 20,
        "dislike_count": 1,
        "favorite_count": 8,
        "share_count": 3,
        "public": true,
        "forbid_comment": false,
        "forbid_share": false,
        "is_liked": false,
        "is_disliked": false,
        "is_favorited": false,
        "author": {
          "user_id": 1,
          "username": "testuser",
          "avatar": "",
          "role": "normal_user"
        },
        "env": { ... }
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 42
  },
  "msg": "Success"
}
```

### GET /api/post/search

搜索帖子。需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `keyword` | string | 否 | | 搜索关键词 |
| `tag` | string | 否 | | 按标签筛选 |
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

**Response**：格式同帖子列表。

### POST /api/update

更新帖子。仅作者可操作。需要登录。

**Request Body**（全部可选）：

```json
{
  "post_id": 1,
  "env": { ... },
  "title": "新标题",
  "cover": "新封面",
  "category": "new_category",
  "tags": ["newtag"],
  "keywords": ["new_keyword"],
  "content": "新内容",
  "public": false,
  "forbid_comment": false,
  "forbid_share": false
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `post_id` | uint | 帖子 ID（必填） |
| 其他字段 | 同创建 | 所有字段可选 |

**Response**：

```json
{
  "code": 0,
  "data": { ... },
  "msg": "Success"
}
```

### DELETE /api/post/:id

删除帖子。仅作者和管理员可操作。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

---

## 交互操作

### POST /api/post/like

点赞或取消点赞帖子（toggle 切换）。需要登录。

**Request Body**：

```json
{
  "post_id": 1
}
```

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

### POST /api/post/dislike

点踩或取消点踩帖子（toggle 切换）。需要登录。

**Request Body**：

```json
{
  "post_id": 1
}
```

### POST /api/post/favorite

收藏或取消收藏帖子（toggle 切换）。需要登录。

**Request Body**：

```json
{
  "post_id": 1
}
```

### GET /api/favorites

获取当前用户的收藏列表。需要登录。

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
        "title": "收藏的帖子标题",
        "cover": "",
        "user_id": 1,
        "tags": ["Go"],
        "category": "tech",
        "keywords": [],
        "content": "内容摘要...",
        "view_count": 100,
        "comment_count": 5,
        "like_count": 20,
        "dislike_count": 1,
        "favorite_count": 8,
        "share_count": 3,
        "public": true,
        "forbid_comment": false,
        "forbid_share": false,
        "is_liked": false,
        "is_disliked": false,
        "is_favorited": true,
        "author": {
          "user_id": 1,
          "username": "testuser",
          "avatar": "",
          "role": "normal_user"
        },
        "env": { ... }
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 10
  },
  "msg": "query favorites success"
}
```

### POST /api/post/share

分享帖子（计数+1，非 toggle）。需要登录。

**Request Body**：

```json
{
  "post_id": 1,
  "share_to": "weibo",
  "share_message": "分享理由"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `post_id` | uint | 是 | 帖子 ID |
| `share_to` | string | 否 | 分享目标 |
| `share_message` | string | 否 | 分享附言 |

---

## AI 问答

### POST /api/post/ask

对帖子进行 AI 问答（SSE 流式输出）。需要登录。

**Request Body**：

```json
{
  "chat_id": "uuid (前端生成，可选)",
  "post_id": 1,
  "prompt": "这篇文章主要讲了什么？",
  "selected_text": "（可选，mode=selected 时需要）",
  "mode": "ask",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxx",
  "model": "gpt-4o-mini"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `chat_id` | string | 否 | 前端 UUID，用于多轮对话。首次不传，后续传入相同值 |
| `post_id` | uint | 是 | 帖子 ID |
| `prompt` | string | 否 | 用户提问内容 |
| `selected_text` | string | 否 | 划词内容（mode=selected 时需要） |
| `mode` | string | 否 | 模式：`summarize`（总结）、`ask`（提问）、`selected`（划词提问） |
| `base_url` | string | 否 | AI 服务地址 |
| `api_key` | string | 否 | AI API Key |
| `model` | string | 否 | 模型名称 |

**SSE 响应格式**：

首帧（包含历史）：
```json
data: {"agent_id":0,"chat_id":"uuid","history":[{"role":"user","content":"你好"}],"status":true}
```

流式块：
```json
data: {"agent_id":0,"chat_id":"uuid","content":"这是回答内容...","status":true}
```

完成帧：
```json
data: {"agent_id":0,"chat_id":"uuid","content":"","status":true,"message":"done"}
```

错误帧：
```json
data: {"agent_id":0,"chat_id":"uuid","content":"错误信息","status":false}
```

**注意**：问答历史存储在 Redis（key: `session_<chatId>`）。
