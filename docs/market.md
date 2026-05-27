# Market API

| Method | Endpoint           | Description            |
| ------ | ------------------ | ---------------------- |
| GET    | /api/market        | Get market overview    |
| GET    | /api/market/history| Get market history     |

---

## GET /api/market

Get current market overview data.

**Response:**

```json
{
  "code": 0,
  "data": {
    "markets": [
      {
        "symbol": "BTC/USDT",
        "price": 65000.00,
        "change": 2.5,
        "volume": 1234567890
      }
    ]
  },
  "msg": "success"
}
```

---

## GET /api/market/history

Get historical market data within a time range.

**Query Parameters:**
- `from` (string, required) - Start time (ISO 8601 or timestamp)
- `to` (string, required) - End time (ISO 8601 or timestamp)

**Example:** `GET /api/market/history?from=2026-01-01&to=2026-01-31`

**Response:**

```json
{
  "code": 0,
  "data": {
    "history": [
      {
        "timestamp": "2026-01-01T00:00:00Z",
        "open": 64000.00,
        "high": 65500.00,
        "low": 63500.00,
        "close": 65000.00,
        "volume": 987654321
      }
    ]
  },
  "msg": "success"
}
```
