package response

import (
	"time"

	"blog.alphazer01214.top/internal/entity"
)

type UserInfo struct {
	UserId               uint            `json:"user_id"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	Username             string          `json:"username"`
	Email                string          `json:"email"`
	Phone                string          `json:"phone"`
	Bio                  string          `json:"bio"`
	Avatar               string          `json:"avatar"`
	Admin                bool            `json:"admin"`
	Role                 entity.RoleType `json:"role"`
	Banned               bool            `json:"banned"`
	FollowerCount        int             `json:"follower_count"`
	FollowingCount       int             `json:"following_count"`
	PostCount            int             `json:"post_count"`
	CommentCount         int             `json:"comment_count"`
	ReceivedLikeCount    int             `json:"received_like_count"`
	ReceivedDislikeCount int             `json:"received_dislike_count"`
	IsFollowed           bool            `json:"is_followed"`
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

type FollowStatus struct {
	IsFollowing bool     `json:"is_following"`
	IsMutual    bool     `json:"is_mutual"`
	TargetUser  UserInfo `json:"target_user"`
}

type UserList struct {
	Items []UserInfo `json:"items"`
}

type FollowList struct {
	Items    []UserInfo `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Total    int64      `json:"total"`
}

type UserSettingResponse struct {
	PostPublic         bool `json:"post_public"`
	CommentPublic      bool `json:"comment_public"`
	FollowListPublic   bool `json:"follow_list_public"`
	FollowerListPublic bool `json:"follower_list_public"`
}
