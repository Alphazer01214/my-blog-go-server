# 认证 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/auth/register` | POST | 注册 | 否 |
| `/api/auth/login` | POST | 登录 | 否 |
| `/api/auth/logout` | POST | 退出 | 是 |

---

## Token 机制

### 存储方式

Token 通过 **Set-Cookie** 下发。前端**无需**读写 `Authorization` header。

| Cookie 名 | 类型 | HttpOnly | Secure | 说明 |
|-----------|------|----------|--------|------|
| `access-token` | JWT | 是 | 条件 | 短期（15min），认证用 |
| `refresh-token` | JWT | 是 | 条件 | 长期（7天），刷新用 |

- **Secure**：仅当请求经过 TLS 或 `X-Forwarded-Proto: https` 时设为 `true`。
- **SameSite**：HTTP 下为 `Lax`，HTTPS 下为 `None`。

### 过期后的行为

`access-token` 过期后，服务端自动检测并触发刷新：

1. 任意请求返回 `{ code: 7, data: { reload: true }, msg: "token expired" }`
2. **前端应路由到登录页**，让用户在登录页面重新发起请求
3. 后端的 `refresh-token` 路由仍可用，会下发新 Token

> **为什么必须让前端跳转？**  
> 刷新流程在服务端 `JWTAuthMiddleware` 中间件中自动完成：若 `access-token` 过期且 `refresh-token` 有效，中间件会自动生成新的 `access-token`（覆盖 Cookie）并放行请求。若两者都过期，才需要前端处理。

---

## 端点详情

### POST /api/auth/register

注册新用户。

**Request Body**：

```json
{
  "username": "testuser",
  "password": "123456",
  "env": {
    "ipv4": "192.168.1.1",
    "ipv6": "",
    "os": "Windows 11",
    "device_info": "Chrome 120"
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | 是 | 3-32位字母数字下划线 |
| `password` | string | 是 | 6-128位 |
| `env` | object | 否 | 环境信息 |

**Response**（`code: 0`）：

```json
{
  "code": 0,
  "data": {
    "env": {
      "ipv4": "192.168.1.1",
      "ipv6": "",
      "os": "Windows 11",
      "device_info": "Chrome 120"
    },
    "username": "testuser"
  },
  "msg": "Success"
}
```

### POST /api/auth/login

登录并获取 Token。

**Request Body**：

```json
{
  "username": "testuser",
  "password": "123456",
  "env": {
    "ipv4": "192.168.1.1",
    "ipv6": "",
    "os": "Windows 11",
    "device_info": "Chrome 120"
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | 是 | 用户名 |
| `password` | string | 是 | 密码 |
| `env` | object | 否 | 环境信息 |

**Response**（`code: 0`）：

```json
{
  "code": 0,
  "data": {
    "env": {
      "ipv4": "192.168.1.1",
      "ipv6": "",
      "os": "Windows 11",
      "device_info": "Chrome 120"
    },
    "token": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "access_token_expire_time": 900,
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token_expire_time": 604800
    },
    "user_info": {
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
  },
  "msg": "Success"
}
```

Token 自动注入到 `Set-Cookie` 响应头。

**重复登录处理**：该用户之前的 refresh-token 会被写入黑名单（基于 Redis token:blacklist 前缀，自动过期）。

### POST /api/auth/logout

需要 `access-token` cookie。

退出后 Token 被加入黑名单，Cookie 被清除。

**Response**（`code: 0`）：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

---

## 前端实现说明

1. **登录流程**：
   - 调用 `POST /api/auth/login`（`credentials: "include"`）
   - 登录成功后 Cookie 自动保存在浏览器中
   - 后续所有请求都自动携带 Cookie

2. **请求处理**：
   - 所有请求设置 `credentials: "include"`（或 `withCredentials: true`）
   - 正常情况无需手动操作 Token

3. **403/Reload 处理**：
   - 当收到 `{ code: 7, data: { reload: true } }` 时，跳转到登录页面

4. **退出流程**：
   - 调用 `POST /api/auth/logout`（`credentials: "include"`）
   - 清除前端状态（Pinia store / localStorage）
