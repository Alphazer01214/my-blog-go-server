# 管理后台 API

> 仅 **超级管理员**（`takamatsu_tomori`）角色可访问。

## 目录

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/admin/users` | GET | 获取所有用户列表 |
| `/api/admin/user/:id` | PUT | 更新用户角色/状态 |
| `/api/admin/post/:id` | DELETE | 删除任意帖子 |
| `/api/admin/posts` | GET | 获取所有帖子列表 |

---

## 端点详情

### GET /api/admin/users

获取所有用户列表。支持分页。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

**Response**：

```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "username": "testuser",
        "email": "test@example.com",
        "avatar_path": "",
        "background_path": "",
        "signature": "签名",
        "role_type": "normal_user",
        "status": 1,
        "gender": 0,
        "grade": 0,
        "followers": 0,
        "followings": 0,
        "post_count": 0,
        "comment_count": 0,
        "like_count": 0,
        "register_ip": "127.0.0.1",
        "created_at": "2024-01-01T00:00:00Z",
        "last_login_at": "2024-01-01T00:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 42
  }
}
```

### PUT /api/admin/user/:id

更新指定用户的角色或状态。

**Request Body**：

```json
{
  "role_type": "vip",
  "status": 1
}
```

- `status`：1=正常，0=停用

**Response**：标准成功响应。

### DELETE /api/admin/post/:id

删除指定帖子。超级管理员可以删除任意帖子。

**Response**：标准成功响应。

### GET /api/admin/posts

获取所有帖子列表。支持分页。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

**Response**：格式同帖子列表接口。
