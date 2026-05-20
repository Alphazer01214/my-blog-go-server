package response

import (
	"time"

	"gorm.io/datatypes"
)

type PostDetail struct {
	ID            uint           `json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Title         string         `json:"title"`
	Cover         string         `json:"cover"`
	UserId        uint           `json:"user_id"`
	Author        UserInfo       `json:"author,omitempty"`
	Tags          datatypes.JSON `json:"tags"`
	CategoryId    uint           `json:"category_id"`
	CategoryName  string         `json:"category_name"`
	Keywords      datatypes.JSON `json:"keywords"`
	Content       string         `json:"content"`
	ViewCount     int            `json:"view_count"`
	CommentCount  int            `json:"comment_count"`
	LikeCount     int            `json:"like_count"`
	DislikeCount  int            `json:"dislike_count"`
	FavoriteCount int            `json:"favorite_count"`
	ShareCount    int            `json:"share_count"`

	Public        bool `json:"public"`
	ForbidComment bool `json:"forbid_comment"`
	ForbidShare   bool `json:"forbid_share"`

	IsLiked     bool `json:"is_liked"`
	IsDisliked  bool `json:"is_disliked"`
	IsFavorited bool `json:"is_favorited"`
}

type PostList struct {
	Items    []PostDetail `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int64        `json:"total"`
}
