package entity

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Post struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	EnvInfo   `json:"env"`

	Title      string         `gorm:"type:varchar(255)" json:"title"`
	Cover      string         `json:"cover"`
	UserId     uint           `json:"user_id"`
	Tags       datatypes.JSON `json:"tags"`
	CategoryId uint           `json:"category_id"`
	Keywords   datatypes.JSON `json:"keywords"`
	Content    string         `gorm:"type:text" json:"content"`

	ViewCount     int `json:"view_count"`
	CommentCount  int `json:"comment_count"`
	LikeCount     int `json:"like_count"`
	DislikeCount  int `json:"dislike_count"`
	FavoriteCount int `json:"favorite_count"`
	ShareCount    int `json:"share_count"`

	Public bool `json:"public"`
}

type Tag struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `json:"name" gorm:"uniqueIndex"`
	UsageCount int    `json:"usage_count" gorm:"default:0"`
}

type Category struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name" gorm:"uniqueKey"`
	Description string `json:"description"`
	PostCount   int    `json:"post_count" gorm:"default:0"`
}

type Comment struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserId uint `json:"user_id"`
	PostId uint `json:"post_id"`
	// RootCommentId 根评论id，如果为0，则表示该评论为根评论
	RootCommentId   uint           `json:"root_comment_id"`
	ParentCommentId uint           `json:"parent_comment_id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	EnvInfo         `json:"env"`

	Content string `json:"content"`

	LikeCount    int `json:"like_count"`
	DislikeCount int `json:"dislike_count"`
	ReplyCount   int `json:"reply_count"`
}
