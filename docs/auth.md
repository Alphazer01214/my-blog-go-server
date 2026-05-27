# Auth API

| Method | Endpoint           | Description        |
| ------ | ------------------ | ------------------ |
| POST   | /api/auth/register | Register new user  |
| POST   | /api/auth/login    | Login              |
| POST   | /api/auth/logout   | Logout             |

---

## POST /api/auth/register

Register a new user account.

**Request Body:**

```json
{
  "username": "string",
  "password": "string",
  "env": {
    "ipv4": "192.168.1.1",
    "ipv6": "",
    "os": "Windows 11",
    "device_info": "Chrome 120"
  }
}
```

`env` 可选，为客户端环境信息。

**Response:**

```json
{
  "code": 0,
  "data": {
    "username": "newuser",
    "env": { "ipv4": "192.168.1.1", "os": "Windows 11" }
  },
  "msg": "register success"
}
```

---

## POST /api/auth/login

登录并获取 Token。Token 自动注入到 `Set-Cookie` 响应头。

**Request Body:**

```json
{
  "username": "string",
  "password": "string",
  "env": {
    "ipv4": "192.168.1.1",
    "ipv6": "",
    "os": "Windows 11",
    "device_info": "Chrome 120"
  }
}
```

`env` 可选。

**Response:**

```json
{
  "code": 0,
  "data": {
    "user_info": {
      "user_id": 1,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z",
      "username": "testuser",
      "email": "",
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
      "is_followed": false
    },
    "token": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "access_token_expire_time": 114514,
      "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token_expire_time": 114514
    },
    "env": { "ipv4": "192.168.1.1", "os": "Windows 11" }
  },
  "msg": "login success"
}
```

---

## POST /api/auth/logout

Log out the current user. Requires authentication.

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "logged out"
}
```
