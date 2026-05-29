# 视频 API

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/video/upload/init` | POST | 初始化视频上传 | 是 |
| `/api/video/upload/chunk` | POST | 上传视频分块（multipart） | 是 |
| `/api/video/upload/merge` | POST | 合并视频分块 | 是 |
| `/api/video` | POST | 创建视频记录 | 是 |
| `/api/video/:id` | GET | 获取视频详情 | 是 |
| `/api/video/list` | GET | 获取视频列表 | 是 |
| `/api/video/:id` | PUT | 更新视频信息 | 是 |
| `/api/video/:id` | DELETE | 删除视频 | 是 |
| `/api/video/like` | POST | 点赞/取消点赞 | 是 |
| `/api/video/dislike` | POST | 点踩/取消点踩 | 是 |
| `/api/video/favorite` | POST | 收藏/取消收藏 | 是 |
| `/api/video/share` | POST | 分享 | 是 |
| `/api/user/:id/videos` | GET | 获取指定用户视频列表 | 否 |
| `/api/video/:id/stream` | GET | 视频流播放 | 否 |
| `/api/video/:id/cover` | GET | 获取视频封面 | 否 |

---

## 视频上传流程

```
1. POST /api/video/upload/init   → 初始化上传（包含视频元数据）
2. POST /api/video/upload/chunk  → 逐块上传（循环）
3. POST /api/video/upload/merge  → 合并文件、生成封面、返回视频信息
```

> 合并完成后自动创建视频记录，无需再调用 `POST /api/video`。  
> 分块临时文件在合并后自动清理。  
> 视频合并使用流式写入，支持大文件上传。

---

## 端点详情

### POST /api/video/upload/init

初始化视频上传。需要登录。

**Request Body**：

```json
{
  "upload_id": "uuid-string",
  "video_name": "video.mp4",
  "video_size": 104857600,
  "chunk_size": 5242880,
  "env": {
    "ipv4": "192.168.1.1",
    "ipv6": "",
    "os": "Windows 11",
    "device_info": "Chrome 120"
  },
  "title": "视频标题",
  "description": "视频描述",
  "video_url": "",
  "video_cover_url": "",
  "category": "tech",
  "tags": ["Go", "Tutorial"],
  "public": true,
  "forbid_comment": false,
  "forbid_share": false
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `upload_id` | string | 是 | 前端生成的 UUID |
| `video_name` | string | 是 | 视频文件名 |
| `video_size` | int | 是 | 视频总大小（字节） |
| `chunk_size` | int | 是 | 分块大小（字节） |
| `env` | object | 否 | 环境信息 |
| `title` | string | 是 | 视频标题 |
| `description` | string | 是 | 视频描述 |
| `video_url` | string | 否 | 视频源 URL |
| `video_cover_url` | string | 否 | 封面 URL |
| `category` | string | 否 | 分类 |
| `tags` | array | 否 | 标签数组 |
| `public` | bool | 否 | 是否公开 |
| `forbid_comment` | bool | 否 | 是否禁止评论 |
| `forbid_share` | bool | 否 | 是否禁止分享 |

**Response**：

```json
{
  "code": 0,
  "data": {
    "sessions": [
      {
        "expire_at": "2024-01-02T00:00:00Z",
        "upload_id": "uuid-string",
        "user_id": 1,
        "video_name": "video.mp4",
        "video_size": 104857600,
        "chunk_size": 5242880,
        "total_chunks": 20,
        "status": 1,
        "uploaded_chunks": [false, false, false]
      }
    ]
  },
  "msg": "Success"
}
```

### POST /api/video/upload/chunk

上传一个视频分块。需要登录。

**Content-Type**：`multipart/form-data`

| 字段 | 类型 | 说明 |
|------|------|------|
| `upload_id` | string | 初始化返回的 ID |
| `chunk_index` | int | 块序号（从 0 开始） |
| `chunk_hash` | string | 块的哈希值（用于校验） |
| `file` | file | 视频块文件 |

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

### POST /api/video/upload/merge

合并视频分块，自动生成封面，创建视频记录。需要登录。

**Request Body**：

```json
{
  "upload_id": "uuid-string"
}
```

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "video upload completed"
}
```

> 合并后视频记录自动创建，`video_src_url` 和 `video_cover_url` 自动填充。

### POST /api/video

创建视频记录（仅用于外部 URL 视频，分块上传请使用上传流程）。需要登录。

**Request Body**：

```json
{
  "env": { ... },
  "title": "视频标题",
  "description": "视频描述",
  "video_url": "/uploads/videos/xxx.mp4",
  "video_cover_url": "/uploads/covers/xxx.png",
  "category": "tech",
  "tags": ["Go"],
  "public": true,
  "forbid_comment": false,
  "forbid_share": false
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 是 | 视频标题 |
| `description` | string | 是 | 视频描述 |
| `video_url` | string | 否 | 视频播放地址 |
| `video_cover_url` | string | 否 | 封面地址 |
| `category` | string | 否 | 分类 |
| `tags` | array | 否 | 标签 |
| `public` | bool | 否 | 是否公开 |
| `forbid_comment` | bool | 否 | 是否禁止评论 |
| `forbid_share` | bool | 否 | 是否禁止分享 |

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

### GET /api/video/:id

获取视频详情。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "title": "视频标题",
    "description": "视频描述",
    "video_src_url": "/uploads/videos/xxx.mp4",
    "video_cover_url": "/uploads/covers/xxx.png",
    "duration": 120,
    "size": 104857600,
    "mime_type": "video/mp4",
    "user_id": 1,
    "author": {
      "user_id": 1,
      "username": "作者",
      "avatar": ""
    },
    "category": "tech",
    "tags": ["Go"],
    "view_count": 100,
    "like_count": 10,
    "dislike_count": 1,
    "comment_count": 3,
    "favorite_count": 5,
    "share_count": 2,
    "public": true,
    "forbid_comment": false,
    "forbid_share": false,
    "status": "published",
    "is_liked": false,
    "is_disliked": false,
    "is_favorited": false
  },
  "msg": "Success"
}
```

### GET /api/video/list

获取视频列表。需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `category` | string | 否 | | 按分类筛选 |
| `keyword` | string | 否 | | 搜索关键词 |
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

**Response**：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "title": "视频标题",
        "description": "视频描述",
        "video_src_url": "",
        "video_cover_url": "",
        "duration": 120,
        "size": 0,
        "mime_type": "",
        "user_id": 1,
        "author": { ... },
        "category": "tech",
        "tags": ["Go"],
        "view_count": 100,
        "like_count": 10,
        "dislike_count": 1,
        "comment_count": 3,
        "favorite_count": 5,
        "share_count": 2,
        "public": true,
        "forbid_comment": false,
        "forbid_share": false,
        "status": "published",
        "is_liked": false,
        "is_disliked": false,
        "is_favorited": false
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 42
  },
  "msg": "Success"
}
```

