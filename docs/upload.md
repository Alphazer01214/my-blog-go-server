# Upload API

| Method | Endpoint                   | Description             |
| ------ | -------------------------- | ----------------------- |
| POST   | /api/upload/init           | Init chunked upload     |
| POST   | /api/upload/chunk          | Upload chunk            |
| POST   | /api/upload/complete       | Complete upload         |
| GET    | /api/upload/:uploadId/status | Get upload status     |
| POST   | /api/upload/:uploadId/abort | Abort upload           |
| GET    | /api/file/:id              | Get file detail         |
| GET    | /api/file/:id/download     | Download file           |
| GET    | /api/files                 | List files              |
| GET    | /api/user/:id/files        | Get user's files        |
| DELETE | /api/file/:id              | Delete file             |

---

## POST /api/upload/init

Initialize a chunked file upload session. Requires authentication.

**Request Body:**

```json
{
  "filename": "string (required)",
  "file_size": 10485760,
  "mime_type": "application/pdf",
  "chunk_size": 1048576
}
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "upload_id": "uuid-string",
    "chunk_size": 1048576,
    "total_chunks": 10
  },
  "msg": "success"
}
```

---

## POST /api/upload/chunk

Upload a single chunk. Requires authentication.

**Request Body:** Multipart form data with:
- `upload_id` (string) - Upload session ID
- `chunk_index` (integer) - Chunk index (0-based)
- `chunk_hash` (string) - Hash of the chunk data
- Binary chunk data in the file field

**Response:**

```json
{
  "code": 0,
  "data": {
    "chunk_index": 0,
    "checksum": "sha256-hash"
  },
  "msg": "chunk uploaded"
}
```

---

## POST /api/upload/complete

Complete a chunked upload and create the file record. Requires authentication.

**Request Body:**

```json
{
  "upload_id": "string (required)",
  "title": "string (optional)",
  "public": true
}
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "user_id": 1,
    "name": "document.pdf",
    "path": "/uploads/2026/01/uuid.pdf",
    "size": 10485760,
    "mime_type": "application/pdf",
    "public": true,
    "duration": 0,
    "cover_url": "",
    "view_count": 0,
    "like_count": 0,
    "dislike_count": 0,
    "favorite_count": 0,
    "share_count": 0
  },
  "msg": "success"
}
```

---

## GET /api/upload/:uploadId/status

Get the status of an upload session. Requires authentication.

**Path Parameters:**
- `uploadId` (string) - Upload session ID

**Response:**

```json
{
  "code": 0,
  "data": {
    "upload_id": "uuid-string",
    "status": "uploading",
    "uploaded_count": 5,
    "total_chunks": 10,
    "uploaded_chunks": [0, 1, 2, 3, 4]
  },
  "msg": "success"
}
```

---

## POST /api/upload/:uploadId/abort

Abort an upload session. Requires authentication.

**Path Parameters:**
- `uploadId` (string) - Upload session ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "upload aborted"
}
```

---

## GET /api/file/:id

Get file detail by ID.

**Path Parameters:**
- `id` (integer) - File ID

**Response:**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "user_id": 1,
    "name": "document.pdf",
    "path": "/uploads/2026/01/uuid.pdf",
    "size": 10485760,
    "mime_type": "application/pdf",
    "public": true,
    "duration": 0,
    "cover_url": "",
    "view_count": 10,
    "like_count": 2,
    "dislike_count": 0,
    "favorite_count": 1,
    "share_count": 0
  },
  "msg": "success"
}
```

---

## GET /api/file/:id/download

Download a file. Returns the raw file binary.

**Path Parameters:**
- `id` (integer) - File ID

**Response:** Binary file stream with appropriate Content-Type header.

---

## GET /api/files

List all public files with pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...FileDetail..." }
  ],
  "msg": "success"
}
```

---

## GET /api/user/:id/files

Get files uploaded by a specific user.

**Path Parameters:**
- `id` (integer) - User ID

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...FileDetail..." }
  ],
  "msg": "success"
}
```

---

## DELETE /api/file/:id

Delete a file. Requires authentication (must be owner or admin).

**Path Parameters:**
- `id` (integer) - File ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "deleted"
}
```
