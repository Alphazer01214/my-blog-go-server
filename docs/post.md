# 帖子 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/post/create` | POST | 创建帖子 | 是 |
| `/api/post/:id` | GET | 获取帖子详情 | 否 |
| `/api/post/:id` | PUT | 更新帖子 | 是 |
| `/api/post/:id` | DELETE | 删除帖子 | 是 |
| `/api/post/list` | GET | 获取帖子列表（支持分类筛选） | 否 |
| `/api/post/:id/interact` | POST | 点赞/点踩/收藏/分享 | 是 |
| `/api/post/:id/favorited` | GET | 查询是否收藏 | 是 |
| `/api/post/search` | GET | 搜索帖子 | 否 |
| `/api/post/ask` | POST | AI 问答（SSE 流式） | 否 |

---

## 端点详情

### POST /api/post/create

创建新帖子。

**Request Body**：

```json
{
  "title": "帖子标题 (最多 200 字)",
  "content": "帖子内容 (Markdown)",
  "tags": ["tag1", "tag2"],
  "category": "分类标识",
  "cover_path": "可选封面路径",
  "type": "text"
}
```

- `type`：帖子类型标识

**Response**：

```json
{
  "data": {
    "id": 1
  }
}
```

### GET /api/post/:id

获取帖子详情。

`category`|`tags`|`user_info`|`is_liked`|`is_favorited`|`like_count`|`comment_count`|`favorite_count`|`share_count`

**Response**：

```json
{
  "data": {
    "id": 1,
    "title": "标题",
    "content": "内容",
    "tags": ["tag1"],
    "category": "分类",
    "cover_path": "",
    "type": "text",
    "like_count": 10,
    "comment_count": 3,
    "favorite_count": 5,
    "share_count": 1,
    "is_liked": false,
    "is_favorited": false,
    "is_owner": false,
    "user_info": {
      "id": 1,
      "username": "作者",
      "avatar_path": "",
      "role_type": "normal_user"
    },
    "role_type": "normal_user",
    "status": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

- `is_liked`、`is_favorited`：未登录或未操作时返回 `false`
- `is_owner`：当前用户是否是帖子作者

### PUT /api/post/:id

更新帖子。仅作者可操作。

**Request Body**（全部可选）：

```json
{
  "title": "新标题",
  "content": "新内容",
  "tags": ["newtag"],
  "category": "新分类",
  "cover_path": "新封面",
  "type": "text"
}
```

### DELETE /api/post/:id

删除帖子。仅作者和管理员可操作。

**Response**：标准成功响应。

### GET /api/post/list

获取帖子列表。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `category` | string | 否 | 全部 | 按分类筛选 |
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |
| `user_id` | int | 否 | 0 | 筛选指定用户的帖子 |

**Response**：

```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "title": "标题",
        "tags": ["tag1"],
        "category": "分类",
        "cover_path": "",
        "type": "text",
        "like_count": 10,
        "comment_count": 3,
        "user_info": {
          "id": 1,
          "username": "作者",
          "avatar_path": ""
        },
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 42
  }
}
```

### POST /api/post/:id/interact

对帖子进行交互操作。

**Request Body**：

```json
{
  "action_type": "like"
}
```

| `action_type` | 说明 | 可重复调用 |
|---------------|------|-----------|
| `like` | 点赞（再次调用取消） | 是（toggle） |
| `dislike` | 点踩（再次调用取消） | 是（toggle） |
| `favorite` | 收藏（再次调用取消） | 是（toggle） |
| `share` | 分享计数+1 | 否（计数递增） |

**Response**：

```json
{
  "data": {
    "action_type": "like",
    "is_active": true
  }
}
```

- `is_active`：`true`=已点赞/收藏，`false`=已取消

### GET /api/post/:id/favorited

查询当前用户是否收藏了该帖子。

**Response**：

```json
{
  "data": {
    "favorited": true
  }
}
```

### GET /api/post/search

搜索帖子。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `q` | string | 是 | | 搜索关键词 |
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

### POST /api/post/ask

对帖子进行 AI 问答（SSE 流式输出）。

**Request Body**：

```json
{
  "post_id": 1,
  "mode": "ask",
  "question": "这篇文章主要讲了什么？",
  "selected_text": "（可选，mode=selected 时需要）",
  "chat_id": "uuid (前端生成，用于多轮对话)"
}
```

**Mode 说明**：

| mode | 说明 |
|------|------|
| `summarize` | 总结文章，忽略 `question` |
| `ask` | 对文章提问 |
| `selected` | 划词提问，需传 `selected_text` |

**`chat_id`**：由前端通过 `crypto.randomUUID()` 生成。首次问答时生成新 UUID，后续轮次传入相同的 `chat_id` 以维持对话上下文。

**SSE 响应格式**：

```json
data: {"chat_id":"uuid","content":"思考中","status":true}
data: {"chat_id":"uuid","content":"这是回答内容...","status":true}
data: {"chat_id":"uuid","content":"","status":true,"message":"done"}
```

- `chat_id`：传入的 `chat_id` 原样返回
- 流式输出中 `content` 为 **增量文本**（非完整内容拼接）
- 最后一个事件 `message` = `"done"` 表示流结束
- 错误时：`{"chat_id":"uuid","content":"错误信息","status":false}`

**注意**：该接口不使用数据库中的 session 表，问答历史存储在 Redis（key: `session_<chatId>`）。
