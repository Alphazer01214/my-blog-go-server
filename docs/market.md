# 行情 API

市场指数实时/历史数据，汇率查询。

## 目录

| 端点 | 方法 | 说明 | 是否需要登录 |
|------|------|------|------------|
| `/api/market` | GET | 获取所有指数最新数据 | 否 |
| `/api/market/history` | GET | 获取指数历史数据（默认最近10分钟） | 否 |
| `/api/market/stream` | GET | SSE 实时推送市场数据 | 否 |
| `/api/market/exchange-rate` | GET | 查询汇率 | 否 |
| `/api/market/currencies` | GET | 获取支持的货币代码列表 | 否 |

---

## 端点详情

### GET /api/market

获取所有市场指数的最新行情。

**Response**：

```json
{
  "code": 0,
  "data": {
    "updated_at": "2024-05-25T16:00:00Z",
    "common": [
      {
        "code": "sh000001",
        "qtcode": "sh000001",
        "name": "上证指数",
        "location": "上海",
        "zxj": "3200.50",
        "zdf": "+0.48%",
        "state": "交易中"
      }
    ],
    "america": [
      {
        "code": ".DJI",
        "qtcode": ".DJI",
        "name": "道琼斯",
        "location": "美国",
        "zxj": "39000.50",
        "zdf": "+0.32%",
        "state": "交易中"
      }
    ],
    "europe": [],
    "asia": [],
    "other": []
  },
  "msg": "Success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `updated_at` | string | 数据更新时间 |
| `common` | array | 国内指数列表 |
| `america` | array | 美洲指数列表 |
| `europe` | array | 欧洲指数列表 |
| `asia` | array | 亚洲指数列表（不含国内） |
| `other` | array | 其他指数列表 |

**IndexItem 字段**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | string | 指数代码 |
| `qtcode` | string | 行情代码 |
| `name` | string | 指数名称 |
| `location` | string | 所属地区 |
| `zxj` | string | 最新价 |
| `zdf` | string | 涨跌幅 |
| `state` | string | 状态 |

### GET /api/market/history

获取指数在时间范围内的历史数据。历史数据仅保留最近10分钟。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `from` | int | 否 | 起始时间戳（毫秒），默认为10分钟前 |
| `to` | int | 否 | 结束时间戳（毫秒），默认为当前时间 |

**Response**：

```json
{
  "code": 0,
  "data": {
    "points": [
      {
        "time": "2024-05-25T16:00:00Z",
        "common": [
          {
            "code": "sh000001",
            "qtcode": "sh000001",
            "name": "上证指数",
            "location": "上海",
            "zxj": "3185.30",
            "zdf": "-0.15%",
            "state": "交易中"
          }
        ],
        "america": [],
        "europe": [],
        "asia": [],
        "other": []
      }
    ]
  },
  "msg": "Success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `points` | array | 历史数据点列表 |
| `points[].time` | string | 数据时间 |
| `points[].common` | array | 国内指数快照 |
| `points[].america` | array | 美洲指数快照 |
| `points[].europe` | array | 欧洲指数快照 |
| `points[].asia` | array | 亚洲指数快照 |
| `points[].other` | array | 其他指数快照 |

### GET /api/market/stream

SSE 实时推送市场数据，每5秒推送一次。

**请求头**：

```
Accept: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Response**（SSE 格式）：

```
data: {"updated_at":"2024-05-25T16:00:00Z","common":[...],"america":[...],"europe":[...],"asia":[...],"other":[...]}

data: {"updated_at":"2024-05-25T16:00:05Z","common":[...],"america":[...],"europe":[...],"asia":[...],"other":[...]}
```

数据格式与 `GET /api/market` 的 `data` 字段相同。

### GET /api/market/exchange-rate

查询货币汇率。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `from` | string | 是 | 源货币代码（如：USD） |
| `to` | string | 是 | 目标货币代码（如：CNY） |
| `money` | number | 否 | 金额，默认为1 |

**请求示例**：

```
GET /api/market/exchange-rate?from=USD&to=CNY&money=100
```

**Response**：

```json
{
  "code": 0,
  "data": {
    "uptime": "2025-12-20 08:00:01",
    "from": "USD",
    "to": "CNY",
    "money": "100",
    "result": 705.11,
    "rate": 7.0511
  },
  "msg": "Success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `uptime` | string | 汇率更新时间 |
| `from` | string | 源货币代码 |
| `to` | string | 目标货币代码 |
| `money` | string | 原始金额 |
| `result` | number | 换算结果 |
| `rate` | number | 汇率 |

### GET /api/market/currencies

获取支持的货币代码列表。

**Response**：

```json
{
  "code": 0,
  "data": [
    {
      "code": "USD",
      "name": "美元"
    },
    {
      "code": "CNY",
      "name": "人民币"
    },
    {
      "code": "EUR",
      "name": "欧元"
    },
    {
      "code": "JPY",
      "name": "日元"
    },
    {
      "code": "GBP",
      "name": "英镑"
    }
  ],
  "msg": "Success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | string | 货币代码 |
| `name` | string | 货币名称 |

---

## 主流货币代码参考

| 代码 | 货币名称 |
|------|----------|
| USD | 美元 |
| CNY | 人民币 |
| EUR | 欧元 |
| JPY | 日元 |
| GBP | 英镑 |
| KRW | 韩元 |
| HKD | 港币 |
| TWD | 新台币 |
| SGD | 新加坡元 |
| AUD | 澳元 |
| CAD | 加元 |
| CHF | 瑞士法郎 |

---

## 配置说明

在 `config.yaml` 中配置：

```yaml
market:
  refresh_interval: 5  # 数据刷新间隔（秒）
  api_url: "https://proxy.finance.qq.com/ifzqgtimg/appstock/app/rank/indexRankDetail2"
  exchange_rate_url: "https://cn.apihz.cn/api/jinrong/huilv.php"
```
