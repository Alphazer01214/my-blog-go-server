package entity

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TradeSignal 交易信号
type TradeSignal struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	UserId    uint   `gorm:"index" json:"user_id"`
	Symbol    string `gorm:"size:32;index" json:"symbol"`    // 如 sh600519
	StockName string `gorm:"size:64" json:"stock_name"`     // 股票名称

	// 信号方向
	Direction string `gorm:"size:16" json:"direction"` // long/short/hold

	// 关键价位
	EntryPrice  float64 `json:"entry_price"`  // 入场价
	TargetPrice float64 `json:"target_price"` // 目标价
	StopLoss    float64 `json:"stop_loss"`    // 止损价

	// 信号状态
	Status      string     `gorm:"size:16;default:'active'" json:"status"` // active/closed/expired
	ClosedPrice *float64   `json:"closed_price,omitempty"`                 // 平仓价
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
	PnL         *float64   `json:"pnl,omitempty"` // 盈亏百分比

	// 分析内容
	Title    string         `gorm:"type:varchar(255)" json:"title"`
	Analysis string         `gorm:"type:text" json:"analysis"`
	Tags     datatypes.JSON `json:"tags"`

	// 统计
	ViewCount   int `json:"view_count"`
	LikeCount   int `json:"like_count"`
	FollowCount int `json:"follow_count"` // 跟单人数
}

// SignalFollow 跟单记录
type SignalFollow struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	UserId   uint      `gorm:"uniqueIndex:idx_user_signal" json:"user_id"`
	SignalId uint      `gorm:"uniqueIndex:idx_user_signal" json:"signal_id"`
	Notified bool      `gorm:"default:false" json:"notified"`
	CreatedAt time.Time `json:"created_at"`
}
