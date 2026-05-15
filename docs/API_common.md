# API 通用说明

## 基础地址

```
http://<host>:<port>
```

## 统一响应格式

所有接口返回 JSON，结构如下：

```json
{
  "code": 0,       // 0=成功, 7=错误
  "data": {},      // 业务数据，具体结构见各接口
  "msg": "Success" // 提示信息
}
```

## 认证机制

采用 **双 Token 模式**（access token + refresh token），通过 HttpOnly Cookie 传递：

| Cookie 名 | 说明 | 有效期 |
|-----------|------|--------|
| `access-token` | 短期访问令牌 | 由服务端配置（默认较短） |
| `refresh-token` | 长期刷新令牌 | 由服务端配置（默认较长） |

**认证流程：**
1. 登录成功后，服务端 Set-Cookie 下发两个 token
2. 后续请求浏览器自动携带 Cookie
3. access token 过期时，中间件自动用 refresh token 签发新 access token，通过响应头 `access-token` 返回
4. 退出登录时，refresh token 加入黑名单

**接口标注：**
- `[公开]` — 无需认证
- `[认证]` — 需携带有效 Cookie

## 分页参数

需要分页的接口统一使用以下 Query 参数：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码，从 1 开始 |
| `page_size` | int | 20 | 每页条数，最大 100 |

分页响应格式：

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 100
}
```

## EnvInfo 环境信息

部分接口请求和响应中包含 `env` 字段，用于记录客户端环境：

```json
{
  "ipv4": "192.168.1.1",
  "ipv6": "",
  "os": "Windows 10",
  "device_info": "Chrome/120.0"
}
```

## RoleType 角色枚举

| 值 | 常量 | 说明 |
|----|------|------|
| 0 | Evil | 封禁用户 |
| 1 | Guest | 游客 |
| 2 | NormalUser | 普通用户 |
| 3 | VIP | VIP 用户 |
| 4 | Moderator | 版主 |
| 5 | TakamatsuTomori | 超级管理员 |

## 错误码

| 值 | 说明 |
|----|------|
| 0 | 成功 |
| 7 | 通用错误 |
