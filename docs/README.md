# Trading Forum API 文档

> 版本：2.0.0  
> 基础 URL：`http://<host>:2333/api`  
> 内容类型：`application/json`（除非另有说明）

---

## 文档目录

| 文档 | 说明 |
|------|------|
| [认证](auth.md) | 注册、登录、退出、Token 机制 |
| [用户](user.md) | 用户信息、资料修改、关注、隐私设置 |
| [帖子](post.md) | 帖子 CRUD、搜索、AI 问答 |
| [评论](comment.md) | 评论 CRUD、树形结构、点赞/点踩 |
| [AI / Agent](ai.md) | Agent CRUD、非流式调用、流式聊天、会话管理 |
| [文件上传](upload.md) | 通用文件分块上传 |
| [视频](video.md) | 视频上传、CRUD、流播放 |
| [行情](market.md) | 市场指数实时/历史数据 |
| [管理后台](admin.md) | 超级管理员运营管理接口 |

---

## 认证方式

使用 **JWT 双 Token** 机制，存储在 **HttpOnly Cookie** 中，前端**无需手动管理 Token**。

详见[认证文档](auth.md)。

---

## 统一响应格式

```json
{
  "code": 0,
  "data": {},
  "msg": "Success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | `0`=成功，`7`=错误 |
| `data` | object/array | 业务数据 |
| `msg` | string | 提示消息 |

**Token 失效响应**（前端需跳转登录页）：

```json
{ "code": 7, "data": { "reload": true }, "msg": "token expired" }
```

---

## 分页规范

**请求参数**（Query String）：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | 页码 |
| `page_size` | int | 否 | 20 | 每页数量（最大 100） |

**响应格式**：

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 42
}
```

---

## SSE 流式格式

AI 相关接口使用 Server-Sent Events。

**响应头**：

```
Content-Type: text/event-stream
Connection: keep-alive
Cache-Control: no-cache
```

**事件格式**（每行以 `data: ` 开头，`\n\n` 结尾）：

```json
data: {"agent_id":1,"chat_id":"uuid","content":"思考中","status":true}
data: {"agent_id":1,"chat_id":"uuid","content":"文本块","status":true}
data: {"agent_id":1,"chat_id":"uuid","content":"","status":true,"message":"done"}
```

**事件类型**：

| 事件 | 说明 |
|------|------|
| 首帧 | 携带 `history` 数组 |
| 流式块 | `content` 为增量文本 |
| 完成帧 | `message` = `"done"` |
| 错误帧 | `status` = `false` |

---

## EnvInfo 结构

部分请求包含可选的环境信息字段：

```json
{
  "ipv4": "192.168.1.1",
  "ipv6": "",
  "os": "Windows 11",
  "device_info": "Chrome 120"
}
```

---

## 枚举常量

### 用户角色（RoleType）

| 值 | 说明 |
|----|------|
| `guest` | 访客（未注册） |
| `evil` | 恶意用户（封禁） |
| `normal_user` | 普通用户（默认） |
| `vip` | VIP 用户 |
| `moderator` | 版主 |
| `takamatsu_tomori` | 超级管理员 |

### 上传状态（UploadStatus）

| 值 | 说明 |
|----|------|
| 1 | Pending（初始化完成） |
| 2 | Uploading（上传中） |
| 3 | Completed（已完成） |
| 4 | Aborted（已取消） |

### AI Provider

| 值 | 说明 |
|----|------|
| `openai` | OpenAI 兼容接口 |
| `ollama` | Ollama 本地模型 |





## 设计风格要求

   - 交易论坛风格 + 年轻化
   - 文章内容渲染规范：
     - Markdown 渲染（图片插入）
     - 话题链接：#话题名 → 跳转搜索页
     - 关键词识别：股票名称 → 跳转东方财富网
