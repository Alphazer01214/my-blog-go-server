package response

import (
	"time"

	"gorm.io/datatypes"
)

// SignalDetail 交易信号详情
type SignalDetail struct {
	ID          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	UserId      uint           `json:"user_id"`
	Author      UserInfo       `json:"author,omitempty"`
	Symbol      string         `json:"symbol"`
	StockName   string         `json:"stock_name"`
	Direction   string         `json:"direction"`
	EntryPrice  float64        `json:"entry_price"`
	TargetPrice float64        `json:"target_price"`
	StopLoss    float64        `json:"stop_loss"`
	Status      string         `json:"status"`
	ClosedPrice *float64       `json:"closed_price,omitempty"`
	ClosedAt    *time.Time     `json:"closed_at,omitempty"`
	PnL         *float64       `json:"pnl,omitempty"`
	Title       string         `json:"title"`
	Analysis    string         `json:"analysis"`
	Tags        datatypes.JSON `json:"tags"`
	ViewCount   int            `json:"view_count"`
	LikeCount   int            `json:"like_count"`
	FollowCount int            `json:"follow_count"`
	IsFollowing bool           `json:"is_following"` // 当前用户是否在跟单
}

// SignalList 信号列表
type SignalList struct {
	Items    []SignalDetail `json:"items"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Total    int64          `json:"total"`
}

// WatchlistDetail 关注列表详情
type WatchlistDetail struct {
	ID          uint               `json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	UserId      uint               `json:"user_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	IsPublic    bool               `json:"is_public"`
	ItemCount   int                `json:"item_count"`
	Items       []WatchlistItemDetail `json:"items,omitempty"`
}

// WatchlistItemDetail 关注列表股票详情
type WatchlistItemDetail struct {
	ID         uint      `json:"id"`
	Symbol     string    `json:"symbol"`
	StockName  string    `json:"stock_name"`
	AddedPrice float64   `json:"added_price"`
	Note       string    `json:"note"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
}

// WatchlistList 关注列表
type WatchlistList struct {
	Items    []WatchlistDetail `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
}
