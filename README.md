# Trading Forum Go

一个基于 Go 语言开发的全功能交易社区后端 API 服务，整合传统论坛功能与 AI 智能助手、实时行情数据推送。

## 技术栈

| 层级 | 技术 |
|------|------|
| 语言/框架 | Go 1.22+, Gin, GORM |
| 数据库 | PostgreSQL 16, Redis 9 |
| 认证 | JWT 双令牌 (Access + Refresh Token) |
| AI | 字节 Eino 框架 + OpenAI/Ollama |
| 视频 | FFmpeg 封面提取 + 分片上传 |
| 部署 | Docker 多阶段构建 |

## 项目结构

```
trading-forum-go/
├── main.go              # 入口文件
├── cmd/                 # CLI 命令
├── config/              # 配置文件
├── internal/
│   ├── api/             # HTTP 处理器
│   ├── service/         # 业务逻辑层
│   ├── entity/          # 数据模型
│   ├── repository/      # 数据访问层
│   ├── router/          # 路由注册
│   ├── agent/           # AI Agent 工具
│   ├── keepalive/       # 后台任务调度
│   ├── utils/           # 工具函数
│   └── ...
├── pkg/
│   └── middleware/       # 中间件
├── docs/                # API 文档
└── Dockerfile
```

## 核心功能

### 1. 用户系统
- 注册、登录、登出
- JWT 双令牌认证 + 自动刷新
- RBAC 角色权限控制 (guest/evil/normal_user/vip/moderator/takamatsu_tomori)
- 用户关注/粉丝系统
- 隐私设置

### 2. 帖子系统
- 完整 CRUD + Markdown 支持
- 点赞/踩/收藏/分享 (Toggle 语义)
- 全文搜索
- 评论数/浏览量统计
- 帖子级可见性控制

### 3. 评论系统
- 树形评论结构 (根评论 + 嵌套回复)
- 递归构建评论树
- 级联删除 + 计数器恢复

### 4. AI Agent 系统
- 用户自定义 Agent (独立模型配置)
- 支持 OpenAI/Ollama 多模型
- ReAct 工具调用循环 (网络搜索/论坛搜索/股票分析)
- SSE 流式响应 + Redis 缓存防断连
- 聊天会话持久化

### 5. 视频系统
- 分片上传 + Redis 会话管理
- SHA-256 校验 + 自动重试
- FFmpeg 封面提取 + 时长检测
- HTTP Range 流式播放

### 6. 行情系统
- 后台协程定时轮询
- Redis Sorted Set 缓存 + TTL 清理
- SSE 实时推送全球指数
- 汇率查询

### 7. 管理后台
- 统计面板
- 用户/帖子/评论/视频/AI 管理
- Token 黑名单管理

## 快速开始

### 环境要求

- Go 1.22+
- PostgreSQL 16+
- Redis 6+
- FFmpeg (视频功能)

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/trading-forum-go.git
cd trading-forum-go
```

### 2. 配置

复制并修改配置文件：

```bash
cp config/config.yaml.example config/config.yaml
```

编辑 `config/config.yaml` 填入你的数据库、Redis、LLM API Key 等配置。

### 3. 运行

```bash
# 自动迁移数据库
go run main.go -migrate

# 启动服务
go run main.go
```

服务默认运行在 `http://localhost:2333`

### Docker 部署

```bash
docker build -t trading-forum-go .
docker run -p 2333:2333 trading-forum-go
```

## API 文档

详细 API 文档请查看 [docs/](./docs/) 目录：

- [认证 API](./docs/auth.md)
- [用户 API](./docs/user.md)
- [帖子 API](./docs/post.md)
- [评论 API](./docs/comment.md)
- [视频 API](./docs/video.md)
- [AI API](./docs/ai.md)
- [行情 API](./docs/market.md)
- [上传 API](./docs/upload.md)
- [管理 API](./docs/admin.md)

## 配置说明

`config/config.yaml` 主要配置项：

| 配置项 | 说明 |
|--------|------|
| `server.port` | 服务端口 |
| `server.mode` | 运行模式 (debug/test/release) |
| `postgres.*` | PostgreSQL 连接配置 |
| `redis.*` | Redis 连接配置 |
| `llm.*` | LLM 模型配置 |
| `jwt.*` | JWT 密钥和过期时间 |
| `market.*` | 行情数据源配置 |

## 开发

### 目录结构说明

- `internal/api`: HTTP 处理器，处理请求和响应
- `internal/service`: 业务逻辑层
- `internal/entity`: GORM 数据模型
- `internal/repository`: 数据访问层
- `internal/router`: 路由注册
- `internal/agent`: AI Agent 工具实现
- `internal/keepalive`: 后台任务调度器
- `pkg/middleware`: CORS、JWT 认证中间件

### 数据库迁移

启动时自动迁移：

```bash
go run main.go -migrate
```

## License

MIT
