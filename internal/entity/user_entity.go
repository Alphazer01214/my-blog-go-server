package entity

import (
	"time"

	"gorm.io/gorm"
)

type RoleType int

const (
	RoleEvil RoleType = iota
	RoleGuest
	RoleNormalUser
	RoleVIP
	RoleModerator
	RoleTakamatsuTomori
)

type User struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Username string `gorm:"size:64;uniqueIndex;not null" json:"username"`
	// Avatar: url
	Admin  bool     `json:"admin"`
	Role   RoleType `json:"role"`
	Banned bool     `json:"banned"`

	Password string `json:"-"` // 这表示忽略 Password 字段
}

// UserProfile 对外公开, main page
type UserProfile struct {
	UserId   uint   `json:"user_id" gorm:"primaryKey"`
	Username string `json:"username"`
	Email    string `gorm:"size:64" json:"email"`
	Phone    string `gorm:"size:64" json:"phone"`
	Bio      string `gorm:"type:text" json:"bio"`
	Avatar   string `gorm:"size:114" json:"avatar"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	FollowerCount        int `json:"follower_count" gorm:"default:0"`
	FollowingCount       int `json:"following_count" gorm:"default:0"`
	PostCount            int `json:"post_count" gorm:"default:0"`
	CommentCount         int `json:"comment_count" gorm:"default:0"`          // 该用户发的评论数
	ReceivedLikeCount    int `json:"received_like_count" gorm:"default:0"`    // 获赞
	ReceivedDislikeCount int `json:"received_dislike_count" gorm:"default:0"` // 踩
}

type UserSetting struct {
	UserId             uint `json:"user_id" gorm:"primaryKey"`
	PostPublic         bool `json:"post_public" gorm:"default:true"`
	CommentPublic      bool `json:"comment_public" gorm:"default:true"`
	FollowListPublic   bool `json:"follow_list_public" gorm:"default:true"`
	FollowerListPublic bool `json:"follower_list_public" gorm:"default:true"`
}

type UserFollow struct {
	FollowerId  uint `json:"follower_id" gorm:"primaryKey"`
	FollowingId uint `json:"following_id" gorm:"primaryKey"`
	IsMutual    bool `json:"is_mutual"`

	CreatedAt time.Time `json:"created_at"`
}

type TokenBlacklist struct {
	gorm.Model
	Token string `json:"token" gorm:"type:text"`
}
