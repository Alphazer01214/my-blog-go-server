# Admin API

All endpoints under `/api/tomori/` require the `takamatsu_tomori` admin role.

| Method | Endpoint                        | Description              |
| ------ | ------------------------------- | ------------------------ |
| GET    | /api/tomori/stats               | Get system stats         |
| GET    | /api/tomori/users               | List all users           |
| POST   | /api/tomori/user/ban            | Ban/unban user           |
| POST   | /api/tomori/user/role           | Change user role         |
| POST   | /api/tomori/user/reset_password | Reset user password      |
| DELETE | /api/tomori/user/:id            | Delete user              |
| GET    | /api/tomori/posts               | List all posts           |
| DELETE | /api/tomori/post/:id            | Delete post              |
| GET    | /api/tomori/comments            | List all comments        |
| DELETE | /api/tomori/comment/:id         | Delete comment           |
| GET    | /api/tomori/files               | List all files           |
| DELETE | /api/tomori/file/:id            | Delete file              |
| GET    | /api/tomori/agents              | List all agents          |
| DELETE | /api/tomori/agent/:id           | Delete agent             |
| GET    | /api/tomori/chats               | List all chat sessions   |
| DELETE | /api/tomori/chat/:id            | Delete chat session      |
| GET    | /api/tomori/blacklist           | List blacklist           |
| POST   | /api/tomori/blacklist/clear     | Clear blacklist          |

---

## GET /api/tomori/stats

Get system-wide statistics.

**Response:**

```json
{
  "code": 0,
  "data": {
    "user_count": 1500,
    "post_count": 8500,
    "comment_count": 32000,
    "file_count": 2400,
    "agent_count": 120,
    "chat_count": 5600
  },
  "msg": "success"
}
```

---

## GET /api/tomori/users

List all users with pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

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

## POST /api/tomori/user/ban

Ban or unban a user.

**Request Body:**

```json
{
  "user_id": 1,
  "banned": true
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "user banned"
}
```

---

## POST /api/tomori/user/role

Change a user's role.

**Request Body:**

```json
{
  "user_id": 1,
  "role": "takamatsu_tomori"
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "role updated"
}
```

---

## POST /api/tomori/user/reset_password

Reset a user's password.

**Request Body:**

```json
{
  "user_id": 1,
  "new_password": "newpass123"
}
```

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "password reset"
}
```

---

## DELETE /api/tomori/user/:id

Delete a user account.

**Path Parameters:**
- `id` (integer) - User ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "user deleted"
}
```

---

## GET /api/tomori/posts

List all posts with pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...PostDetail..." }
  ],
  "msg": "success"
}
```

---

## DELETE /api/tomori/post/:id

Delete a post as admin.

**Path Parameters:**
- `id` (integer) - Post ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "post deleted"
}
```

---

## GET /api/tomori/comments

List all comments with pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...Comment..." }
  ],
  "msg": "success"
}
```

---

## DELETE /api/tomori/comment/:id

Delete a comment as admin.

**Path Parameters:**
- `id` (integer) - Comment ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "comment deleted"
}
```

---

## GET /api/tomori/files

List all files with pagination.

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

## DELETE /api/tomori/file/:id

Delete a file as admin.

**Path Parameters:**
- `id` (integer) - File ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "file deleted"
}
```

---

## GET /api/tomori/agents

List all AI agents in the system.

**Response:**

```json
{
  "code": 0,
  "data": [
    { "...AgentInfo..." }
  ],
  "msg": "success"
}
```

---

## DELETE /api/tomori/agent/:id

Delete an AI agent as admin.

**Path Parameters:**
- `id` (integer) - Agent ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "agent deleted"
}
```

---

## GET /api/tomori/chats

List all chat sessions with pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `page_size` (integer, default: 10) - Items per page

**Response:**

```json
{
  "code": 0,
  "data": [
    {
      "id": "chat-uuid",
      "agent_id": 1,
      "user_id": 1,
      "title": "Chat session",
      "messages": [],
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ],
  "msg": "success"
}
```

---

## DELETE /api/tomori/chat/:id

Delete a chat session as admin.

**Path Parameters:**
- `id` (integer/string) - Chat session ID

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "chat deleted"
}
```

---

## GET /api/tomori/blacklist

List blacklisted entries with pagination.

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
      "user_id": 5,
      "reason": "spam",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "msg": "success"
}
```

---

## POST /api/tomori/blacklist/clear

Clear all blacklist entries.

**Response:**

```json
{
  "code": 0,
  "data": {},
  "msg": "blacklist cleared"
}
```
