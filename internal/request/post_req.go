package request

import (
	"blog.alphazer01214.top/internal/entity"

	"gorm.io/datatypes"
)

type PostCreateRequest struct {
	Env        entity.EnvInfo `json:"env"`
	Title      string         `json:"title"`
	Cover      string         `json:"cover"`
	CategoryId uint           `json:"category_id"`
	Tags       datatypes.JSON `json:"tags"`
	Keywords   datatypes.JSON `json:"keywords"`
	Content    string         `json:"content"`
	Public     bool           `json:"public"`
}

type PostUpdateRequest struct {
	Id         uint           `json:"post_id"`
	Env        entity.EnvInfo `json:"env"`
	Title      string         `json:"title"`
	Cover      string         `json:"cover"`
	CategoryId uint           `json:"category_id"`
	Tags       datatypes.JSON `json:"tags"`
	Keywords   datatypes.JSON `json:"keywords"`
	Content    string         `json:"content"`
	Public     bool           `json:"public"`
}

type PostActionRequest struct {
	PostId uint `json:"post_id" binding:"required"`
}

type PostShareRequest struct {
	PostId       uint   `json:"post_id" binding:"required"`
	ShareTo      string `json:"share_to"`
	ShareMessage string `json:"share_message"`
}
