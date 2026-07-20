package entity

import (
	"time"

	"gorm.io/gorm"
)

// Watchlist 自选股关注列表
type Watchlist struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	UserId      uint   `gorm:"uniqueIndex:idx_user_name" json:"user_id"`
	Name        string `gorm:"uniqueIndex:idx_user_name;size:64" json:"name"` // 如 "核心持仓", "观察池"
	Description string `gorm:"size:256" json:"description"`
	IsPublic    bool   `gorm:"default:false" json:"is_public"`
}

// WatchlistItem 自选股明细
type WatchlistItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`

	WatchlistId uint    `gorm:"uniqueIndex:idx_watchlist_symbol" json:"watchlist_id"`
	Symbol      string  `gorm:"uniqueIndex:idx_watchlist_symbol;size:32" json:"symbol"`
	StockName   string  `gorm:"size:64" json:"stock_name"`
	AddedPrice  float64 `json:"added_price"` // 加入时的价格
	Note        string  `gorm:"size:256" json:"note"`
	SortOrder   int     `json:"sort_order"`
}
