package entity

import (
	"time"

	"blog.alphazer01214.top/internal/constant"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Post struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	EnvInfo   `json:"env"`

	Title    string         `gorm:"type:varchar(255)" json:"title"`
	Cover    string         `json:"cover"`
	UserId   uint           `json:"user_id"`
	Tags     datatypes.JSON `json:"tags"`
	Category string         `json:"category"`
	Keywords datatypes.JSON `json:"keywords"`
	Content  string         `gorm:"type:text" json:"content"`

	ViewCount     int `json:"view_count"`
	CommentCount  int `json:"comment_count"`
	LikeCount     int `json:"like_count"`
	DislikeCount  int `json:"dislike_count"`
	FavoriteCount int `json:"favorite_count"`
	ShareCount    int `json:"share_count"`

	Public        bool `json:"public"`
	ForbidComment bool `json:"forbid_comment"`
	ForbidShare   bool `json:"forbid_share"`
}

type Tag struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `json:"name" gorm:"uniqueIndex"`
	UsageCount int    `json:"usage_count" gorm:"default:0"`
}

type Comment struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserId uint `json:"user_id"`

	TargetType constant.TargetType `json:"target_type"`
	TargetId   uint                `json:"target_id"`
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

// 以下是 video 相关，包括上传等

type Video struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// VideoSrcUrl and VideoCoverUrl 都是上传后再获取的
	VideoSrcUrl   string `json:"video_src_url"`
	VideoCoverUrl string `json:"video_cover_url"`
	Title         string `gorm:"type:varchar(255)" json:"title"`
	Description   string `json:"description"`
	Duration      int    `json:"duration"` // 视频长度 秒
	Size          int64  `json:"size"`
	MimeType      string `gorm:"size:128" json:"mime_type"`

	UserId  uint `json:"user_id"`
	EnvInfo `json:"env"`

	Tags     datatypes.JSON `json:"tags"`
	Category string         `json:"category"`

	ViewCount     int `json:"view_count"`
	LikeCount     int `json:"like_count"`
	DislikeCount  int `json:"dislike_count"`
	CommentCount  int `json:"comment_count"`
	FavoriteCount int `json:"favorite_count"`
	ShareCount    int `json:"share_count"`

	Public        bool `json:"public"`
	ForbidComment bool `json:"forbid_comment"`
	ForbidShare   bool `json:"forbid_share"`

	Status string `gorm:"type:varchar(32);default:'published'" json:"status"`
}

// VideoUploadSession 不存数据库存 redis
type VideoUploadSession struct {
	//CreatedAt time.Time `json:"created_at"`
	//UpdatedAt time.Time `json:"updated_at"`
	ExpireAt time.Time `json:"expire_at"`

	UploadId    string                `json:"upload_id"`
	UserId      uint                  `json:"user_id"`
	VideoName   string                `json:"video_name"`
	VideoSize   int64                 `json:"video_size"`
	ChunkSize   int64                 `json:"chunk_size"`
	TotalChunks int64                 `json:"total_chunks"`
	Status      constant.UploadStatus `json:"status"`

	UploadedChunks []bool `json:"uploaded_chunks"`
}

type VideoUploadChunk struct {
	UploadId   string `json:"upload_id"`
	ChunkIndex int
	//ChunkSize  int
	ChunkHash string
}
