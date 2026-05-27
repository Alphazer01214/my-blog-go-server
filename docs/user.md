# User API

| Method | Endpoint              | Description              |
| ------ | --------------------- | ------------------------ |
| GET    | /api/me               | Get current user info    |
| GET    | /api/user/:id         | Get user by ID           |
| GET    | /api/user/:id/posts   | Get user's posts         |
| GET    | /api/user/:id/comments| Get user's comments      |
| GET    | /api/user/:id/followers| Get user's followers    |
| GET    | /api/user/:id/following| Get user's following    |
| GET    | /api/all_users        | List all users           |
| POST   | /api/update_password  | Update password          |
| POST   | /api/update_profile   | Update profile           |
| POST   | /api/user/follow      | Follow/unfollow user     |
| GET    | /api/settings         | Get user settings        |
| POST   | /api/settings         | Update user settings     |

---

## GET /api/me

Get the current authenticated user's information.

**Response:**

```json
{
  "code": 0,
  "data": {
    "user_id": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "username": "testuser",
    "email": "test@example.com",
    "phone": "",
    "bio": "Hello world",
    "avatar": "https://example.com/avatar.png",
    "admin": false,
    "role": "user",
    "banned": false,
    "follower_count": 10,
    "following_count": 5,
    "post_count": 3,
    "comment_count": 12,
    "received_like_count": 50,
    "received_dislike_count": 2,
    "is_followed": false
  },
  "msg": "success"
}
```

---

## GET /api/user/:id

Get a specific user's public information.

**Path Parameters:**
- `id` (integer) - User ID

**Response:** Same structure as `GET /api/me`.

---

## GET /api/user/:id/posts

Get posts created by a specific user.

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
    {
      "id": 1,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z",
      "title": "My Post",
      "cover": "",
      "user_id": 1,
      "author": { "...UserInfo..." },
      "tags": ["tag1"],
      "category": "general",
      "keywords": [],
      "content": "Post content here",
      "view_count": 100,
      "comment_count": 5,
      "like_count": 10,
      "dislike_count": 0,
      "favorite_count": 3,
      "share_count": 1,
      "public": true,
      "forbid_comment": false,
      "forbid_share": false,
      "is_liked": false,
      "is_disliked": false,
      "is_favorited": false,
      "env": { "ipv4": "192.168.1.1", "os": "Windows 11", "device_info": "Chrome 120" }
    }
  ],
  "msg": "success"
}
```

---

## GET /api/user/:id/comments

Get comments made by a specific user.

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
    {
      "comment_id": 1,
      "user_id": 1,
      "target_id": 10,
      "target_type": "post",
      "root_comment_id": 0,
      "parent_comment_id": 0,
      "content": "Nice post!",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z",
      "author": { "...UserInfo..." },
      "like_count": 2,
      "dislike_count": 0,
      "reply_count": 1,
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

## GET /api/user/:id/followers

Get followers of a specific user.

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
    {
      "user_id": 2,
      "username": "follower1",
      "avatar": "",
      "bio": "",
      "is_followed": true
    }
  ],
  "msg": "success"
}
```

---

## GET /api/user/:id/following

Get users that a specific user is following.

**Path Parameters:**
- `id` (integer) - User ID

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:** Same structure as `GET /api/user/:id/followers`.

---

## GET /api/all_users

Get a list of all users. Requires authentication.

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...UserInfo..." }
  ],
  "msg": "success"
}
```

---

## POST /api/update_password

Update the current user's password. Requires authentication.

**Request Body:**

```json
{
  "old_password": "string",
  "new_password": "string",
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" }
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "password updated"
}
```

---

## POST /api/update_profile

Update the current user's profile. Requires authentication.

**Request Body:**

```json
{
  "new_username": "string (optional)",
  "new_email": "string (optional)",
  "new_phone": "string (optional)",
  "new_bio": "string (optional)",
  "new_avatar": "string (optional)",
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" }
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "profile updated"
}
```

---

## POST /api/user/follow

Follow or unfollow a user. Requires authentication.

**Request Body:**

```json
{
  "following_id": 2
}
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "is_following": true,
    "is_mutual": false,
    "target_user": { "...UserInfo..." }
  },
  "msg": "success"
}
```

---

## GET /api/settings

Get current user's privacy settings. Requires authentication.

**Response:**

```json
{
  "code": 0,
  "data": {
    "post_public": true,
    "comment_public": true,
    "follow_list_public": true,
    "follower_list_public": true
  },
  "msg": "success"
}
```

---

## POST /api/settings

Update current user's privacy settings. Requires authentication.

**Request Body:**

```json
{
  "post_public": true
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "settings updated"
}
```
