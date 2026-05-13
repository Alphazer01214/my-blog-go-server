package response

import (
	"time"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
)

type UserInfo struct {
	UserId       uint              `json:"user_id"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	Phone        string            `json:"phone"`
	Bio          string            `json:"bio"`
	Avatar       string            `json:"avatar"`
	Admin        bool              `json:"admin"`
	Role         constant.RoleType `json:"role"`
	PostCount    int64             `json:"post_count"`
	CommentCount int64             `json:"comment_count"`
}

type UserUpdate struct {
	Env      *entity.EnvInfo `json:"env"`
	UserInfo UserInfo        `json:"user_info"`
}

type UserQueryOne struct {
	UserInfo UserInfo `json:"user_info"`
}

type UserQueryMultiple struct {
	Users []UserInfo `json:"users"`
}