### PUT /api/video/:id

更新视频信息。仅作者可操作。需要登录。

**Request Body**（全部可选）：

```json
{
  "video_id": 1,
  "title": "新标题",
  "description": "新描述",
  "video_cover_url": "新封面路径",
  "category": "new_category",
  "tags": ["newtag"],
  "public": false,
  "forbid_comment": false,
  "forbid_share": false
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

### DELETE /api/video/:id

删除视频。仅作者和管理员可操作。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

---

## 交互操作

### POST /api/video/like

点赞或取消点赞视频（toggle 切换）。需要登录。

**Request Body**：

```json
{
  "video_id": 1
}
```

### POST /api/video/dislike

点踩或取消点踩视频（toggle 切换）。需要登录。

**Request Body**：

```json
{
  "video_id": 1
}
```

### POST /api/video/favorite

收藏或取消收藏视频（toggle 切换）。需要登录。

**Request Body**：

```json
{
  "video_id": 1
}
```

### POST /api/video/share

分享视频（计数+1，非 toggle）。需要登录。

**Request Body**：

```json
{
  "video_id": 1,
  "share_to": "weibo",
  "share_message": "分享理由"
}
```

以上交互端点响应均为：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

---

## 用户视频

### GET /api/user/:id/videos

获取指定用户的视频列表。支持分页。不需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | |
| `page_size` | int | 否 | 20 | |

**Response**：格式同 `/api/video/list`。

---

## 流播放

### GET /api/video/:id/stream

视频流播放。返回视频文件的二进制流，支持 HTTP Range 请求（断点续传/拖拽进度条）。**不需要登录**（公开访问）。

**Response**：`video/mp4` 二进制流，支持 `Accept-Ranges: bytes`。

### GET /api/video/:id/cover

获取视频封面图片（由服务端在合并时自动从视频中截取）。**不需要登录**（公开访问）。

**Response**：`image/jpeg` 二进制流。
