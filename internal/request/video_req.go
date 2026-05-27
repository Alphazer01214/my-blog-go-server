package request

import (
	"blog.alphazer01214.top/internal/entity"
	"gorm.io/datatypes"
)

type VideoCreateRequest struct {
	Env         entity.EnvInfo `json:"env"`
	Title       string         `json:"title" binding:"required"`
	Description string         `json:"description" binding:"required"`
	// VideoSrcUrl 和 VideoCoverUrl 前端传入时需要保证空白，后续在 service 自动填充
	VideoSrcUrl   string         `json:"video_url"`
	VideoCoverUrl string         `json:"video_cover_url"`
	Category      string         `json:"category"`
	Tags          datatypes.JSON `json:"tags"`

	Public        bool `json:"public"`
	ForbidComment bool `json:"forbid_comment"`
	ForbidShare   bool `json:"forbid_share"`
}

type VideoUpdateRequest struct {
	VideoId       uint           `json:"video_id" binding:"required"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	VideoCoverUrl string         `json:"video_cover_url"`
	Category      string         `json:"category"`
	Tags          datatypes.JSON `json:"tags"`

	Public        bool `json:"public"`
	ForbidComment bool `json:"forbid_comment"`
	ForbidShare   bool `json:"forbid_share"`
}

type VideoActionRequest struct {
	VideoId uint `json:"video_id" binding:"required"`
}

type VideoShareRequest struct {
	VideoId      uint   `json:"video_id" binding:"required"`
	ShareTo      string `json:"share_to"`
	ShareMessage string `json:"share_message"`
}

type VideoSearchRequest struct {
	Keyword string `form:"keyword"`
	Tag     string `form:"tag"`
}

type VideoInitUploadRequest struct {
	UploadId  string `json:"upload_id" binding:"required"`
	VideoName string `json:"video_name" binding:"required"`
	VideoSize int64  `json:"video_size" binding:"required"`
	ChunkSize int64  `json:"chunk_size" binding:"required"`

	Env           entity.EnvInfo `json:"env"`
	Title         string         `json:"title" binding:"required"`
	Description   string         `json:"description" binding:"required"`
	VideoSrcUrl   string         `json:"video_url"`
	VideoCoverUrl string         `json:"video_cover_url"`
	Category      string         `json:"category"`
	Tags          datatypes.JSON `json:"tags"`
	Public        bool           `json:"public"`
	ForbidComment bool           `json:"forbid_comment"`
	ForbidShare   bool           `json:"forbid_share"`
}

type VideoInitUpload struct {
	UploadId  string `json:"upload_id" binding:"required"`
	VideoName string `json:"video_name" binding:"required"`
	VideoSize int64  `json:"video_size" binding:"required"`
	ChunkSize int64  `json:"chunk_size" binding:"required"`
}

type VideoChunkUpload struct {
	UploadId   string `json:"upload_id" binding:"required"`
	ChunkIndex int    `json:"chunk_index" binding:"required"`
	ChunkHash  string `json:"chunk_hash" binding:"required"`
}

type VideoMergeRequest struct {
	UploadId string `json:"upload_id" binding:"required"`
}
