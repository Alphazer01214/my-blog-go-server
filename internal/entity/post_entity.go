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
	Title     string         `gorm:"type:varchar(255)" json:"title"`
	Cover     string         `json:"cover"`
	UserId    uint           `json:"user_id"`
	Tags      datatypes.JSON `json:"tags"`
	Category  string         `json:"category"`
	Keywords  datatypes.JSON `json:"keywords"`
	Content   string         `gorm:"type:text" json:"content"`
	//ContentNodes datatypes.JSONArrayExpression

	ViewCount    int `json:"view_count"`
	CommentCount int `json:"comment_count"`
	LikeCount    int `json:"like_count"`
	DislikeCount int `json:"dislike_count"`

	Public bool `json:"public"`
}

type Like struct {
	UserId uint      `json:"user_id" gorm:"uniqueKey"`
	LikeAt time.Time `json:"like_at" gorm:"autoCreateTime"`
}

type Dislike struct {
	UserId    uint      `json:"user_id" gorm:"uniqueKey"`
	DislikeAt time.Time `json:"dislike_at" gorm:"autoCreateTime"`
}

type PostLike struct {
	Like
	PostId uint `json:"post_id" gorm:"uniqueKey"`
}

type PostDislike struct {
	Dislike
	PostId uint `json:"post_id" gorm:"uniqueKey"`
}
