# 用户 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/user` | GET | 获取当前登录用户完整信息 | 是 |
| `/api/user/:id` | GET | 获取指定用户公开信息 | 否 |
| `/api/user/settings` | GET | 获取用户隐私设置 | 是 |
| `/api/user/settings` | PUT | 更新用户隐私设置 | 是 |
| `/api/user/:id` | PUT | 更新用户资料 | 是 |
| `/api/user/follow/:target_id` | POST | 关注/取消关注用户 | 是 |
| `/api/user/follows` | GET | 获取关注列表 | 是 |
| `/api/user/followers` | GET | 获取粉丝列表 | 是 |
| `/api/user/avatar` | POST | 上传头像 | 是 |
| `/api/user/background` | POST | 上传背景图 | 是 |
| `/api/user/streak` | GET | 获取打卡日历数据 | 是 |
| `/api/user/streak` | POST | 签到打卡 | 是 |
| `/api/user/checkin` | GET | 获取签到状态 | 是 |

---

## 端点详情

### GET /api/user

获取当前登录用户的完整信息（包括邮箱、粉丝数、状态统计等）。

**Response**：

```json
{
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "avatar_path": "/uploads/avatars/xxx.png",
    "background_path": "/uploads/backgrounds/yyy.png",
    "signature": "个人签名",
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
}
```

### GET /api/user/:id

获取指定用户的公开资料（仅公开字段）。不需要登录。

**Response**：

```json
{
  "data": {
    "id": 1,
    "username": "testuser",
    "avatar_path": "...",
    "background_path": "...",
    "signature": "个人签名",
    "role_type": "normal_user",
    "status": 1,
    "gender": 0,
    "followers": 0,
    "followings": 0,
    "is_followed": false
  }
}
```

- `is_followed`：当前登录用户是否关注了该用户（未登录时始终为 `false`）。

### POST /api/user/:id

更新当前用户资料。需要登录。

**Request Body**：

```json
{
  "signature": "string (最多 200 字)",
  "gender": 0
}
```

- `gender`：0=保密，1=男，2=女

### GET /api/user/settings

获取当前用户的隐私设置。

**Response**：

```json
{
  "data": {
    "user_id": 1,
    "allow_strange_dm": true,
    "show_online_status": true
  }
}
```

### PUT /api/user/settings

更新隐私设置。

**Request Body**：

```json
{
  "allow_strange_dm": true,
  "show_online_status": false
}
```

---

## 关注功能

### POST /api/user/follow/:target_id

关注或取消关注指定用户（开关切换）。

**不需要 Request Body。**

**Response**：

```json
{
  "data": {
    "followed": true
  }
}
```

- `followed`：`true`=已关注，`false`=已取消

### GET /api/user/follows

获取当前用户的关注列表。支持分页。

**Response**：

```json
{
  "data": {
    "items": [
      {
        "id": 2,
        "username": "otheruser",
        "avatar_path": "...",
        "signature": "...",
        "role_type": "normal_user"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 5
  }
}
```

### GET /api/user/followers

获取当前用户的粉丝列表。响应格式同上。

---

## 上传

### POST /api/user/avatar

上传头像。`multipart/form-data`。字段 `file`（图片文件）。

### POST /api/user/background

上传背景图。同上。

响应格式同[文件上传文档](upload.md)。

---

## 签到 & 打卡

### GET /api/user/streak

获取打卡日历数据。

**Response**：

```json
{
  "data": {
    "streak": 0,
    "records": ["2025-05-01", "2025-05-02"]
  }
}
```

### POST /api/user/streak

签到打卡。无 Request Body。

**Response**：

```json
{
  "data": {
    "message": "checkin success"
  }
}
```

### GET /api/user/checkin

获取今日是否已签到。

**Response**：

```json
{
  "data": {
    "checkin": true
  }
}
```
