package response

import (
	"time"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/datatypes"
)

type VideoDetail struct {
	ID            uint           `json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	VideoSrcUrl   string         `json:"video_src_url"`
	VideoCoverUrl string         `json:"video_cover_url"`
	Duration      int            `json:"duration"`
	Size          int64          `json:"size"`
	MimeType      string         `json:"mime_type"`
	UserId        uint           `json:"user_id"`
	Author        UserInfo       `json:"author,omitempty"`
	Category      string         `json:"category"`
	Tags          datatypes.JSON `json:"tags"`
	ViewCount     int            `json:"view_count"`
	LikeCount     int            `json:"like_count"`
	DislikeCount  int            `json:"dislike_count"`
	CommentCount  int            `json:"comment_count"`
	FavoriteCount int            `json:"favorite_count"`
	ShareCount    int            `json:"share_count"`
	Public        bool           `json:"public"`
	ForbidComment bool           `json:"forbid_comment"`
	ForbidShare   bool           `json:"forbid_share"`
	Status        string         `json:"status"`
	IsLiked       bool           `json:"is_liked"`
	IsDisliked    bool           `json:"is_disliked"`
	IsFavorited   bool           `json:"is_favorited"`
}

type VideoList struct {
	Items    []VideoDetail `json:"items"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Total    int64         `json:"total"`
}

type VideoUploadSessions struct {
	Sessions []entity.VideoUploadSession `json:"sessions"`
}
