package response

import (
	"time"
)

type CommentList struct {
	Items    []Comment `json:"items"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Total    int64     `json:"total"`
}

// Comment 用于 response 的评论结构体，存在一个递归嵌套，放所有 root id 或 parent id 是该评论的评论
type Comment struct {
	CommentId       uint      `json:"comment_id"`
	UserId          uint      `json:"user_id"`
	PostId          uint      `json:"post_id"`
	RootCommentId   uint      `json:"root_comment_id"`
	ParentCommentId uint      `json:"parent_comment_id"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Author UserInfo `json:"author,omitempty"`

	Likes    uint `json:"likes"`
	Dislikes uint `json:"dislikes"`
	Replies  uint `json:"replies"`

	IsLiked    bool `json:"is_liked"`
	IsDisliked bool `json:"is_disliked"`

	ReplyComments []Comment `json:"reply_comments"`
}
