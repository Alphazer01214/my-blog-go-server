# 压力测试脚本

本目录包含 TradingForumGo 项目的压力测试脚本。

## 目录结构

```
test/
├── http/                    # HTTP 请求文件（用于 VS Code REST Client 插件）
│   └── register.http       # 注册、登录等接口测试
├── k6/                      # k6 压测脚本
│   ├── login_test.js       # 登录接口压测
│   ├── register_test.js    # 注册接口压测
│   ├── market_test.js      # 行情数据接口压测
│   ├── post_test.js        # 帖子接口压测
│   ├── sse_test.js         # SSE 流式推送压测
│   └── mixed_test.js       # 混合场景压测
└── README.md               # 本文件
```

## 前置条件

### 1. 安装 k6

**Windows (Chocolatey)**
```bash
choco install k6
```

**macOS (Homebrew)**
```bash
brew install k6
```

**Docker**
```bash
docker pull grafana/k6
```

### 2. 启动服务

确保 PostgreSQL、Redis 和应用服务已启动：

```bash
# 在项目根目录
go run main.go
```

### 3. 准备测试数据

运行注册脚本创建测试用户：

```bash
k6 run test/k6/register_test.js
```

然后在数据库中手动创建一个测试用户（用于登录测试）：

```sql
-- 连接到 PostgreSQL
psql -U tomori -d tomori_db

-- 插入测试用户（密码为 123456 的 bcrypt 哈希）
INSERT INTO users (username, password, role, banned) 
VALUES ('testuser', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'normal_user', false);
```

## 运行测试

### 运行单个测试

```bash
# 登录压测
k6 run test/k6/login_test.js

# 注册压测
k6 run test/k6/register_test.js

# 行情数据压测
k6 run test/k6/market_test.js

# 帖子压测
k6 run test/k6/post_test.js

# SSE 压测
k6 run test/k6/sse_test.js

# 混合场景压测
k6 run test/k6/mixed_test.js
```

### 自定义参数运行

```bash
# 自定义并发数和持续时间
k6 run --vus 100 --duration 60s test/k6/login_test.js

# 输出 JSON 报告
k6 run --out json=results.json test/k6/login_test.js

# 输出到 InfluxDB（配合 Grafana 可视化）
k6 run --out influxdb=http://localhost:8086/k6 test/k6/login_test.js
```

### Docker 运行

```bash
# 使用 Docker 运行
docker run --rm -i -v $(pwd)/test/k6:/scripts grafana/k6 run /scripts/login_test.js
```

## 测试场景说明

### login_test.js - 登录压测
- **目标**: 测试登录接口性能
- **并发**: 50 → 100 用户
- **持续时间**: 约 3.5 分钟
- **关注指标**: QPS、P95 响应时间、错误率

### register_test.js - 注册压测
- **目标**: 测试注册接口性能
- **并发**: 20 用户
- **持续时间**: 约 2 分钟
- **关注指标**: 注册成功率、响应时间

### market_test.js - 行情数据压测
- **目标**: 测试行情接口性能（无认证）
- **并发**: 100 → 200 用户
- **持续时间**: 约 3.5 分钟
- **关注指标**: QPS、缓存命中率

### post_test.js - 帖子压测
- **目标**: 测试帖子 CRUD 性能
- **并发**: 50 → 100 用户
- **持续时间**: 约 3.5 分钟
- **操作分布**: 60% 列表、20% 搜索、20% 详情

### sse_test.js - SSE 流式推送压测
- **目标**: 测试 SSE 连接稳定性
- **并发**: 50 → 100 连接
- **持续时间**: 约 3.5 分钟
- **关注指标**: 连接数、断连率

### mixed_test.js - 混合场景压测
- **目标**: 模拟真实用户行为
- **并发**: 50 → 100 用户
- **持续时间**: 约 7 分钟
- **操作分布**: 
  - 50% 读操作（列表、搜索、行情）
  - 30% 写操作（创建帖子）
  - 20% 认证操作（登录）

## 关键指标说明

| 指标 | 说明 | 目标值 |
|------|------|--------|
| **QPS (http_reqs)** | 每秒请求数 | >500 |
| **Avg Response Time** | 平均响应时间 | <200ms |
| **P95 Response Time** | 95% 请求的响应时间 | <500ms |
| **P99 Response Time** | 99% 请求的响应时间 | <1000ms |
| **Error Rate (http_req_failed)** | 错误率 | <1% |
| **Data Received** | 接收数据量 | - |
| **Data Sent** | 发送数据量 | - |

## 性能瓶颈分析

| 问题 | 可能原因 | 优化建议 |
|------|---------|---------|
| QPS 低 | 数据库连接池不足 | 增加 `max_open_conns` |
| 响应时间长 | SQL 查询慢 | 添加索引、优化查询 |
| 错误率高 | 连接数超限 | 增加系统 `ulimit` |
| 内存溢出 | 连接泄漏 | 检查连接关闭 |

## 可视化监控（可选）

### 启动 InfluxDB + Grafana

```bash
# 启动 InfluxDB
docker run -d -p 8086:8086 --name influxdb influxdb:1.8

# 启动 Grafana
docker run -d -p 3000:3000 --name grafana grafana/grafana
```

### 运行测试并输出到 InfluxDB

```bash
k6 run --out influxdb=http://localhost:8086/k6 test/k6/mixed_test.js
```

### 在 Grafana 中查看

1. 打开 http://localhost:3000
2. 添加 InfluxDB 数据源（URL: http://influxdb:8086）
3. 导入 k6 Dashboard 模板（ID: 2587）

## 使用 VS Code REST Client

安装 [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) 插件后，可以直接在 VS Code 中运行 `.http` 文件：

1. 打开 `test/http/register.http`
2. 点击 "Send Request" 按钮
3. 查看响应结果

## 注意事项

1. **测试环境**: 建议在独立的测试环境运行，避免影响开发环境
2. **数据清理**: 压测后清理测试数据，避免污染数据库
3. **资源监控**: 压测时监控 CPU、内存、网络、数据库连接数
4. **逐步加压**: 使用 stages 逐步增加压力，观察系统行为
