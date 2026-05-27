# 文件上传 API

通用文件分块上传系统。支持大文件分片、断点续传。

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/upload` | POST | 上传完成初始化 | 是 |
| `/api/upload/chunk` | PUT | 上传文件块（二进制） | 是 |
| `/api/upload/complete` | POST | 合并文件块 | 是 |
| `/api/upload/uploaded` | GET | 获取已上传块序号 | 是 |
| `/api/upload/cancel` | POST | 取消上传 | 是 |
| `/api/files/:id` | GET | 获取已上传文件信息 | 是 |
| `/api/files` | GET | 获取文件列表 | 是 |

---

## 上传流程

```
1. POST /api/upload         → 初始化，获取 upload_id
2. PUT  /api/upload/chunk   → 逐块上传二进制数据（循环）
3. POST /api/upload/complete → 通知服务端合并文件
```

**可选**：发送完部分块后，GET `/api/upload/uploaded` 查询已收到的块，实现断点续传。

---

## 端点详情

### POST /api/upload

初始化上传任务。

**Request Body**：

```json
{
  "filename": "image.png",
  "md5": "文件 MD5（可选）",
  "target": "post"
}
```

- `target`：上传用途标识，如 `post`、`avatar`、`background` 等

**Response**：

```json
{
  "data": {
    "upload_id": "uuid",
    "chunk_size": 5242880,
    "chunk_total": 3,
    "upload_status": "pending",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

- `chunk_size`：每块大小（默认 5MB = 5242880 字节）
- `chunk_total`：总块数

### PUT /api/upload/chunk

上传一个文件块。

**Headers**：

| Header | 说明 |
|--------|------|
| `X-Upload-Id` | 初始化返回的 upload_id（字符串） |
| `X-Chunk-Index` | 当前块序号（从 0 开始，字符串） |

**Body**：文件块的**原始二进制数据**（非 multipart）。

**Response**：`204 No Content`（成功无 body）。

### POST /api/upload/complete

通知服务端所有块已上传完成，开始合并。

**Request Body**：

```json
{
  "upload_id": "uuid",
  "md5": "（可选）服务端会校验 MD5",
  "filename": "最终文件名（可选）"
}
```

**Response**：

```json
{
  "data": {
    "upload_id": "uuid",
    "file_path": "/uploads/2024/image.png",
    "file_size": 15728640,
    "upload_status": "completed"
  }
}
```

### GET /api/upload/uploaded

获取已上传的块序号列表（用于断点续传）。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `upload_id` | string | 是 | 初始化返回的 upload_id |

**Response**：

```json
{
  "data": {
    "upload_id": "uuid",
    "uploaded_chunks": [0, 1]
  }
}
```

### POST /api/upload/cancel

取消上传并清理临时文件。

**Request Body**：

```json
{
  "upload_id": "uuid"
}
```

### GET /api/files/:id

获取指定文件记录。

**Response**：

```json
{
  "data": {
    "id": 1,
    "upload_id": "uuid",
    "original_name": "image.png",
    "file_path": "/uploads/2024/image.png",
    "file_size": 15728640,
    "mime_type": "image/png",
    "upload_status": "completed",
    "md5": "文件的 md5",
    "target": "post",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### GET /api/files

获取当前用户的所有文件列表。支持分页。

**Response**：

```json
{
  "data": {
    "items": [ /* 同 GET /api/files/:id 格式 */ ],
    "page": 1,
    "page_size": 20,
    "total": 5
  }
}
```
