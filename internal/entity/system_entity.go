package entity

import "time"

// Notification 通知实体
type Notification struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserId     uint      `gorm:"index" json:"user_id"`        // 接收者
	Type       string    `gorm:"size:32" json:"type"`         // like/comment/follow/unlike
	SourceId   uint      `json:"source_id"`                   // 来源对象ID
	SourceType string    `gorm:"size:32" json:"source_type"`  // post/comment/user
	Content    string    `gorm:"type:text" json:"content"`    // 通知内容摘要
	IsRead     bool      `gorm:"default:false" json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}
