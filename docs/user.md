# 用户 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/me` | GET | 获取当前登录用户完整信息 | 是 |
| `/api/user/:id` | GET | 获取指定用户公开信息 | 否 |
| `/api/update_profile` | POST | 更新用户资料 | 是 |
| `/api/update_password` | POST | 修改密码 | 是 |
| `/api/settings` | GET | 获取用户隐私设置 | 是 |
| `/api/settings` | POST | 更新用户隐私设置 | 是 |
| `/api/user/follow` | POST | 关注/取消关注用户 | 是 |
| `/api/user/:id/followers` | GET | 获取粉丝列表 | 否 |
| `/api/user/:id/following` | GET | 获取关注列表 | 否 |
| `/api/user/:id/posts` | GET | 获取指定用户帖子列表 | 否 |
| `/api/user/:id/comments` | GET | 获取指定用户评论列表 | 否 |
| `/api/all_users` | GET | 获取所有用户列表 | 否 |

---

## 端点详情

### GET /api/me

获取当前登录用户的完整信息（包括邮箱、粉丝数、状态统计等）。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "phone": "13800138000",
    "bio": "个人简介",
    "avatar": "/uploads/avatars/xxx.png",
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
  "msg": "Success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `user_id` | uint | 用户 ID |
| `username` | string | 用户名 |
| `email` | string | 邮箱 |
| `phone` | string | 手机号 |
| `bio` | string | 个人简介 |
| `avatar` | string | 头像路径 |
| `admin` | bool | 是否管理员 |
| `role` | string | 角色（见枚举） |
| `banned` | bool | 是否被封禁 |
| `follower_count` | int | 粉丝数 |
| `following_count` | int | 关注数 |
| `post_count` | int | 帖子数 |
| `comment_count` | int | 评论数 |
| `received_like_count` | int | 收到的点赞数 |
| `received_dislike_count` | int | 收到的点踩数 |
| `is_followed` | bool | 当前用户是否关注了该用户（本人查询时始终为 `false`） |
| `created_at` | string | 注册时间 |
| `updated_at` | string | 更新时间 |

### GET /api/user/:id

获取指定用户的公开资料（仅公开字段）。不需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "phone": "",
    "bio": "个人简介",
    "avatar": "/uploads/avatars/xxx.png",
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
  "msg": "Success"
}
```

- `is_followed`：当前登录用户是否关注了该用户（未登录时始终为 `false`）。

### POST /api/update_profile

更新当前用户资料。需要登录。

**Request Body**（全部可选）：

```json
{
  "env": {
    "ipv4": "192.168.1.1",
    "ipv6": "",
    "os": "Windows 11",
    "device_info": "Chrome 120"
  },
  "new_username": "new_name",
  "new_email": "new@example.com",
  "new_phone": "13800138000",
  "new_bio": "新的个人简介",
  "new_avatar": "/uploads/avatars/new.png"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `env` | object | 环境信息（可选） |
| `new_username` | string | 新用户名 |
| `new_email` | string | 新邮箱 |
| `new_phone` | string | 新手机号 |
| `new_bio` | string | 新个人简介 |
| `new_avatar` | string | 新头像路径 |

**Response**：

```json
{
  "code": 0,
  "data": {
    "env": { ... },
    "user_info": { "user_id": 1, "username": "new_name", ... }
  },
  "msg": "Success"
}
```

### POST /api/update_password

修改密码。需要登录。

**Request Body**：

```json
{
  "env": {
    "ipv4": "192.168.1.1",
    "ipv6": "",
    "os": "Windows 11",
    "device_info": "Chrome 120"
  },
  "old_password": "123456",
  "new_password": "654321"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `old_password` | string | 是 | 旧密码 |
| `new_password` | string | 是 | 新密码 |

**Response**：

```json
{
  "code": 0,
  "data": { ... },
  "msg": "Success"
}
```

---

## 隐私设置

### GET /api/settings

获取当前用户的隐私设置。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "post_public": true,
    "comment_public": true,
    "follow_list_public": true,
    "follower_list_public": true
  },
  "msg": "Success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `post_public` | bool | 帖子是否公开 |
| `comment_public` | bool | 评论是否公开 |
| `follow_list_public` | bool | 关注列表是否公开 |
| `follower_list_public` | bool | 粉丝列表是否公开 |

### POST /api/settings

更新隐私设置。需要登录。

**Request Body**（全部可选）：

```json
{
  "post_public": false
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

---

## 关注功能

### POST /api/user/follow

关注或取消关注指定用户（开关切换）。需要登录。

**Request Body**：

```json
{
  "following_id": 2
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `following_id` | uint | 是 | 要关注的用户 ID |

**Response**：

```json
{
  "code": 0,
  "data": {
    "is_following": true,
    "is_mutual": false,
    "target_user": {
      "user_id": 2,
      "username": "otheruser",
      "email": "other@example.com",
      "phone": "",
      "bio": "",
      "avatar": "",
      "admin": false,
      "role": "normal_user",
      "banned": false,
      "follower_count": 5,
      "following_count": 3,
      "post_count": 1,
      "comment_count": 2,
      "received_like_count": 10,
      "received_dislike_count": 0,
      "is_followed": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  },
  "msg": "Success"
}
```

### GET /api/user/:id/followers

获取指定用户的粉丝列表。支持分页。不需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "user_id": 3,
        "username": "follower1",
        "email": "",
        "phone": "",
        "bio": "",
        "avatar": "",
        "admin": false,
        "role": "normal_user",
        "banned": false,
        "follower_count": 0,
        "following_count": 1,
        "post_count": 0,
        "comment_count": 0,
        "received_like_count": 0,
        "received_dislike_count": 0,
        "is_followed": false,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 5
  },
  "msg": "Success"
}
```

### GET /api/user/:id/following

获取指定用户的关注列表。响应格式同上。

---

## 用户内容

### GET /api/user/:id/posts

获取指定用户的帖子列表。支持分页。不需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

**Response**：格式同帖子列表接口。

### GET /api/user/:id/comments

获取指定用户的评论列表。支持分页。不需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

**Response**：格式同评论列表接口。

### GET /api/all_users

获取所有用户列表（简化信息）。

**Response**：

```json
{
  "code": 0,
  "data": {
    "users": [
      {
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
        "post_count": 0,
        "comment_count": 0,
        "received_like_count": 0,
        "received_dislike_count": 0,
        "is_followed": false,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "msg": "Success"
}
```
