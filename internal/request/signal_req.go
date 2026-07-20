package request

import "gorm.io/datatypes"

// SignalCreateRequest 创建交易信号请求
type SignalCreateRequest struct {
	Symbol      string         `json:"symbol" binding:"required"`
	StockName   string         `json:"stock_name"`
	Direction   string         `json:"direction" binding:"required"` // long/short/hold
	EntryPrice  float64        `json:"entry_price" binding:"required"`
	TargetPrice float64        `json:"target_price"`
	StopLoss    float64        `json:"stop_loss"`
	Title       string         `json:"title" binding:"required"`
	Analysis    string         `json:"analysis"`
	Tags        datatypes.JSON `json:"tags"`
}

// SignalUpdateRequest 更新交易信号请求
type SignalUpdateRequest struct {
	Status      string   `json:"status"` // closed
	ClosedPrice *float64 `json:"closed_price"`
}

// SignalListRequest 信号列表查询请求
type SignalListRequest struct {
	Symbol    string `form:"symbol"`
	Direction string `form:"direction"`
	Status    string `form:"status"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}
