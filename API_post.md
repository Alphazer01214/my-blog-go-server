# 文章模块 API 文档

## 基础信息

- Base URL: `http://localhost:2333/api`
- 数据格式: `application/json`
- 通用响应格式同用户模块（`code: 0` 成功，`code: 7` 失败）

## 认证方式

受保护接口使用 Cookie 双令牌认证（`access-token` + `refresh-token`）。
前端需设置 `credentials: 'include'`。

## PostDetail 响应模型

```json
{
  "id": 1,
  "created_at": "2026-05-12T10:00:00Z",
  "updated_at": "2026-05-12T10:00:00Z",
  "title": "文章标题",
  "cover": "https://example.com/cover.jpg",
  "user_id": 1,
  "author": {
    "user_id": 1,
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z",
    "username": "testuser",
    "email": "user@example.com",
    "phone": "13800138000",
    "bio": "hello world",
    "avatar": "https://example.com/avatar.png",
    "admin": false,
    "role": 2,
    "post_count": 15,
    "comment_count": 42
  },
  "tags": ["go", "gin"],
  "category": "技术",
  "keywords": ["go", "后端"],
  "content": "文章正文内容",
  "view_count": 0,
  "comment_count": 0,
  "like_count": 0,
  "dislike_count": 0,
  "is_liked": false,
  "is_disliked": false,
  "public": true
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 文章 ID |
| `title` | string | 标题 |
| `cover` | string | 封面图 URL |
| `user_id` | uint | 作者 ID |
| `author` | UserInfo | 作者信息（完整 UserInfo 结构体） |
| `tags` | array | 标签数组 |
| `category` | string | 分类 |
| `keywords` | array | 关键词数组 |
| `content` | string | 正文（markdown） |
| `view_count` | int | 浏览数 |
| `comment_count` | int | 评论数 |
| `like_count` | int | 点赞数 |
| `dislike_count` | int | 点踩数 |
| `is_liked` | bool | 当前登录用户是否已点赞（未登录时为 `false`） |
| `is_disliked` | bool | 当前登录用户是否已点踩（未登录时为 `false`） |
| `public` | bool | 是否公开 |

---

## 1) 创建文章

`POST /api/create` — 需要 Cookie

**Request Body:**

```json
{
  "title": "我的第一篇文章",
  "cover": "https://example.com/cover.jpg",
  "category": "技术",
  "tags": ["go", "gin"],
  "keywords": ["go", "gin"],
  "content": "# 正文\nmarkdown 内容",
  "public": true,
  "env": {}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 是 | 标题 |
| `content` | string | 是 | 正文 |
| `category` | string | 否 | 分类 |
| `tags` | array | 否 | 标签数组 |
| `keywords` | array | 否 | 关键词数组 |
| `cover` | string | 否 | 封面 URL |
| `public` | bool | 否 | 是否公开（默认 `false`） |
| `env` | object | 否 | 环境信息 |

**Response:**

```json
{
  "code": 0,
  "data": { "id": 1, "title": "...", "author": { ... }, "..." : "..." },
  "msg": "post success"
}
```

---

## 2) 更新文章

`POST /api/update` 或 `POST /api/post/update` — 需要 Cookie

**Request Body:**

```json
{
  "post_id": 1,
  "title": "新标题",
  "cover": "https://example.com/new_cover.jpg",
  "category": "新分类",
  "tags": ["新标签"],
  "keywords": ["新关键词1", "新关键词2"],
  "content": "新正文",
  "public": true,
  "env": {}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `post_id` | uint | 是 | 文章 ID |
| `title` | string | 否 | 新标题 |
| `content` | string | 否 | 新正文 |
| `category` | string | 否 | 新分类 |
| `tags` | array | 否 | 新标签数组 |
| `keywords` | array | 否 | 新关键词数组 |
| `cover` | string | 否 | 新封面 |
| `public` | bool | 否 | 是否公开 |

> 权限：仅文章作者可更新。

**Response:**

```json
{
  "code": 0,
  "data": { "id": 1, "title": "新标题", "author": {...}, "..." : "..." },
  "msg": "update success"
}
```

---

## 3) 删除文章

`DELETE /api/post/:id` — 需要 Cookie

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 文章 ID |

> 权限：仅文章作者可删除（软删除）。

**Response:**

```json
{ "code": 0, "data": {}, "msg": "delete success" }
```

---

## 4) 获取单篇文章

`GET /api/post/:id` — 无需认证（未登录也可查看公开文章）

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 文章 ID |

> 隐私：私密文章只有作者本人可查看；未登录时 `is_liked`/`is_disliked` 均为 `false`。

**Response:** 见 PostDetail 模型。

---

## 5) 获取文章列表（分页）

`GET /api/post` — 无需认证

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码（默认 `1`） |
| `page_size` | int | 否 | 每页数量（默认 `20`，最大 `100`） |

> 隐私：非公开文章对非作者隐藏内容；列表按创建时间倒序。

**Response:**

```json
{
  "code": 0,
  "data": {
    "items": [ { "id": 1, "title": "...", "author": {...}, "..." : "..." } ],
    "page": 1,
    "page_size": 20,
    "total": 42
  },
  "msg": "query success"
}
```

---

## 6) 点赞文章（切换）

`POST /api/post/like` — 需要 Cookie

**Request Body:** `{ "post_id": 1 }`

> 行为：已点赞则取消点赞，已点踩则取消点踩并点赞，未操作则点赞。点赞与点踩互斥。

---

## 7) 点踩文章（切换）

`POST /api/post/dislike` — 需要 Cookie

**Request Body:** `{ "post_id": 1 }`

> 行为与点赞对称。
