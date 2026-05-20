# 论坛管理 API（TakamatsuTomori 超级管理员）

- [通用说明](./API_common.md)

---

## 概述

本模块所有端点要求调用者角色为 `TakamatsuTomori`（`role = 5`）。当前实现中角色鉴权在各 Handler 方法内进行：先通过 JWT 中间件验证登录态，再检查 `claims.RoleType`。

---

## 1. 论坛统计

**GET** `/api/tomori/stats` `[认证 + 管理员]`

### 响应 data

```json
{
  "user_count": 100,
  "post_count": 500,
  "comment_count": 3000,
  "file_count": 50,
  "agent_count": 20,
  "chat_count": 200
}
```

---

## 2. 用户管理

### 2.1 用户列表

**GET** `/api/tomori/users` `[认证 + 管理员]`

Query 参数：`page` / `page_size`（见通用分页规范）

### 响应 data

```json
{
  "items": [UserInfo, ...],
  "page": 1,
  "page_size": 20,
  "total": 100
}
```

> `items` 中每个元素为 `UserInfo` 结构（含 `role`、`banned` 字段），同[通用用户信息](./API_user.md#5-根据-id-查询用户)

### 2.2 封禁/解封

**POST** `/api/tomori/user/ban` `[认证 + 管理员]`

```json
{ "user_id": 1, "banned": true }
```

### 2.3 修改角色

**POST** `/api/tomori/user/role` `[认证 + 管理员]`

```json
{ "user_id": 1, "role": 2 }
```

| `role` 值 | 含义 |
|-----------|------|
| 0 | Evil（封禁用户） |
| 1 | Guest（游客） |
| 2 | NormalUser（普通用户） |
| 3 | VIP |
| 4 | Moderator（版主） |
| 5 | TakamatsuTomori（超级管理员） |

### 2.4 重置密码

**POST** `/api/tomori/user/reset_password` `[认证 + 管理员]`

```json
{ "user_id": 1, "new_password": "newpwd123" }
```

### 2.5 删除用户

**DELETE** `/api/tomori/user/:id` `[认证 + 管理员]`

> 级联删除 profile、setting、关注关系。

---

## 3. 帖子管理

### 3.1 帖子列表（含私密）

**GET** `/api/tomori/posts` `[认证 + 管理员]`

Query 参数：`page` / `page_size`

### 响应 data

`PostList` 结构，同[帖子列表](./API_post.md#2-查询帖子列表)，但包含所有私密帖子且 `content` 不被替换。

### 3.2 删除帖子

**DELETE** `/api/tomori/post/:id` `[认证 + 管理员]`

> 跳过作者校验，可删除任意帖子。

---

## 4. 评论管理

### 4.1 评论列表

**GET** `/api/tomori/comments` `[认证 + 管理员]`

Query 参数：`page` / `page_size`

### 响应 data

平铺评论列表（无递归树），每条为 `Comment` 结构。

### 4.2 删除评论

**DELETE** `/api/tomori/comment/:id` `[认证 + 管理员]`

> 跳过作者校验，级联删除该评论的所有子回复，同步更新 `comment_count` / `reply_count`。

---

## 5. 文件管理

### 5.1 文件列表（含私密）

**GET** `/api/tomori/files` `[认证 + 管理员]`

Query 参数：`page` / `page_size`

### 响应 data

`FileList` 结构，同[文件列表](./API_file.md#8-文件列表)，但包含所有私密文件。

### 5.2 删除文件

**DELETE** `/api/tomori/file/:id` `[认证 + 管理员]`

> 跳过作者校验，同时删除磁盘文件。

---

## 6. AI 智能体管理

### 6.1 智能体列表

**GET** `/api/tomori/agents` `[认证 + 管理员]`

### 响应 data

```json
{
  "items": [
    {
      "agent_id": 1,
      "user_id": 2,
      "username": "alice",
      "name": "my-assistant",
      "model_name": "gpt-4",
      "provider": "openai",
      "activate": true,
      "created_at": "2024-01-01 12:00:00"
    }
  ],
  "total": 1
}
```

> 相比普通 AgentInfo，管理员视图额外包含 `user_id` 和 `username`。

### 6.2 删除智能体

**DELETE** `/api/tomori/agent/:id` `[认证 + 管理员]`

> 跳过所有者校验。

---

## 7. 聊天会话管理

### 7.1 聊天会话列表

**GET** `/api/tomori/chats` `[认证 + 管理员]`

Query 参数：`page` / `page_size`

### 响应 data

同 [查询聊天会话列表](./API_ai.md#7-查询聊天会话列表)

### 7.2 删除聊天会话

**DELETE** `/api/tomori/chat/:chat_id` `[认证 + 管理员]`

> `chat_id` 为会话 UUID。

---

## 8. 系统维护

### 8.1 Token 黑名单列表

**GET** `/api/tomori/blacklist` `[认证 + 管理员]`

Query 参数：`page` / `page_size`

### 响应 data

```json
{
  "items": [
    {
      "id": "1",
      "token": "eyJhbGciOiJIUzI1Ni...",
      "created_at": "2024-01-01 12:00:00"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 10
}
```

> `token` 字段截断至前 20 字符，防止完整 token 泄露。

### 8.2 清空黑名单

**POST** `/api/tomori/blacklist/clear` `[认证 + 管理员]`

> 清空所有 Token 黑名单记录。

---

## 前端接入注意

1. 管理员身份由 `/api/me` 返回的 `role == 5` 判断
2. 所有 `/api/tomori/*` 请求需携带登录 Cookie（`credentials: "include"`）
3. 角色不足时返回 `code: 7, msg: "tomori access only"`
4. 管理后台页面建议独立路由 `/admin/*` 隔离
