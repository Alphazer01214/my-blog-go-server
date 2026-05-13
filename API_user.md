# 用户认证模块 API 文档

## 基础信息

- Base URL: `http://localhost:2333/api`
- 数据格式: `application/json`

## 通用响应格式

```json
{ "code": 0, "data": {}, "msg": "success" }
```

- `code: 0` = 成功, `code: 7` = 失败
- 所有接口均返回 HTTP 200，通过 code 区分

## 认证方式

受保护接口使用 Cookie 双令牌认证:

| Cookie 名 | 说明 |
|-----------|------|
| `access-token` | 访问令牌，过期后自动刷新 |
| `refresh-token` | 刷新令牌，用于签发新 access-token |

登录成功后服务端通过 `Set-Cookie` 写入上述两个 Cookie。
中间件会自动刷新过期的 access-token 并通过响应头 `access-token` / `access-expire-at` 返回新令牌。

## UserInfo 响应模型

```json
{
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
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `user_id` | uint | 用户 ID |
| `username` | string | 用户名 |
| `email` | string | 邮箱 |
| `phone` | string | 手机号 |
| `bio` | string | 个人简介 |
| `avatar` | string | 头像 URL |
| `admin` | bool | 是否管理员 |
| `role` | int | 角色类型 |
| `post_count` | int64 | 发帖数 |
| `comment_count` | int64 | 评论数 |

---

## 1) 用户注册

`POST /api/auth/register` — 无需认证

**Request Body:**

```json
{ "username": "testuser", "password": "testpass123", "env": {} }
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | 是 | 用户名（唯一） |
| `password` | string | 是 | 密码 |
| `env` | object | 否 | 环境信息 |

**Response:**

```json
{ "code": 0, "data": { "env": {}, "username": "testuser" }, "msg": "register success" }
```

---

## 2) 用户登录

`POST /api/auth/login` — 无需认证

**Request Body:**

```json
{ "username": "testuser", "password": "testpass123", "env": {} }
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "env": {},
    "token": {
      "access_token": "eyJhbGciOi...",
      "access_token_expire_time": 3600,
      "refresh_token": "eyJhbGciOi...",
      "refresh_token_expire_time": 604800
    },
    "user_info": {
      "user_id": 1, "username": "testuser",
      "created_at": "2026-05-12T10:00:00Z",
      "updated_at": "2026-05-12T10:00:00Z",
      "email": "", "phone": "",
      "bio": "", "avatar": "", "admin": false, "role": 2,
      "post_count": 0, "comment_count": 0
    }
  },
  "msg": "login success"
}
```

同时下发 Cookie：`Set-Cookie: access-token=...`、`Set-Cookie: refresh-token=...`

> 前端 fetch 需设置 `credentials: 'include'`。

---

## 3) 用户登出

`POST /api/auth/logout` — 需要 Cookie

**Response:** `{ "code": 0, "data": {}, "msg": "Logout successful" }`

---

## 4) 获取当前用户信息

`GET /api/me` — 需要 Cookie

**Response:** 见 UserInfo 模型。

---

## 5) 更新个人资料

`POST /api/update_profile` — 需要 Cookie

**Request Body:**

```json
{
  "new_username": "newname", "new_email": "new@example.com",
  "new_phone": "13900139000", "new_bio": "new bio",
  "new_avatar": "https://example.com/new_avatar.png",
  "env": {}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `new_username` | string | 否 | 新用户名 |
| `new_email` | string | 否 | 新邮箱 |
| `new_phone` | string | 否 | 新手机号 |
| `new_bio` | string | 否 | 新简介 |
| `new_avatar` | string | 否 | 新头像 URL |
| `env` | object | 否 | 环境信息 |

**Response:**

```json
{
  "code": 0,
  "data": { "env": {}, "user_info": { ...UserInfo... } },
  "msg": "UserUpdate profile successful"
}
```

---

## 6) 修改密码

`POST /api/update_password` — 需要 Cookie

**Request Body:**

```json
{ "old_password": "oldpass123", "new_password": "newpass456", "env": {} }
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `old_password` | string | 是 | 旧密码 |
| `new_password` | string | 是 | 新密码 |
| `env` | object | 否 | 环境信息 |

**Response:** `{ "code": 0, "data": { "env": {} }, "msg": "UserUpdate password successful" }`

---

## 7) 查询用户

`GET /api/user/:id` — 无需认证

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户 ID |

**Response:** 见 UserInfo 模型。

---

## 8) 获取所有用户

`GET /api/all_users` — 无需认证

**Response:**

```json
{
  "code": 0,
  "data": [ { ...UserInfo... } ],
  "msg": "Query all users successful"
}
```

---

## 9) 获取用户文章列表

`GET /api/user/:id/posts` — 无需认证

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | uint | 是 | 用户 ID |
| `page` | int | 否 | 页码（默认 `1`） |
| `page_size` | int | 否 | 每页数量（默认 `20`，最大 `100`） |

> 返回该用户发布的所有文章，按时间倒序，分页。

**Response:** 见 PostList（与 `GET /api/post` 结构相同）。

---

## 10) 获取用户评论列表

`GET /api/user/:id/comments` — 无需认证

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | uint | 是 | 用户 ID |
| `page` | int | 否 | 页码（默认 `1`） |
| `page_size` | int | 否 | 每页数量（默认 `20`，最大 `100`） |

> 返回该用户发布的所有评论，按时间倒序，分页（扁平列表，不含嵌套）。

**Response:**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "comment_id": 1, "user_id": 1, "post_id": 1, "content": "...",
        "author": { "user_id": 1, "username": "...", "..." : "..." },
        "likes": 0, "dislikes": 0,
        "is_liked": false, "is_disliked": false,
        "..." : "..."
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 42
  },
  "msg": "query user comments success"
}
```

> 每条评论包含 `author` 字段（统一 UserInfo 结构体）和 `is_liked`/`is_disliked` 字段（当前登录用户的点赞/点踩状态，未登录时为 `false`）。
