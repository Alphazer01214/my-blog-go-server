package entity

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	EnvInfo  `json:"env"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	UserId   uint   `json:"user_id"`
	Category string `json:"category"`
	Keywords string `json:"keywords"`
	Content  string `json:"content"`

	ViewCount    int `json:"view_count"`
	CommentCount int `json:"comment_count"`
	LikeCount    int `json:"like_count"`
	DislikeCount int `json:"dislike_count"`

	Public bool `json:"public"`
}

type PostComment struct {
	UserId   uint `json:"user_id"`
	EnvInfo  `json:"env"`
	Location string `json:"location"`
	Content  string `json:"content"`
}

type Like struct {
	UserId uint `json:"user_id"`
	LikeAt time.Time
}

// Dislike should not in any response
type Dislike struct {
	UserId    uint `json:"user_id"`
	DislikeAt time.Time
}
