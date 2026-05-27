# 评论 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/comment/create` | POST | 创建评论 | 是 |
| `/api/comment/:id` | GET | 获取评论详情 | 否 |
| `/api/comment/:id` | DELETE | 删除评论 | 是 |
| `/api/comment/list` | GET | 获取评论列表（支持树形） | 否 |
| `/api/comment/:id/reply` | POST | 回复评论 | 是 |
| `/api/comment/:id/interact` | POST | 点赞/点踩评论 | 是 |
| `/api/user/comments` | GET | 获取当前用户评论列表 | 是 |

---

## 端点详情

### POST /api/comment/create

创建新评论。

**Request Body**：

```json
{
  "post_id": 1,
  "parent_id": 0,
  "content": "评论内容 (最多 10000 字，至少 1 字)"
}
```

- `parent_id`：`0`=顶级评论，非零=回复某条评论
- 如需引用其他用户，可在 `content` 中手动包含 `@username`

**Response**：

```json
{
  "data": {
    "id": 1
  }
}
```

### GET /api/comment/:id

获取单条评论详情。

**Response**：

```json
{
  "data": {
    "id": 1,
    "post_id": 1,
    "parent_id": 0,
    "content": "评论内容",
    "user_info": {
      "id": 1,
      "username": "testuser",
      "avatar_path": ""
    },
    "like_count": 3,
    "reply_count": 5,
    "is_liked": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

- `is_liked`：未登录或未操作时返回 `false`
- `reply_count`：该评论的回复数量（仅统计直接子评论，非全树）

### DELETE /api/comment/:id

删除评论。仅评论作者可操作。

### GET /api/comment/list

获取指定帖子的评论列表。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `post_id` | int | 是 | | 帖子 ID |
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

**响应格式为树形结构**：顶级评论的 `children` 数组包含直接回复。

**Response**：

```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "parent_id": 0,
        "content": "顶级评论",
        "user_info": { "id": 1, "username": "user1", "avatar_path": "" },
        "like_count": 5,
        "reply_count": 2,
        "is_liked": false,
        "children": [
          {
            "id": 2,
            "parent_id": 1,
            "content": "回复内容",
            "user_info": { "id": 2, "username": "user2", "avatar_path": "" },
            "reply_to": { "id": 1, "username": "user1" },
            "like_count": 1,
            "reply_count": 0,
            "is_liked": false
          }
        ],
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 10
  }
}
```

- 顶级评论的 `parent_id` 为 `0`
- 子评论的 `parent_id` 指向被回复的评论 ID
- `reply_to` 字段仅子评论有，包含被回复者的 `id` 和 `username`
- `children` 仅**下一级**回复（非全树展开）

### POST /api/comment/:id/reply

回复指定评论。

**Request Body**：

```json
{
  "content": "回复内容"
}
```

**Response**：

```json
{
  "data": {
    "id": 3
  }
}
```

### POST /api/comment/:id/interact

对评论进行交互操作。

**Request Body**：

```json
{
  "action_type": "like"
}
```

| `action_type` | 说明 |
|---------------|------|
| `like` | 点赞（toggle） |
| `dislike` | 点踩（toggle） |

### GET /api/user/comments

获取当前用户的所有评论列表。支持分页。

**Response**：

```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "post_id": 1,
        "content": "评论内容",
        "post_title": "帖子标题",
        "like_count": 3,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 10
  }
}
```
