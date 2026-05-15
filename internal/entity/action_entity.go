package entity

import (
	"time"

	"gorm.io/datatypes"
)

// action entity 用于记录点赞、踩、收藏、分享等操作

type TargetType int

const (
	TargetPost TargetType = iota + 1
	TargetComment
	TargetUser
	TargetTag
)

// ActionType 指定几个交互行为
type ActionType int

const (
	ActionLike ActionType = iota + 1
	ActionDislike
	ActionFavorite
	ActionShare
)

//type Like struct {
//	UserId     uint `json:"user_id"`
//	TargetId   uint `json:"target_id"`
//	TargetType `json:"target_type"`
//
//	CreatedAt time.Time `json:"created_at"`
//}
//
//type Dislike struct {
//	UserId     uint `json:"user_id"`
//	TargetId   uint `json:"target_id"`
//	TargetType `json:"target_type"`
//
//	CreatedAt time.Time `json:"created_at"`
//}
//
//// Favorite 限定只能收藏 post
//type Favorite struct {
//	UserId uint `json:"user_id"`
//	PostId uint `json:"post_id"`
//
//	CreatedAt time.Time `json:"created_at"`
//}
//
//type Share struct {
//	UserId     uint `json:"user_id"`
//	TargetId   uint `json:"target_id"`
//	TargetType `json:"target_type"`
//	ShareTo    string `json:"share_to"`
//
//	CreatedAt time.Time `json:"created_at"`
//}

type Action struct {
	UserId   uint `gorm:"primaryKey" json:"user_id"`
	TargetId uint `gorm:"primaryKey" json:"target_id"`

	ActType ActionType `gorm:"primaryKey;column:action_type" json:"action_type"`
	TgtType TargetType `gorm:"primaryKey;column:target_type" json:"target_type"` // ActionFavorite 限定只能收藏 post

	ExtraInfo datatypes.JSON `json:"extra_info"`
	CreatedAt time.Time      `json:"created_at"`
}

// ShareInfo 仅用于 ActionShare
type ShareInfo struct {
	ShareTo      string `json:"share_to"`
	ShareMessage string `json:"share_message"`
}
