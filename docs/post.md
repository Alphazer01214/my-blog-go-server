# Post API

| Method | Endpoint            | Description          |
| ------ | ------------------- | -------------------- |
| GET    | /api/post/:id       | Get post detail      |
| GET    | /api/post           | List posts           |
| POST   | /api/create         | Create post          |
| POST   | /api/update         | Update post          |
| DELETE | /api/post/:id       | Delete post          |
| POST   | /api/post/like      | Like post            |
| POST   | /api/post/dislike   | Dislike post         |
| POST   | /api/post/favorite  | Favorite post        |
| POST   | /api/post/share     | Share post           |
| GET    | /api/post/search    | Search posts         |
| POST   | /api/post/ask       | Ask AI about post    |

---

## GET /api/post/:id

Get a single post by ID.

**Path Parameters:**
- `id` (integer) - Post ID

**Response:**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "title": "Trading Strategy",
    "cover": "https://example.com/cover.jpg",
    "user_id": 1,
    "author": { "...UserInfo..." },
    "tags": ["strategy", "stocks"],
    "category": "trading",
    "keywords": ["moving average"],
    "content": "Full post content here...",
    "view_count": 150,
    "comment_count": 8,
    "like_count": 25,
    "dislike_count": 1,
    "favorite_count": 10,
    "share_count": 3,
    "public": true,
    "forbid_comment": false,
    "forbid_share": false,
    "is_liked": false,
    "is_disliked": false,
    "is_favorited": false,
    "env": { "ipv4": "192.168.1.1", "os": "Windows 11", "device_info": "Chrome 120" }
  },
  "msg": "success"
}
```

---

## GET /api/post

List posts with pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...PostDetail..." }
  ],
  "msg": "success"
}
```

---

## POST /api/create

Create a new post. Requires authentication.

**Request Body:**

```json
{
  "title": "string (required)",
  "content": "string (required)",
  "cover": "string (optional)",
  "category": "string (optional)",
  "tags": ["string"] (optional),
  "keywords": ["string"] (optional),
  "public": true,
  "forbid_comment": false,
  "forbid_share": false,
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" }
}
```

**Response:**

```json
{
  "code": 0,
  "data": { "...PostDetail..." },
  "msg": "success"
}
```

---

## POST /api/update

Update an existing post. Requires authentication (must be author).

**Request Body:**

```json
{
  "post_id": 1,
  "title": "string (optional)",
  "content": "string (optional)",
  "cover": "string (optional)",
  "category": "string (optional)",
  "tags": ["string"] (optional),
  "keywords": ["string"] (optional),
  "public": true,
  "forbid_comment": false,
  "forbid_share": false,
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" }
}
```

**Response:**

```json
{
  "code": 0,
  "data": { "...PostDetail..." },
  "msg": "success"
}
```

---

## DELETE /api/post/:id

Delete a post. Requires authentication (must be author or admin).

**Path Parameters:**
- `id` (integer) - Post ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "deleted"
}
```

---

## POST /api/post/like

Like a post. Requires authentication.

**Request Body:**

```json
{
  "post_id": 1
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "liked"
}
```

---

## POST /api/post/dislike

Dislike a post. Requires authentication.

**Request Body:**

```json
{
  "post_id": 1
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "disliked"
}
```

---

## POST /api/post/favorite

Favorite a post. Requires authentication.

**Request Body:**

```json
{
  "post_id": 1
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "favorited"
}
```

---

## POST /api/post/share

Share a post. Requires authentication.

**Request Body:**

```json
{
  "post_id": 1,
  "share_to": "string (optional)",
  "share_message": "string (optional)"
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "shared"
}
```

---

## GET /api/post/search

Search posts by keyword.

**Query Parameters:**
- `keyword` (string, required) - Search keyword
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...PostDetail..." }
  ],
  "msg": "success"
}
```

---

## POST /api/post/ask

对帖子进行 AI 问答（SSE 流式）。需要登录。

**Request Body:**

```json
{
  "post_id": 1,
  "session_id": "uuid（前端生成，新对话用 crypto.randomUUID()，追问传相同值）",
  "mode": "ask",
  "prompt": "这篇文章主要讲了什么？",
  "selected_text": "（可选，mode=selected 时需要）"
}
```

**Mode 说明：**

| mode | 说明 |
|------|------|
| `summarize` | 总结文章，忽略 `prompt` |
| `ask` | 对文章提问 |
| `selected` | 划词提问，需传 `selected_text` |

**SSE 响应：**

```
data: {"session_id":"uuid","content":"思考中","status":true}
data: {"session_id":"uuid","content":"这是回答内容...","status":true}
data: {"session_id":"uuid","content":"","status":true,"message":"done"}
```

- `content` 为增量文本（非完整内容拼接）
- 完成帧 `message` = `"done"`
- 错误帧 `status` = `false`
