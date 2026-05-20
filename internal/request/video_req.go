package request

import (
	"blog.alphazer01214.top/internal/entity"
	"gorm.io/datatypes"
)

type VideoCreateRequest struct {
	Env           entity.EnvInfo `json:"env"`
	Title         string         `json:"title" binding:"required"`
	Description   string         `json:"description" binding:"required"`
	VideoSrcUrl   string         `json:"video_url" binding:"required"`
	VideoCoverUrl string         `json:"video_cover_url" binding:"required"`
	CategoryId    uint           `json:"category_id"`
	Tags          datatypes.JSON `json:"tags"`

	Public        bool `json:"public"`
	ForbidComment bool `json:"forbid_comment"`
	ForbidShare   bool `json:"forbid_share"`
}
