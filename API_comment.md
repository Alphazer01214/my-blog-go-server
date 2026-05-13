# 评论系统 API 文档

## 基础信息

- Base URL（默认配置）: `http://localhost:2333/api`
- 认证方式沿用主系统: Cookie 双令牌（`access-token` + `refresh-token`）

---

## 接口总览

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| `POST` | `/api/comment` | 需要 | 创建评论/回复 |
| `DELETE` | `/api/comment/:id` | 需要 | 删除评论（仅作者） |
| `POST` | `/api/comment/like` | 需要 | 点赞/取消点赞评论 |
| `POST` | `/api/comment/dislike` | 需要 | 点踩/取消点踩评论 |
| `GET` | `/api/post/:id/comments` | 不需要 | 获取文章的根评论列表（含嵌套楼中楼） |
| `GET` | `/api/comments/:id/replies` | 不需要 | 获取某根评论下的嵌套回复树 |

---

## 数据模型

### 响应体 Comment

```json
{
  "comment_id": 1,
  "user_id": 1,
  "post_id": 1,
  "root_comment_id": 0,
  "parent_comment_id": 0,
  "content": "这是一条评论",
  "created_at": "2026-05-12T10:00:00Z",
  "updated_at": "2026-05-12T10:00:00Z",
  "author": {
    "user_id": 1,
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z",
    "username": "testuser",
    "email": "",
    "phone": "",
    "bio": "",
    "avatar": "https://example.com/avatar.png",
    "admin": false,
    "role": 0,
    "post_count": 15,
    "comment_count": 42
  },
  "likes": 0,
  "dislikes": 0,
  "is_liked": false,
  "is_disliked": false,
  "replies": 2,
  "reply_comments": [
    {
      "comment_id": 2,
      "user_id": 2,
      "post_id": 1,
      "root_comment_id": 1,
      "parent_comment_id": 1,
      "content": "回复根评论",
      "created_at": "2026-05-12T10:05:00Z",
      "updated_at": "2026-05-12T10:05:00Z",
      "author": { "user_id": 2, "username": "user2", "..." : "..." },
      "likes": 0,
      "dislikes": 0,
      "is_liked": true,
      "is_disliked": false,
      "replies": 1,
      "reply_comments": [
        { "...": "..." }
      ]
    }
  ]
}
```

字段说明:

| 字段 | 类型 | 说明 |
|------|------|------|
| `comment_id` | uint | 评论 ID |
| `user_id` | uint | 评论者用户 ID |
| `post_id` | uint | 所属文章 ID |
| `root_comment_id` | uint | 根评论 ID（`0` 表示该评论本身就是根评论） |
| `parent_comment_id` | uint | 父评论 ID（`0` 表示直接回复根评论） |
| `content` | string | 评论内容 |
| `author` | UserInfo | 评论者信息（完整 UserInfo 结构体） |
| `likes` | uint | 点赞数 |
| `dislikes` | uint | 点踩数 |
| `is_liked` | bool | 当前登录用户是否已点赞（未登录时为 `false`） |
| `is_disliked` | bool | 当前登录用户是否已点踩（未登录时为 `false`） |
| `replies` | uint | 直接子回复数 |
| `reply_comments` | array | **递归嵌套**的子回复列表，类型同 Comment |

> 注意：响应的 `id` 字段名是 `comment_id`（不是 entity 中的 `id`）。

---

## 接口详情

### 1) 创建评论

- 接口: `POST /api/comment`
- 认证: 需要 Cookie

请求体:

```json
{
  "post_id": 1,
  "content": "这是一条评论",
  "root_comment_id": 0,
  "parent_comment_id": 0,
  "env": {}
}
```

参数说明:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `post_id` | uint | 是 | 文章 ID |
| `content` | string | 是 | 评论内容 |
| `root_comment_id` | uint | 否 | 默认 `0` 表示是根评论。填非零整数则为对应根评论的 ID |
| `parent_comment_id` | uint | 否 | 默认 `0` 表示无父评论。填非零整数则为父评论 ID |
| `env` | object | 否 | 环境信息 |

成功响应:

```json
{
  "code": 0,
  "data": {
    "comment_id": 1,
    "user_id": 1,
    "post_id": 1,
    "root_comment_id": 0,
    "parent_comment_id": 0,
    "content": "这是一条评论",
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z",
    "author": {
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
      "post_count": 5,
      "comment_count": 10
    },
    "likes": 0,
    "dislikes": 0,
    "is_liked": false,
    "is_disliked": false,
    "replies": 0,
    "reply_comments": []
  },
  "msg": "comment success"
}
```

嵌套回复示例:

- **回复根评论**: `root_comment_id = 根评论ID`, `parent_comment_id = 根评论ID`
- **回复子评论**: `root_comment_id = 根评论ID`, `parent_comment_id = 被回复的子评论ID`

---

### 2) 删除评论

- 接口: `DELETE /api/comment/:id`
- 认证: 需要 Cookie（仅评论作者可删）

路径参数:

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 评论 ID |

成功响应:

```json
{
  "code": 0,
  "data": {},
  "msg": "delete success"
}
```

> 删除根评论会级联删除所有子回复；删除子回复仅删除该条。

---

### 3) 点赞评论

- 接口: `POST /api/comment/like`
- 认证: 需要 Cookie

请求体:

```json
{
  "comment_id": 1
}
```

> 切换逻辑：已点赞 → 取消点赞（`likes` 减 1）；已点踩 → 移除点踩（`dislikes` 减 1）+ 点赞（`likes` 加 1）；无操作 → 直接点赞（`likes` 加 1）。

