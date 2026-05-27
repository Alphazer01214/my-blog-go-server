# 认证 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/auth/register` | POST | 注册 | 否 |
| `/api/auth/login` | POST | 登录 | 否 |
| `/api/auth/logout` | POST | 退出 | 是 |
| `/api/auth/refresh` | GET | 刷新 Token | 否（使用 refresh-token cookie） |

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
  "username": "string (3-32位字母数字下划线)",
  "password": "string (6-128位)"
}
```

**Response**（`code: 0`）：

```json
{
  "data": {
    "username": "testuser"
  }
}
```

### POST /api/auth/login

**Request Body**：

```json
{
  "username": "string",
  "password": "string",
  "env_info": {
    "ipv4": "string",
    "ipv6": "string",
    "os": "string",
    "device_info": "string"
  }
}
```

`env_info` 可选。

**Response**（`code: 0`）：

```json
{
  "data": {
    "username": "testuser",
    "user_id": 1,
    "role_type": "normal_user"
  }
}
```

Token 自动注入到 `Set-Cookie` 响应头。

**重复登录处理**：该用户之前的 refresh-token 会被写入黑名单（基于 DB token_blacklist 表）。

### POST /api/auth/logout

需要 `access-token` cookie。

退出后 Token 被加入黑名单，Cookie 被清除。

**Response**（`code: 0`）：

```json
{
  "data": null
}
```

### GET /api/auth/refresh

使用 `refresh-token` cookie 获取新的 Token 对。

**Response**（`code: 0`）：与登录相同的新 Token 注入到 Cookie。

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
   - 调用 `POST /api/auth/logout（credentials: "include"）`
   - 清除前端状态（Pinia store / localStorage）
