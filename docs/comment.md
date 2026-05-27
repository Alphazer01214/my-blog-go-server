# Comment API

| Method | Endpoint                   | Description            |
| ------ | -------------------------- | ---------------------- |
| POST   | /api/comment               | Create comment         |
| DELETE | /api/comment/:id           | Delete comment         |
| POST   | /api/comment/like          | Like comment           |
| POST   | /api/comment/dislike       | Dislike comment        |
| GET    | /api/post/:id/comments     | Get post comments      |
| GET    | /api/video/:id/comments    | Get video comments     |
| GET    | /api/user/:id/comments     | Get user's comments    |

---

## POST /api/comment

Create a new comment. Requires authentication.

**Request Body:**

```json
{
  "content": "string (required)",
  "target_type": "string (optional, e.g. 'post' or 'video')",
  "target_id": 1,
  "post_id": 1,
  "root_comment_id": 0,
  "parent_comment_id": 0,
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" }
}
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "comment_id": 1,
    "user_id": 1,
    "target_id": 10,
    "target_type": "post",
    "root_comment_id": 0,
    "parent_comment_id": 0,
    "content": "Great post!",
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "author": { "...UserInfo..." },
    "like_count": 0,
    "dislike_count": 0,
    "reply_count": 0,
    "is_liked": false,
    "is_disliked": false,
    "reply_comments": [],
    "env": { "ipv4": "192.168.1.1", "os": "Windows 11", "device_info": "Chrome 120" }
  },
  "msg": "success"
}
```

---

## DELETE /api/comment/:id

Delete a comment. Requires authentication (must be author or admin).

**Path Parameters:**
- `id` (integer) - Comment ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "deleted"
}
```

---

## POST /api/comment/like

Like a comment. Requires authentication.

**Request Body:**

```json
{
  "comment_id": 1
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "liked"
}
```

---

## POST /api/comment/dislike

Dislike a comment. Requires authentication.

**Request Body:**

```json
{
  "comment_id": 1
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "disliked"
}
```

---

## GET /api/post/:id/comments

Get comments for a specific post.

**Path Parameters:**
- `id` (integer) - Post ID

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    {
      "comment_id": 1,
      "user_id": 2,
      "target_id": 10,
      "target_type": "post",
      "root_comment_id": 0,
      "parent_comment_id": 0,
      "content": "Nice analysis!",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z",
      "author": { "...UserInfo..." },
      "like_count": 5,
      "dislike_count": 0,
      "reply_count": 2,
      "is_liked": false,
      "is_disliked": false,
      "reply_comments": [],
    "env": { "ipv4": "192.168.1.1", "os": "Windows 11", "device_info": "Chrome 120" }
    }
  ],
  "msg": "success"
}
```

---

## GET /api/video/:id/comments

Get comments for a specific video.

**Path Parameters:**
- `id` (integer) - Video ID

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:** Same structure as `GET /api/post/:id/comments`.

---

## GET /api/user/:id/comments

Get all comments made by a specific user.

**Path Parameters:**
- `id` (integer) - User ID

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:** Same structure as `GET /api/post/:id/comments`.
