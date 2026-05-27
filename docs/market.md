# 行情 API

市场指数实时/历史数据。

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/market/indices` | GET | 获取所有指数最新数据 | 否 |
| `/api/market/indices/history` | GET | 获取指定指数历史数据 | 否 |

---

## 端点详情

### GET /api/market/indices

获取所有市场指数的最新行情。

**Response**：

```json
{
  "data": [
    {
      "index_name": "A股大盘",
      "short_name": "A股",
      "current_value": 3200.50,
      "change_value": 15.20,
      "change_percent": 0.48,
      "is_up": true,
      "type": "stock"
    }
  ]
}
```

- `is_up`：`true`=上涨，`false`=下跌
- `type`：指数类型标识

### GET /api/market/indices/history

获取指定指数在时间范围内的历史数据。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `index_name` | string | 是 | | 指数名称 |
| `from` | int | 是 | | 起始时间戳（毫秒） |
| `to` | int | 是 | | 结束时间戳（毫秒） |

**Response**：

```json
{
  "data": [
    {
      "index_name": "A股大盘",
      "value": 3185.30,
      "change_value": 0.48,
      "change_percent": 0.48,
      "timestamp": 1716624000000,
      "time": "2024-05-25 16:00:00"
    }
  ]
}
```
