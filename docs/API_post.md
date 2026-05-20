# 帖子 API

- [通用说明](./API_common.md)

---

## 1. 创建帖子

**POST** `/api/create` `[认证]`

### 请求体

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "title": "string",
  "cover": "string",
  "category_id": 1,
  "tags": ["tag1", "tag2"],
  "keywords": ["key1", "key2"],
  "content": "string",
  "public": true,
  "forbid_comment": false,
  "forbid_share": false
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `env` | EnvInfo | 否 | 客户端环境信息 |
| `title` | string | 是 | 标题 |
| `cover` | string | 否 | 封面图 URL |
| `category_id` | uint | 否 | 分类 ID |
| `tags` | JSON | 否 | 标签数组 |
| `keywords` | JSON | 否 | 关键词数组 |
| `content` | string | 是 | 正文内容 |
| `public` | bool | 否 | 是否公开（受用户设置影响） |
| `forbid_comment` | bool | 否 | 禁止评论 |
| `forbid_share` | bool | 否 | 禁止分享 |

### 响应 data

`PostDetail` 结构，同 [查询单个帖子](#3-查询单个帖子)

---

## 2. 查询帖子列表

**GET** `/api/post` `[认证]`

### Query 参数

| 参数 | 类型 | 默认值 |
|------|------|--------|
| `page` | int | 1 |
| `page_size` | int | 20（最大 100） |

### 响应 data

```json
{
  "items": [PostDetail, ...],
  "page": 1,
  "page_size": 20,
  "total": 50
}
```

> 非作者查看私密帖子时 `content` 会被替换为 `"this is private post"`

---

## 3. 查询单个帖子

**GET** `/api/post/:id` `[认证]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 帖子 ID |

### 响应 data

`PostDetail` 结构：

```json
{
  "id": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "title": "帖子标题",
  "cover": "https://example.com/cover.jpg",
  "user_id": 1,
  "author": { "user_id": 1, "username": "author", "..." : "..." },
  "tags": ["Go", "后端"],
  "category_id": 1,
  "category_name": "技术",
  "keywords": ["tutorial"],
  "content": "正文内容",
  "view_count": 100,
  "comment_count": 5,
  "like_count": 10,
  "dislike_count": 1,
  "favorite_count": 3,
  "share_count": 2,
  "public": true,
  "forbid_comment": false,
  "forbid_share": false,
  "is_liked": false,
  "is_disliked": false,
  "is_favorited": false
}
```

> 非作者查看私密帖子时 `content` 会被替换为 `"this is private post"`

---

## 4. 更新帖子

**POST** `/api/update` `[认证]`

> 仅帖子作者可更新

### 请求体

```json
{
  "post_id": 1,
  "title": "string",
  "cover": "string",
  "category_id": 1,
  "tags": ["tag1"],
  "keywords": ["key1"],
  "content": "string",
  "public": true,
  "forbid_comment": false,
  "forbid_share": false
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `post_id` | uint | 是 | 帖子 ID |
| `title` | string | 是 | 标题 |
| `cover` | string | 否 | 封面图 |
| `category_id` | uint | 否 | 分类 ID |
| `tags` | JSON | 否 | 标签 |
| `keywords` | JSON | 否 | 关键词 |
| `content` | string | 是 | 正文 |
| `public` | bool | 否 | 是否公开 |
| `forbid_comment` | bool | 否 | 禁止评论 |
| `forbid_share` | bool | 否 | 禁止分享 |

> 所有字段会**整体替换**，不传的字段可能被置空

### 响应 data

更新前的 `PostDetail` 结构

---

## 5. 删除帖子

**DELETE** `/api/post/:id` `[认证]`

> 仅帖子作者可删除

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 帖子 ID |

### 响应 data

空对象

---

## 6. 点赞 / 取消点赞

**POST** `/api/post/like` `[认证]`

> Toggle 模式：已点赞则取消，未点赞则点赞（如有踩则先移除踩）

### 请求体

```json
{
  "post_id": 1
}
```

### 响应 data

空对象

---

## 7. 点踩 / 取消点踩

**POST** `/api/post/dislike` `[认证]`

> Toggle 模式：已踩则取消，未踩则点踩（如有赞则先移除赞）

### 请求体

```json
{
  "post_id": 1
}
```

### 响应 data

空对象

---

## 8. 收藏 / 取消收藏

**POST** `/api/post/favorite` `[认证]`

> Toggle 模式：已收藏则取消，未收藏则收藏

### 请求体

```json
{
  "post_id": 1
}
```

### 响应 data

空对象

---

## 9. 分享帖子

**POST** `/api/post/share` `[认证]`

> 每次请求都会记录一条分享记录（非 toggle）

### 请求体

```json
{
  "post_id": 1,
  "share_to": "wechat",
  "share_message": "这篇写的不错"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `post_id` | uint | 是 | 帖子 ID |
| `share_to` | string | 否 | 分享目标平台 |
| `share_message` | string | 否 | 分享附言 |

### 响应 data

空对象
