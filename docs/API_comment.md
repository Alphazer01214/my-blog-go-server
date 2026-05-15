# 评论 API

- [通用说明](./API_common.md)

---

## 评论树结构说明

评论采用**两级树结构**：
- **根评论**：直接回复帖子，`root_comment_id = 0`, `parent_comment_id = 0`
- **子回复**：回复根评论或其他子回复，`root_comment_id` 指向根评论，`parent_comment_id` 指向直接父评论

响应中的 `reply_comments` 字段递归嵌套该评论的所有子回复。

---

## 1. 创建评论

**POST** `/api/comment` `[认证]`

### 请求体

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "post_id": 1,
  "content": "评论内容",
  "root_comment_id": 0,
  "parent_comment_id": 0
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `env` | EnvInfo | 否 | 客户端环境 |
| `post_id` | uint | 是 | 所属帖子 ID |
| `content` | string | 是 | 评论内容 |
| `root_comment_id` | uint | 否 | 根评论 ID（回复时使用，根评论传 0） |
| `parent_comment_id` | uint | 否 | 父评论 ID（回复时使用，根评论传 0） |

**场景示例：**
- 发根评论：`root_comment_id=0, parent_comment_id=0`
- 回复根评论：`root_comment_id=<根评论ID>, parent_comment_id=<根评论ID>`
- 回复子回复：`root_comment_id=<根评论ID>, parent_comment_id=<被回复的子回复ID>`

### 响应 data

```json
{
  "comment_id": 1,
  "user_id": 2,
  "post_id": 1,
  "root_comment_id": 0,
  "parent_comment_id": 0,
  "content": "评论内容",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "author": {
    "user_id": 2,
    "username": "commenter",
    "avatar": "",
    "..." : "..."
  },
  "likes": 0,
  "dislikes": 0,
  "replies": 0,
  "is_liked": false,
  "is_disliked": false,
  "reply_comments": []
}
```

---

## 2. 查询帖子评论

**GET** `/api/post/:id/comments` `[公开]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 帖子 ID |

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
      "comment_id": 1,
      "user_id": 2,
      "post_id": 1,
      "root_comment_id": 0,
      "parent_comment_id": 0,
      "content": "根评论内容",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "author": { "user_id": 2, "username": "user", "..." : "..." },
      "likes": 5,
      "dislikes": 0,
      "replies": 3,
      "is_liked": true,
      "is_disliked": false,
      "reply_comments": [
        {
          "comment_id": 2,
          "user_id": 3,
          "post_id": 1,
          "root_comment_id": 1,
          "parent_comment_id": 1,
          "content": "这是回复",
          "created_at": "2024-01-01T01:00:00Z",
          "updated_at": "2024-01-01T01:00:00Z",
          "author": { "user_id": 3, "username": "replier", "..." : "..." },
          "likes": 2,
          "dislikes": 0,
          "replies": 0,
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
}
```

> `items` 只包含根评论，每个根评论的 `reply_comments` 递归包含其所有子回复树。`is_liked` / `is_disliked` 在登录态下反映当前用户状态。

---

## 3. 查询评论的回复

**GET** `/api/comments/:id/replies` `[公开]`

> 用于**分页加载根评论的直接子回复**

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 根评论 ID |

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
      "comment_id": 2,
      "parent_comment_id": 1,
      "root_comment_id": 1,
      "content": "回复内容",
      "...": "...",
      "reply_comments": [
        {
          "comment_id": 3,
          "parent_comment_id": 2,
          "content": "对回复的回复",
          "...": "..."
        }
      ]
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 5
}
```

---

## 4. 删除评论

**DELETE** `/api/comment/:id` `[认证]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 评论 ID |

### 权限说明

- 根评论：仅作者本人可删除（会级联删除所有子回复）
- 子回复：仅作者本人可删除

### 响应 data

空对象

---

## 5. 评论点赞 / 取消点赞

**POST** `/api/comment/like` `[认证]`

> Toggle 模式：已赞则取消，未赞则点赞（有踩则先移除踩）

### 请求体

```json
{
  "comment_id": 1
}
```

### 响应 data

空对象

---

## 6. 评论点踩 / 取消点踩

**POST** `/api/comment/dislike` `[认证]`

> Toggle 模式：已踩则取消，未踩则点踩（有赞则先移除赞）

### 请求体

```json
{
  "comment_id": 1
}
```

### 响应 data

空对象