成功响应:

```json
{
  "code": 0,
  "data": {},
  "msg": "like success"
}
```

---

### 4) 点踩评论

- 接口: `POST /api/comment/dislike`
- 认证: 需要 Cookie

请求体:

```json
{
  "comment_id": 1
}
```

> 切换逻辑与点赞对称。

成功响应:

```json
{
  "code": 0,
  "data": {},
  "msg": "dislike success"
}
```

---

### 5) 获取文章评论列表（含楼中楼）

- 接口: `GET /api/post/:id/comments`
- 认证: 不需要

路径参数:

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 文章 ID |

查询参数:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码（默认 `1`） |
| `page_size` | int | 否 | 每页数量（默认 `20`，最大 `100`） |

> 返回该文章的**根评论**列表（`root_comment_id = 0`），按创建时间倒序。
> 每条根评论的 `reply_comments` 字段**递归**包含其所有嵌套子回复，前端可直接渲染楼中楼。

成功响应:

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "comment_id": 1,
        "user_id": 1,
        "post_id": 1,
        "root_comment_id": 0,
        "parent_comment_id": 0,
        "content": "这是根评论",
        "created_at": "2026-05-12T10:00:00Z",
        "updated_at": "2026-05-12T10:00:00Z",
        "author": { "user_id": 1, "username": "user1", "avatar": "...", "..." : "..." },
        "likes": 5,
        "dislikes": 1,
        "is_liked": true,
        "is_disliked": false,
        "replies": 3,
        "reply_comments": [
          {
            "comment_id": 2,
            "user_id": 2,
            "post_id": 1,
            "root_comment_id": 1,
            "parent_comment_id": 1,
            "content": "回复根评论",
            "created_at": "2026-05-12T10:05:00Z",
            "updated_at": "2026-05-12T10:05:00Z",
            "author": { "user_id": 2, "username": "user2", "avatar": "...", "..." : "..." },
            "likes": 2,
            "dislikes": 0,
            "is_liked": false,
            "is_disliked": false,
            "replies": 1,
            "reply_comments": [
              {
                "comment_id": 4,
                "user_id": 1,
                "post_id": 1,
                "root_comment_id": 1,
                "parent_comment_id": 2,
                "content": "回复楼中楼",
                "created_at": "2026-05-12T10:10:00Z",
                "updated_at": "2026-05-12T10:10:00Z",
                "author": { "user_id": 1, "username": "user1", "avatar": "...", "..." : "..." },
                "likes": 1,
                "dislikes": 0,
                "is_liked": false,
                "is_disliked": true,
                "replies": 0,
                "reply_comments": []
              }
            ]
          },
          {
            "comment_id": 3,
            "user_id": 3,
            "post_id": 1,
            "root_comment_id": 1,
            "parent_comment_id": 1,
            "content": "另一个人也回复根评论",
            "created_at": "2026-05-12T10:06:00Z",
            "updated_at": "2026-05-12T10:06:00Z",
            "author": { "user_id": 3, "username": "user3", "avatar": "...", "..." : "..." },
            "likes": 0,
            "dislikes": 0,
            "is_liked": false,
            "is_disliked": false,
            "replies": 0,
            "reply_comments": []
          }
        ]
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 1
  },
  "msg": "query success"
}
```

---

### 6) 获取评论回复树

- 接口: `GET /api/comments/:id/replies`
- 认证: 不需要

路径参数:

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 根评论 ID |

查询参数:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码（默认 `1`） |
| `page_size` | int | 否 | 每页数量（默认 `20`，最大 `100`） |

> 返回该根评论下的直接回复树（递归包含嵌套子回复），按创建时间正序。
> 注意：返回的 items 是直接 `reply_comments`，而不是根评论本身。

成功响应:

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "comment_id": 2,
        "user_id": 2,
        "post_id": 1,
        "root_comment_id": 1,
        "parent_comment_id": 1,
        "content": "回复根评论",
        "created_at": "2026-05-12T10:05:00Z",
        "updated_at": "2026-05-12T10:05:00Z",
        "author": { "user_id": 2, "username": "user2", "avatar": "...", "..." : "..." },
        "likes": 2,
        "dislikes": 0,
        "is_liked": false,
        "is_disliked": false,
        "replies": 1,
        "reply_comments": [
          {
            "comment_id": 4,
            "user_id": 1,
            "post_id": 1,
            "root_comment_id": 1,
            "parent_comment_id": 2,
            "content": "回复楼中楼",
            "created_at": "2026-05-12T10:10:00Z",
            "updated_at": "2026-05-12T10:10:00Z",
            "author": { "user_id": 1, "username": "user1", "avatar": "...", "..." : "..." },
            "likes": 1,
            "dislikes": 0,
            "is_liked": false,
            "is_disliked": true,
            "replies": 0,
            "reply_comments": []
          }
        ]
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 3
  },
  "msg": "query success"
}
```

---

## 评论嵌套结构说明

```
文章
└── 根评论 (root_comment_id = 0, parent_comment_id = 0)
    ├── 回复1 (root_comment_id = 根评论ID, parent_comment_id = 根评论ID)
    │   └── 回复1.1 (root_comment_id = 根评论ID, parent_comment_id = 回复1的ID)
    └── 回复2 (root_comment_id = 根评论ID, parent_comment_id = 根评论ID)
```

- 前端获取根评论列表（含嵌套树）: `GET /api/post/:id/comments`
- 前端获取某个根评论的回复树: `GET /api/comments/:root_comment_id/replies`
- 渲染时直接按 `reply_comments` 递归展示楼中楼，无需额外请求
