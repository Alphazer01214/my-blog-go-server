# 文件上传下载 API

- [通用说明](./API_common.md)

---

## 概述

文件模块支持大文件分块上传、断点续传、多线程下载。

**上传流程：**
1. `POST /api/upload/init` — 创建上传会话，获取 `upload_id`
2. `PUT /api/upload/chunk` — 并行上传分块（支持断点续传）
3. `POST /api/upload/complete` — 完成上传，服务端合并分块

**配置参数（config.yaml）：**

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `server.upload_dir` | `./data/uploads` | 文件存储根目录 |
| `server.upload_chunk_size` | `10485760` | 默认分块大小（10MB） |

---

## 1. 创建上传会话

**POST** `/api/upload/init` `[认证]`

> 开始上传前调用，返回 `upload_id` 和分块信息。

### 请求体

```json
{
  "file_name": "my_video.mp4",
  "file_size": 1073741824,
  "mime_type": "video/mp4",
  "chunk_size": 10485760
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_name` | string | 是 | 原始文件名 |
| `file_size` | int64 | 是 | 文件总字节数 |
| `mime_type` | string | 是 | MIME 类型 |
| `chunk_size` | int | 否 | 分块大小（字节），默认 10MB |

### 响应 data

```json
{
  "upload_id": "uuid-xxx",
  "chunk_size": 10485760,
  "total_chunks": 103
}
```

---

## 2. 上传分块

**PUT** `/api/upload/chunk` `[认证]`

> 上传单个分块，可并行调用多个分块。支持断点续传（已上传的分块会跳过）。

### 请求头

| Header | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `X-Upload-Id` | string | 是 | 上传会话 ID |
| `X-Chunk-Index` | int | 是 | 分块索引（从 0 开始） |

### 请求体

分块二进制数据（raw body）

### 响应 data

```json
{
  "chunk_index": 0,
  "checksum": "d41d8cd98f00b204e9800998ecf8427e"
}
```

> `checksum` 为分块 MD5 值，客户端可用来校验完整性。

---

## 3. 完成上传

**POST** `/api/upload/complete` `[认证]`

> 所有分块上传完成后调用，服务端合并分块并创建视频记录。

### 请求体

```json
{
  "upload_id": "uuid-xxx",
  "title": "我的视频",
  "public": true
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `upload_id` | string | 是 | 上传会话 ID |
| `title` | string | 否 | 视频标题 |
| `public` | bool | 否 | 是否公开（默认 false） |

### 响应 data

`FileDetail` 结构，同 [获取文件详情](#5-获取文件详情)

---

## 4. 查询上传进度

**GET** `/api/upload/status?upload_id=xxx` `[公开]`

> 查询已上传的分块列表，用于断点续传时跳过已上传的分块。

### Query 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `upload_id` | string | 是 | 上传会话 ID |

### 响应 data

```json
{
  "upload_id": "uuid-xxx",
  "status": "uploading",
  "uploaded_count": 50,
  "total_chunks": 103,
  "uploaded_chunks": [0, 1, 2, 5, 6, 7, ...]
}
```

| 字段 | 说明 |
|------|------|
| `status` | `pending` / `uploading` / `completed` / `aborted` |
| `uploaded_chunks` | 已成功上传的分块索引列表 |

---

## 5. 取消上传

**POST** `/api/upload/abort?upload_id=xxx` `[认证]`

> 取消上传并清理临时分块文件。

### 响应 data

空对象

---

## 6. 获取文件详情

**GET** `/api/files/:id` `[公开]`

> 私有文件仅上传者可查看。

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 文件 ID |

### 响应 data

```json
{
  "id": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "user_id": 1,
  "name": "my_video.mp4",
  "path": "/uploads/files/uuid.mp4",
  "size": 1073741824,
  "mime_type": "video/mp4",
  "public": true,
  "duration": 120,
  "width": 1920,
  "height": 1080,
  "cover_url": "/uploads/covers/uuid.jpg",
  "view_count": 100,
  "like_count": 10,
  "dislike_count": 1,
  "favorite_count": 3,
  "share_count": 2
}
```

---

## 7. 下载文件

**GET** `/api/files/:id/download` `[公开]`

> 支持 HTTP Range 请求，可用于多线程下载或视频流式播放。
> 私有文件仅上传者可下载。

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 文件 ID |

### 请求头（可选，用于断点续传/多线程）

| Header | 说明 |
|--------|------|
| `Range` | `bytes=0-10485759` |

### 响应

- 完整下载：`200 OK` + 文件内容
- Range 请求：`206 Partial Content` + 部分文件内容
- `Accept-Ranges: bytes` 头表示支持 Range 请求

---

## 8. 文件列表

**GET** `/api/files` `[公开]`

> 未登录时仅返回公开文件，登录后返回所有文件。

### Query 参数

| 参数 | 类型 | 默认值 |
|------|------|--------|
| `page` | int | 1 |
| `page_size` | int | 20 |

### 响应 data

```json
{
  "items": [FileDetail, ...],
  "page": 1,
  "page_size": 20,
  "total": 50
}
```

---

## 9. 用户文件列表

**GET** `/api/user/:id/files` `[公开]`

> 未登录且非本人时仅返回公开文件。

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户 ID |

### Query 参数

| 参数 | 类型 | 默认值 |
|------|------|--------|
| `page` | int | 1 |
| `page_size` | int | 20 |

### 响应 data

同 [文件列表](#8-文件列表)

---

## 10. 删除文件

**DELETE** `/api/files/:id` `[认证]`

> 仅文件上传者可删除。

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 文件 ID |

### 响应 data

空对象

---

## 11. 文件点赞 / 取消点赞

**POST** `/api/files/like` `[认证]`

> Toggle 模式：已赞则取消，未赞则点赞（如有踩则先移除踩）

### 请求体

```json
{
  "file_id": 1
}
```

### 响应 data

空对象

---

## 断点续传示例

```
1. 客户端计算文件 MD5，检查本地缓存是否有 upload_id
2. GET /api/upload/status?upload_id=xxx → 获取已上传分块 [0,1,2,5]
3. 跳过 0,1,2,5，并行上传 3,4,6,7,... 分块
4. 所有分块上传完成后 POST /api/upload/complete
```

## 多线程下载示例

```
1. GET /api/files/:id → 获取文件总大小 size=100MB
2. 分成 4 个 Range 请求并行下载:
   Range: bytes=0-26214399        (0-25MB)
   Range: bytes=26214400-52428799  (25-50MB)
   Range: bytes=52428800-78643199  (50-75MB)
   Range: bytes=78643200-104857599 (75-100MB)
3. 所有分块下载完成后拼接为完整文件
```
