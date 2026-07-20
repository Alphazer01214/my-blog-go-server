package request

// WatchlistCreateRequest 创建关注列表请求
type WatchlistCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// WatchlistUpdateRequest 更新关注列表请求
type WatchlistUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    *bool  `json:"is_public"`
}

// WatchlistAddStockRequest 添加股票请求
type WatchlistAddStockRequest struct {
	Symbol    string  `json:"symbol" binding:"required"`
	StockName string  `json:"stock_name"`
	AddedPrice float64 `json:"added_price"`
	Note      string  `json:"note"`
}

// WatchlistListRequest 关注列表查询请求
type WatchlistListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
