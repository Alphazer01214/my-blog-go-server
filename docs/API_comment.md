# 评论 API

- [通用说明](./API_common.md)

---

## 类型枚举

### TargetType（评论目标类型）

| 值 | 常量 | 说明 |
|----|------|------|
| 1 | `TargetPost` | 帖子 |
| 2 | `TargetComment` | 评论 |
| 3 | `TargetVideo` | 视频/文件 |
| 4 | `TargetUser` | 用户 |
| 5 | `TargetTag` | 标签 |

> **当前实现**：后端目前仅支持对帖子评论，`target_type` 固定为 `1`（TargetPost）。前端传 `post_id` 即可，其余 `target_type` 后续扩展。

### ActionType（交互行为类型）

| 值 | 常量 | 说明 |
|----|------|------|
| 1 | `ActionLike` | 点赞 |
| 2 | `ActionDislike` | 点踩 |
| 3 | `ActionFavorite` | 收藏 |
| 4 | `ActionShare` | 分享 |

> 前端在点赞/点踩接口中无需传 `action_type`，后端根据接口自动确定。

---

## 评论树结构说明

评论采用 **N 级递归树结构**，通过 `root_comment_id` 和 `parent_comment_id` 两个字段定位层级关系：

```
帖子 / 视频
├── 根评论 A  (root=0, parent=0)          ← 直接在帖子/视频下评论
│   ├── 回复 A1 (root=A, parent=A)        ← 回复根评论 A
│   │   ├── 回复 A1a (root=A, parent=A1)  ← 回复子回复 A1，root 始终指向 A
│   │   └── 回复 A1b (root=A, parent=A1)
│   └── 回复 A2 (root=A, parent=A)        ← 回复根评论 A
├── 根评论 B  (root=0, parent=0)
│   └── 回复 B1 (root=B, parent=B)
```

### 规则速查

| 场景 | `root_comment_id` | `parent_comment_id` |
|------|-------------------|----------------------|
| 在帖子/视频评论区直接发言 | `0` | `0` |
| 回复某条**根评论** | 被回复的根评论 ID | 被回复的根评论 ID（与 root 相同） |
| 回复某条**子回复** | 该子回复所属的根评论 ID | 被回复的子回复 ID |

> **核心原则**：`root_comment_id` 始终指向**整棵评论树的根节点**，`parent_comment_id` 指向**直接父评论**。根评论是直接挂在帖子/视频下的评论。

### 响应递归嵌套

响应中的 `reply_comments` 字段**递归嵌套**该评论的所有子回复，无深度限制：

```json
{
  "comment_id": 1,           // 根评论
  "root_comment_id": 0,
  "parent_comment_id": 0,
  "reply_comments": [
    {
      "comment_id": 2,       // 对根评论的直接回复
      "root_comment_id": 1,
      "parent_comment_id": 1,
      "reply_comments": [
        {
          "comment_id": 3,   // 对回复的回复
          "root_comment_id": 1,
          "parent_comment_id": 2,
          "reply_comments": []
        }
      ]
    }
  ]
}
```

> **前端注意**：渲染时递归遍历 `reply_comments`，建议设置最大渲染深度或用"展开更多"按钮防止 DOM 过深。

---

## 1. 创建评论

**POST** `/api/comment` `[认证]`

### 请求体

```json
{
  "env": { "ipv4": "", "ipv6": "", "os": "", "device_info": "" },
  "post_id": 1,
  "content": "评论内容",
  "root_comment_id": 0,
  "parent_comment_id": 0
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `env` | EnvInfo | 否 | 客户端环境 |
| `post_id` | uint | 是 | 评论目标 ID（当前仅支持帖子 ID，对应 `target_id`，`target_type` 固定为 `1`） |
| `content` | string | 是 | 评论内容 |
| `root_comment_id` | uint | 否 | 根评论 ID，直接评论帖子/视频时传 `0` |
| `parent_comment_id` | uint | 否 | 父评论 ID，直接评论帖子/视频时传 `0`；回复时传被回复的评论 ID |

### 响应 data

```json
{
  "comment_id": 1,
  "user_id": 2,
  "post_id": 1,
  "root_comment_id": 0,
  "parent_comment_id": 0,
  "content": "评论内容",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "author": { "user_id": 2, "username": "commenter", "..." : "..." },
  "like_count": 0,
  "dislike_count": 0,
  "reply_count": 0,
  "is_liked": false,
  "is_disliked": false,
  "reply_comments": []
}
```

> **注意**：`post_id` 字段名为历史遗留，实为 `target_id`。评论挂载的目标由后端 `target_type`（固定 1）+ `target_id` 决定。

---

## 2. 查询帖子评论

**GET** `/api/post/:id/comments` `[公开]`

> 分页获取帖子的根评论，每个根评论的 `reply_comments` 递归包含其所有子回复树。
> 未登录时 `is_liked` / `is_disliked` 恒为 `false`。

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 帖子 ID |

### Query 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码 |
| `page_size` | int | 20 | 每页数量（最大 100） |

### 响应 data

```json
{
  "items": [Comment, ...],
  "page": 1,
  "page_size": 20,
  "total": 1
}
```

> `items` 为当前页的根评论列表，每个根评论的 `reply_comments` 递归包含其所有子回复（不分页，返回全部）。
> `total` 为数据库中符合条件的根评论总数。

---

## 3. 查询用户评论

**GET** `/api/user/:id/comments` `[公开]`

> 获取用户发布的评论列表（平铺，无树结构）。
> 受用户隐私设置控制：若用户关闭评论公开（`comment_public = false`），则返回空列表。
> 未登录时 `is_liked` / `is_disliked` 恒为 `false`。

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户 ID |

### Query 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码 |
| `page_size` | int | 20 | 每页数量（最大 100） |

### 响应 data

```json
{
  "items": [Comment, ...],
  "page": 1,
  "page_size": 20,
  "total": 1
}
```

> 返回平铺评论列表。
> `total` 为数据库中该用户的评论总数。

---

## 4. 删除评论

**DELETE** `/api/comment/:id` `[认证]`

### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 评论 ID |

### 权限说明

- 仅评论作者本人可删除
- 删除根评论会级联删除其所有子回复，并更新帖子的 `comment_count`
- 删除子回复会更新根评论的 `reply_count`、父评论的 `reply_count` 以及帖子的 `comment_count`

### 响应 data

空对象

---

## 5. 评论点赞 / 取消点赞

**POST** `/api/comment/like` `[认证]`

> Toggle 模式：已赞则取消点赞，未赞则点赞（如果已踩会先移除踩）

### 请求体

```json
{
  "comment_id": 1
}
```

### 响应 data

空对象

---

## 6. 评论点踩 / 取消点踩

**POST** `/api/comment/dislike` `[认证]`

> Toggle 模式：已踩则取消点踩，未踩则点踩（如果已赞会先移除赞）

### 请求体

```json
{
  "comment_id": 1
}
```

### 响应 data

空对象
