package request

import "blog.alphazer01214.top/internal/entity"

type CommentCreateRequest struct {
	Env             entity.EnvInfo `json:"env"`
	PostId          uint           `json:"post_id" binding:"required"`
	Content         string         `json:"content" binding:"required"`
	RootCommentId   uint           `json:"root_comment_id"`
	ParentCommentId uint           `json:"parent_comment_id"`
}

type CommentActionRequest struct {
	CommentId uint `json:"comment_id" binding:"required"`
}
