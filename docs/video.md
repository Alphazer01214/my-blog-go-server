# 视频 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/video/init` | POST | 初始化视频上传 | 是 |
| `/api/video/chunk` | POST | 上传视频分块（multipart） | 是 |
| `/api/video/complete` | POST | 完成上传并合并 | 是 |
| `/api/video/:id` | GET | 获取视频详情 | 否 |
| `/api/video/:id` | DELETE | 删除视频 | 是 |
| `/api/video/:id` | PUT | 更新视频信息 | 是 |
| `/api/video/list` | GET | 获取视频列表 | 否 |
| `/api/user/:id/videos` | GET | 获取指定用户视频列表 | 否 |
| `/api/video/:id/interact` | POST | 点赞/收藏/分享操作 | 是 |
| `/api/video/:id/stream` | GET | 视频流播放（HLS） | 否 |
| `/api/video/:id/cover` | GET | 获取视频封面 | 否 |

---

## 端点详情

### POST /api/video/init

初始化视频上传，返回 upload_id。

**Request Body**：

```json
{
  "title": "视频标题",
  "description": "视频描述",
  "cover_path": "封面路径（可选，先传空字符串）"
}
```

**Response**：

```json
{
  "data": {
    "upload_id": "uuid"
  }
}
```

### POST /api/video/chunk

上传一个视频分块。

**Content-Type**：`multipart/form-data`

| 字段 | 类型 | 说明 |
|------|------|------|
| `upload_id` | string | 初始化返回的 ID |
| `chunk` | file | 视频块文件 |
| `index` | int | 块序号（从 0 开始） |
| `total` | int | 总块数 |

**Response**：`204 No Content`

### POST /api/video/complete

通知服务端所有块已上传，开始转码处理。

**Request Body**：

```json
{
  "upload_id": "uuid"
}
```

**Response**：

```json
{
  "data": {
    "id": 1,
    "upload_status": "processing"
  }
}
```

- 上传状态将变为 `processing`，服务端会异步转码

### GET /api/video/:id

获取视频详情。

**Response**：

```json
{
  "data": {
    "id": 1,
    "title": "标题",
    "description": "描述",
    "cover_path": "/uploads/covers/xxx.png",
    "play_url": "/api/video/1/stream",
    "video_path": "/uploads/videos/xxx.mp4",
    "duration": 120.5,
    "status": "completed",
    "like_count": 10,
    "favorite_count": 5,
    "share_count": 1,
    "comment_count": 3,
    "view_count": 100,
    "is_liked": false,
    "is_favorited": false,
    "is_owner": false,
    "upload_status": "completed",
    "user_info": {
      "id": 1,
      "username": "作者",
      "avatar_path": ""
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### DELETE /api/video/:id

删除视频。仅作者和管理员可操作。

### PUT /api/video/:id

更新视频信息。仅作者可操作。

**Request Body**（全部可选）：

```json
{
  "title": "新标题",
  "description": "新描述",
  "cover_path": "新封面路径"
}
```

### GET /api/video/list

获取视频列表。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

### GET /api/user/:id/videos

获取指定用户的视频列表。支持分页。

**Response**：同 `/api/video/list`。

### POST /api/video/:id/interact

交互操作（同帖子）。

**Request Body**：

```json
{
  "action_type": "like"
}
```

| `action_type` | 说明 |
|---------------|------|
| `like` | 点赞（toggle） |
| `favorite` | 收藏（toggle） |
| `share` | 分享（计数+1） |

### GET /api/video/:id/stream

视频流播放。返回视频文件的二进制流。

**没有 JWT 中间件保护**（公开访问）。

**Response**：`video/mp4` 或 `application/x-mpegURL` 等。

### GET /api/video/:id/cover

获取视频封面图片。

**没有 JWT 中间件保护**（公开访问）。

**Response**：图片二进制流。
