package entity

import (
	"time"

	"gorm.io/gorm"
)

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

	Content  string `json:"content"`
	Likes    uint   `json:"likes"`
	Dislikes uint   `json:"dislikes"`
	Replies  uint   `json:"replies"`
}

type CommentLike struct {
	Like
	CommentId uint `json:"comment_id" gorm:"uniqueKey"`
}

type CommentDislike struct {
	Dislike
	CommentId uint `json:"comment_id" gorm:"uniqueKey"`
}
