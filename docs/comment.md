# 评论 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/comment` | POST | 创建评论 | 是 |
| `/api/comment/:id` | DELETE | 删除评论 | 是 |
| `/api/post/:id/comments` | GET | 获取帖子评论列表 | 是 |
| `/api/video/:id/comments` | GET | 获取视频评论列表 | 是 |
| `/api/comment/like` | POST | 点赞/取消点赞评论 | 是 |
| `/api/comment/dislike` | POST | 点踩/取消点踩评论 | 是 |

---

## 端点详情

### POST /api/comment

创建新评论。需要登录。

**Request Body**：

```json
{
  "env": {
    "ipv4": "192.168.1.1",
    "ipv6": "",
    "os": "Windows 11",
    "device_info": "Chrome 120"
  },
  "post_id": 1,
  "target_type": "post",
  "target_id": 1,
  "content": "评论内容 (最多 10000 字，至少 1 字)",
  "root_comment_id": 0,
  "parent_comment_id": 0
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `env` | object | 否 | 环境信息 |
| `post_id` | uint | 否 | 关联帖子 ID |
| `target_type` | string | 是 | 目标类型：`post`、`video` |
| `target_id` | uint | 是 | 目标 ID（帖子或视频 ID） |
| `content` | string | 是 | 评论内容 |
| `root_comment_id` | uint | 否 | 根评论 ID（0=顶级评论） |
| `parent_comment_id` | uint | 否 | 父评论 ID（0=顶级评论） |

- `root_comment_id`：顶级评论为 `0`，回复时传入顶级评论 ID
- `parent_comment_id`：顶级评论为 `0`，回复时传入被回复的评论 ID
- 如需引用其他用户，可在 `content` 中手动包含 `@username`

**Response**：

```json
{
  "code": 0,
  "data": {
    "comment_id": 1,
    "user_id": 1,
    "target_id": 1,
    "target_type": "post",
    "root_comment_id": 0,
    "parent_comment_id": 0,
    "content": "评论内容",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "env": { ... },
    "author": {
      "user_id": 1,
      "username": "testuser",
      "avatar": ""
    },
    "like_count": 0,
    "dislike_count": 0,
    "reply_count": 0,
    "is_liked": false,
    "is_disliked": false,
    "reply_comments": []
  },
  "msg": "Success"
}
```

### DELETE /api/comment/:id

删除评论。仅评论作者可操作。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

---

## 评论列表

### GET /api/post/:id/comments

获取指定帖子的评论列表。需要登录。

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
        "comment_id": 1,
        "user_id": 1,
        "target_id": 1,
        "target_type": "post",
        "root_comment_id": 0,
        "parent_comment_id": 0,
        "content": "顶级评论",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "env": { ... },
        "author": {
          "user_id": 1,
          "username": "user1",
          "avatar": ""
        },
        "like_count": 5,
        "dislike_count": 0,
        "reply_count": 2,
        "is_liked": false,
        "is_disliked": false,
        "reply_comments": [
          {
            "comment_id": 2,
            "user_id": 2,
            "target_id": 1,
            "target_type": "post",
            "root_comment_id": 1,
            "parent_comment_id": 1,
            "content": "回复内容",
            "created_at": "2024-01-01T00:01:00Z",
            "updated_at": "2024-01-01T00:01:00Z",
            "env": { ... },
            "author": {
              "user_id": 2,
              "username": "user2",
              "avatar": ""
            },
            "like_count": 1,
            "dislike_count": 0,
            "reply_count": 0,
            "is_liked": false,
            "is_disliked": false,
            "reply_comments": []
          }
        ]
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 10
  },
  "msg": "Success"
}
```

| 字段 | 说明 |
|------|------|
| `comment_id` | 评论 ID |
| `user_id` | 评论者 ID |
| `target_id` | 目标 ID |
| `target_type` | 目标类型 |
| `root_comment_id` | 根评论 ID（0=顶级） |
| `parent_comment_id` | 父评论 ID（0=顶级） |
| `content` | 评论内容 |
| `author` | 作者信息（UserInfo 子集） |
| `like_count` | 点赞数 |
| `dislike_count` | 点踩数 |
| `reply_count` | 直接子评论数 |
| `is_liked` | 当前用户是否点赞 |
| `is_disliked` | 当前用户是否点踩 |
| `reply_comments` | 回复评论列表（仅下一级） |

### GET /api/video/:id/comments

获取指定视频的评论列表。格式同上。

---

## 交互操作

### POST /api/comment/like

点赞或取消点赞评论（toggle 切换）。需要登录。

**Request Body**：

```json
{
  "comment_id": 1
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

### POST /api/comment/dislike

点踩或取消点踩评论（toggle 切换）。需要登录。

**Request Body**：

```json
{
  "comment_id": 1
}
```

**Response**：同上。
