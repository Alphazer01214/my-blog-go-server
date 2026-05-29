# 文件上传 API

通用文件分块上传系统。支持大文件分片、断点续传。

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/upload/init` | POST | 初始化上传任务 | 是 |
| `/api/upload/chunk` | PUT | 上传文件块（二进制） | 是 |
| `/api/upload/complete` | POST | 合并文件块 | 是 |
| `/api/upload/status` | GET | 获取上传状态 | 否 |
| `/api/upload/abort` | POST | 取消上传 | 是 |
| `/api/files/:id` | GET | 获取文件信息 | 否 |
| `/api/files/:id/download` | GET | 下载文件 | 否 |
| `/api/files` | GET | 获取文件列表 | 否 |
| `/api/user/:id/files` | GET | 获取指定用户文件列表 | 否 |
| `/api/files/:id` | DELETE | 删除文件 | 是 |

---

## 上传流程

```
1. POST /api/upload/init       → 初始化，获取 upload_id
2. PUT  /api/upload/chunk      → 逐块上传二进制数据（循环）
3. POST /api/upload/complete   → 通知服务端合并文件
```

**可选**：发送完部分块后，GET `/api/upload/status` 查询已收到的块，实现断点续传。

---

## 端点详情

### POST /api/upload/init

初始化上传任务。需要登录。

**Request Body**：

```json
{
  "file_name": "image.png",
  "file_size": 15728640,
  "mime_type": "image/png",
  "chunk_size": 5242880
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_name` | string | 是 | 文件名 |
| `file_size` | int | 是 | 文件总大小（字节） |
| `mime_type` | string | 是 | MIME 类型 |
| `chunk_size` | int | 否 | 分块大小（字节），默认 5MB |

**Response**：

```json
{
  "code": 0,
  "data": {
    "upload_id": "uuid-string",
    "chunk_size": 5242880,
    "total_chunks": 3
  },
  "msg": "Success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `upload_id` | string | 上传任务 ID |
| `chunk_size` | int | 实际分块大小 |
| `total_chunks` | int | 总块数 |

### PUT /api/upload/chunk

上传一个文件块。需要登录。

**Headers**：

| Header | 说明 |
|--------|------|
| `X-Upload-Id` | 初始化返回的 upload_id |
| `X-Chunk-Index` | 当前块序号（从 0 开始，字符串） |

**Body**：文件块的**原始二进制数据**（非 multipart）。

**Response**：

```json
{
  "code": 0,
  "data": {
    "chunk_index": 0,
    "checksum": "md5-of-chunk"
  },
  "msg": "Success"
}
```

### POST /api/upload/complete

通知服务端所有块已上传完成，开始合并。需要登录。

**Request Body**：

```json
{
  "upload_id": "uuid-string",
  "title": "文件显示名称",
  "public": true
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `upload_id` | string | 是 | 上传任务 ID |
| `title` | string | 否 | 文件显示名称 |
| `public` | bool | 否 | 是否公开（默认 false） |

**Response**：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "user_id": 1,
    "name": "image.png",
    "path": "/uploads/2024/image.png",
    "size": 15728640,
    "mime_type": "image/png",
    "public": true,
    "duration": 0,
    "width": 0,
    "height": 0,
    "cover_url": "",
    "view_count": 0,
    "like_count": 0,
    "dislike_count": 0,
    "favorite_count": 0,
    "share_count": 0
  },
  "msg": "Success"
}
```

### GET /api/upload/status

获取上传状态（用于断点续传）。不需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `upload_id` | string | 是 | 初始化返回的 upload_id |

**Response**：

```json
{
  "code": 0,
  "data": {
    "upload_id": "uuid-string",
    "status": "uploading",
    "uploaded_count": 2,
    "total_chunks": 3,
    "uploaded_chunks": [0, 1]
  },
  "msg": "Success"
}
```

### POST /api/upload/abort

取消上传并清理临时文件。需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `upload_id` | string | 是 | 上传任务 ID |

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

---

## 文件管理

### GET /api/files/:id

获取指定文件记录。不需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "user_id": 1,
    "name": "image.png",
    "path": "/uploads/2024/image.png",
    "size": 15728640,
    "mime_type": "image/png",
    "public": true,
    "duration": 0,
    "width": 1920,
    "height": 1080,
    "cover_url": "",
    "view_count": 10,
    "like_count": 2,
    "dislike_count": 0,
    "favorite_count": 1,
    "share_count": 0
  },
  "msg": "Success"
}
```

### GET /api/files/:id/download

下载文件。返回文件二进制流。不需要登录。

**Response**：文件二进制数据，Content-Type 根据文件类型自动设置。

### GET /api/files

获取文件列表。支持分页。不需要登录。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
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
        "user_id": 1,
        "name": "image.png",
        "path": "/uploads/2024/image.png",
        "size": 15728640,
        "mime_type": "image/png",
        "public": true,
        "duration": 0,
        "width": 0,
        "height": 0,
        "cover_url": "",
        "view_count": 0,
        "like_count": 0,
        "dislike_count": 0,
        "favorite_count": 0,
        "share_count": 0
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 5
  },
  "msg": "Success"
}
```

### GET /api/user/:id/files

获取指定用户的文件列表。不需要登录。响应格式同上。

### DELETE /api/files/:id

删除文件。需要登录。

**Response**：

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```
