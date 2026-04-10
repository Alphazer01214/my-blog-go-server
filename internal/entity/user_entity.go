package entity

import (
	"blog.alphazer01214.top/internal/constant"
	"gorm.io/gorm"
)

type User struct {
	// gorm.Model 已经包含了 ID, CreatedAt, UpdatedAt, DeletedAt
	gorm.Model
	// Id        int    `json:"id"`
	Username string `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email    string `gorm:"size:64" json:"email"`
	Phone    string `gorm:"size:64" json:"phone"`
	Bio      string `gorm:"type:text" json:"bio"`
	// Avatar: url
	Avatar string `gorm:"size:114" json:"avatar"`
	// CreatedAt time.Time
	// UpdatedAt time.Time
	Admin  bool              `json:"admin"`
	Role   constant.RoleType `json:"role"`
	Banned bool              `json:"banned"`

	//Agents datatypes.JSONArrayExpression `gorm:"type:json" json:"agents"`

	Password string `json:"-"` // 这表示忽略 Password 字段
}

type TokenBlacklist struct {
	gorm.Model
	Token string `json:"token" gorm:"type:text"`
}
