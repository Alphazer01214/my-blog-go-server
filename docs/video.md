# Video API

| Method | Endpoint                  | Description             |
| ------ | ------------------------- | ----------------------- |
| GET    | /api/video/:id/stream     | Stream video            |
| GET    | /api/video/:id/cover      | Get video cover image   |
| POST   | /api/video/init           | Init chunked upload     |
| POST   | /api/video/chunk          | Upload video chunk      |
| POST   | /api/video/merge          | Merge uploaded chunks   |
| POST   | /api/video/create         | Create video            |
| GET    | /api/video/:id            | Get video detail        |
| GET    | /api/video                | List videos             |
| GET    | /api/video/user/:userId   | Get user's videos       |
| PUT    | /api/video/:id            | Update video            |
| DELETE | /api/video/:id            | Delete video            |
| POST   | /api/video/:id/like       | Like video              |
| POST   | /api/video/:id/dislike    | Dislike video           |
| POST   | /api/video/:id/favorite   | Favorite video          |
| POST   | /api/video/:id/share      | Share video             |

---

## GET /api/video/:id/stream

Stream video binary data. Supports range requests.

**Path Parameters:**
- `id` (integer) - Video ID

**Response:** Binary video stream.

---

## GET /api/video/:id/cover

Get video cover image.

**Path Parameters:**
- `id` (integer) - Video ID

**Response:** Binary image data.

---

## POST /api/video/init

Initialize a chunked video upload session. Requires authentication.

**Request Body:**

```json
{
  "upload_id": "string (required)",
  "video_name": "string (required)",
  "video_size": 104857600,
  "chunk_size": 1048576,
  "title": "string (optional)",
  "description": "string (optional)",
  "category": "string (optional)",
  "tags": ["string"] (optional),
  "public": true,
  "forbid_comment": false,
  "forbid_share": false,
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" }
}
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "upload_id": "uuid-string",
    "chunk_size": 1048576,
    "total_chunks": 100
  },
  "msg": "success"
}
```

---

## POST /api/video/chunk

Upload a single video chunk. Requires authentication.

**Request Body:** Multipart form data with:
- `upload_id` (string) - Upload session ID
- `chunk_index` (integer) - Chunk index (0-based)
- `chunk_hash` (string) - Hash of the chunk data
- Binary chunk data in the file field

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "chunk uploaded"
}
```

---

## POST /api/video/merge

Merge all uploaded chunks into a final video. Requires authentication.

**Request Body:**

```json
{
  "upload_id": "string"
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "video merged"
}
```

---

## POST /api/video/create

Create a video record without chunked upload. Requires authentication.

**Request Body:**

```json
{
  "title": "string (required)",
  "description": "string (required)",
  "category": "string (optional)",
  "tags": ["string"] (optional),
  "public": true,
  "forbid_comment": false,
  "forbid_share": false,
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" }
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
    "title": "Trading Tutorial",
    "description": "Learn basics",
    "video_src_url": "https://example.com/video.mp4",
    "video_cover_url": "",
    "duration": 300,
    "size": 104857600,
    "mime_type": "video/mp4",
    "user_id": 1,
    "author": { "...UserInfo..." },
    "category": "tutorial",
    "tags": ["beginner"],
    "view_count": 0,
    "like_count": 0,
    "dislike_count": 0,
    "comment_count": 0,
    "favorite_count": 0,
    "share_count": 0,
    "public": true,
    "forbid_comment": false,
    "forbid_share": false,
    "status": "ready",
    "is_liked": false,
    "is_disliked": false,
    "is_favorited": false
  },
  "msg": "success"
}
```

---

## GET /api/video/:id

Get video detail by ID.

**Path Parameters:**
- `id` (integer) - Video ID

**Response:** Same as `POST /api/video/create` response data.

---

## GET /api/video

List videos with pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...VideoDetail..." }
  ],
  "msg": "success"
}
```

---

## GET /api/video/user/:userId

Get videos uploaded by a specific user.

**Path Parameters:**
- `userId` (integer) - User ID

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...VideoDetail..." }
  ],
  "msg": "success"
}
```

---

## PUT /api/video/:id

Update a video. Requires authentication (must be owner).

**Path Parameters:**
- `id` (integer) - Video ID

**Request Body:**

```json
{
  "title": "string (optional)",
  "description": "string (optional)",
  "video_cover_url": "string (optional)",
  "category": "string (optional)",
  "tags": ["string"] (optional),
  "public": true,
  "forbid_comment": false,
  "forbid_share": false
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "video updated"
}
```

---

## DELETE /api/video/:id

Delete a video. Requires authentication (must be owner or admin).

**Path Parameters:**
- `id` (integer) - Video ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "deleted"
}
```

---

## POST /api/video/:id/like

Like a video. Requires authentication.

**Path Parameters:**
- `id` (integer) - Video ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "liked"
}
```

---

## POST /api/video/:id/dislike

Dislike a video. Requires authentication.

**Path Parameters:**
- `id` (integer) - Video ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "disliked"
}
```

---

## POST /api/video/:id/favorite

Favorite a video. Requires authentication.

**Path Parameters:**
- `id` (integer) - Video ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "favorited"
}
```

---

## POST /api/video/:id/share

Share a video. Requires authentication.

**Path Parameters:**
- `id` (integer) - Video ID

**Request Body:**

```json
{
  "share_to": "string (optional)",
  "share_message": "string (optional)"
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "shared"
}
```
