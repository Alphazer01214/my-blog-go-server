# 认证与用户 API

- [通用说明](./API_common.md)

---

## 1. 注册

**POST** `/api/auth/register` `[公开]`

### 请求体

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "username": "string (必填)",
  "password": "string (必填)"
}
```

### 响应 data

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "username": "注册成功的用户名"
}
```

---

## 2. 登录

**POST** `/api/auth/login` `[公开]`

### 请求体

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "username": "string (必填)",
  "password": "string (必填)"
}
```

### 响应 data

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "token": {
    "access_token": "jwt...",
    "access_token_expire_time": 3600,
    "refresh_token": "jwt...",
    "refresh_token_expire_time": 604800
  },
  "user_info": {
    "user_id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "username": "example",
    "email": "",
    "phone": "",
    "bio": "",
    "avatar": "",
    "admin": false,
    "role": 2,
    "banned": false,
    "follower_count": 0,
    "following_count": 0,
    "post_count": 0,
    "comment_count": 0,
    "received_like_count": 0,
    "received_dislike_count": 0,
    "is_followed": false
  }
}
```

> 登录成功会通过 Set-Cookie 下发 `access-token` 和 `refresh-token`

---

## 3. 退出登录

**POST** `/api/auth/logout` `[认证]`

### 请求体

无

### 响应 data

空对象

> 成功后会清除 Cookie 并将 refresh token 加入黑名单

---

## 4. 查询当前用户

**GET** `/api/me` `[认证]`

### 响应 data

```json
{
  "user_id": 1,
  "username": "example",
  "email": "user@example.com",
  "phone": "13800138000",
  "bio": "个人简介",
  "avatar": "https://example.com/avatar.png",
  "admin": false,
  "role": 2,
  "banned": false,
  "follower_count": 10,
  "following_count": 5,
  "post_count": 3,
  "comment_count": 15,
  "received_like_count": 42,
  "received_dislike_count": 1,
  "is_followed": false
}
```

---

## 5. 根据 ID 查询用户

**GET** `/api/user/:id` `[公开]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户 ID |

### 响应 data

同 [查询当前用户](#4-查询当前用户) 的 data 结构。

> `is_followed` 在登录状态下会反映当前用户是否关注了该用户

---

## 6. 获取所有用户

**GET** `/api/all_users` `[公开]`

### 响应 data

```json
[
  {
    "user_id": 1,
    "username": "user1",
    "...": "..."
  },
  {
    "user_id": 2,
    "username": "user2",
    "...": "..."
  }
]
```

> 返回 UserInfo 数组，登录状态下 `is_followed` 有值

---

## 7. 修改密码

**POST** `/api/update_password` `[认证]`

### 请求体

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "old_password": "string",
  "new_password": "string"
}
```

### 响应 data

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "user_info": {}
}
```

---

## 8. 修改个人资料

**POST** `/api/update_profile` `[认证]`

### 请求体

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "new_username": "string (可选)",
  "new_email": "string (可选)",
  "new_phone": "string (可选)",
  "new_bio": "string (可选)",
  "new_avatar": "string (可选, URL)"
}
```

> 仅传入需要修改的字段，不传的字段保持不变

### 响应 data

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "user_info": { "..." }
}
```

---

## 9. 关注 / 取消关注

**POST** `/api/user/follow` `[认证]`

> Toggle 模式：已关注则取消，未关注则创建

### 请求体

```json
{
  "following_id": 2
}
```

### 响应 data

```json
{
  "is_following": true,
  "is_mutual": false,
  "target_user": {
    "user_id": 2,
    "username": "target_user",
    "..." : "..."
  }
}
```

| 字段 | 说明 |
|------|------|
| `is_following` | 操作后的关注状态 |
| `is_mutual` | 是否互相关注 |
| `target_user` | 被操作的用户信息 |

---

## 10. 查询粉丝列表

**GET** `/api/user/:id/followers` `[公开]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户 ID |

### Query 参数

| 参数 | 类型 | 默认值 |
|------|------|--------|
| `page` | int | 1 |
| `page_size` | int | 20 |

### 响应 data

```json
{
  "items": [ { "user_id": 1, "username": "follower1", "...": "..." } ],
  "page": 1,
  "page_size": 20,
  "total": 100
}
```

---

## 11. 查询关注列表

**GET** `/api/user/:id/following` `[公开]`

### 参数与响应

同 [查询粉丝列表](#10-查询粉丝列表)，`items` 为该用户关注的人

---

## 12. 查询用户帖子

**GET** `/api/user/:id/posts` `[公开]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户 ID |

### Query 参数

| 参数 | 类型 | 默认值 |
|------|------|--------|
| `page` | int | 1 |
| `page_size` | int | 20 |

### 响应 data

`PostList` 结构，同 [查询帖子列表](./API_post.md#2-查询帖子列表)

---

## 13. 查询用户评论

**GET** `/api/user/:id/comments` `[公开]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户 ID |

### Query 参数

| 参数 | 类型 | 默认值 |
|------|------|--------|
| `page` | int | 1 |
| `page_size` | int | 20 |

### 响应 data

`CommentList` 结构，同 [查询帖子评论](./API_comment.md#2-查询帖子评论)

---

## 14. 获取用户设置

**GET** `/api/settings` `[认证]`

### 响应 data

```json
{
  "post_public": true
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `post_public` | bool | 发帖是否默认公开 |

---

## 15. 修改用户设置

**POST** `/api/settings` `[认证]`

### 请求体

```json
{
  "post_public": true
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `post_public` | bool | 是 | 发帖是否默认公开 |

### 响应 data

空对象
