package response

import (
	"time"

	"gorm.io/datatypes"
)

type PostDetail struct {
	ID           uint           `json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Title        string         `json:"title"`
	Cover        string         `json:"cover"`
	UserId       uint           `json:"user_id"`
	Author       UserInfo       `json:"author,omitempty"`
	Tags         datatypes.JSON `json:"tags"`
	Category     string         `json:"category"`
	Keywords     datatypes.JSON `json:"keywords"`
	Content      string         `json:"content"`
	ViewCount    int            `json:"view_count"`
	CommentCount int            `json:"comment_count"`
	LikeCount    int            `json:"like_count"`
	DislikeCount int            `json:"dislike_count"`

	IsLiked    bool `json:"is_liked"`
	IsDisliked bool `json:"is_disliked"`
	Public     bool `json:"public"`
}

type PostList struct {
	Items    []PostDetail `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int64        `json:"total"`
}
