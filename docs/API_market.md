# 市场数据 API

- [通用说明](./API_common.md)

---

## 概述

市场模块代理腾讯财经全球指数数据，通过 Redis 缓存（周期由配置决定，默认 5 秒），支持实时数据与历史数据查询。

---

## 1. 获取最新全球指数

**GET** `/api/market` `[公开]`

### 响应 data

```json
{
  "updated_at": "2024-05-15T12:00:00Z",
  "common": [
    {
      "code": "DJI",
      "qtcode": "s_usDJI",
      "name": "道琼斯",
      "location": "纽约",
      "zxj": "50063.46",
      "zdf": "0.75",
      "state": "close"
    }
  ],
  "america": [ ... ],
  "europe": [ ... ],
  "asia": [ ... ],
  "other": [ ... ]
}
```

### IndexItem 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | string | 指数代码 |
| `qtcode` | string | 行情代码 |
| `name` | string | 指数名称 |
| `location` | string | 所在城市/地区 |
| `zxj` | string | 最新价 |
| `zdf` | string | 涨跌幅（%） |
| `state` | string | 市场状态：`open` / `close` / `break` |

---

## 2. 获取历史数据

**GET** `/api/market/history` `[公开]`

> 返回指定时间范围内的指数快照，用于前端绘制走势图

### Query 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `from` | int64 | 1 小时前 | 起始时间（Unix 毫秒） |
| `to` | int64 | 当前时间 | 结束时间（Unix 毫秒） |

### 响应 data

```json
{
  "points": [
    {
      "time": "2024-05-15T12:00:00Z",
      "common": [ { "code": "DJI", "qtcode": "s_usDJI", "name": "道琼斯", "zxj": "50063.46", "zdf": "0.75", "state": "close" } ],
      "america": [ ... ],
      "europe": [ ... ],
      "asia": [ ... ],
      "other": [ ... ]
    }
  ]
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `points` | HistoryPoint[] | 按时间升序排列的历史快照 |
| `points[].time` | ISO8601 | 该快照的抓取时间 |
| `points[].common` 等 | IndexItem[] | 同实时接口的区域指数结构 |

### 使用示例

```
GET /api/market/history?from=1715770000000&to=1715773600000
```

---

## 缓存策略

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `market.refresh_interval` | 5 秒 | 实时数据缓存时长 |
| `market.api_url` | QQ Finance | 上游数据源 |

- 历史数据保留 **24 小时**（Redis sorted set，自动过期）
- 上游故障时实时接口返回空数组 + 当前时间戳
