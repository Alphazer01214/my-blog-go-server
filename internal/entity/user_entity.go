package entity

import (
	"blog.alphazer01214.top/internal/constant"
	"gorm.io/gorm"
)

type User struct {
	// gorm.Model 已经包含了 ID, CreatedAt, UpdatedAt, DeletedAt
	gorm.Model
	//Id        int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Bio      string `json:"bio"`
	// Avatar: url
	Avatar string `json:"avatar"`
	//CreatedAt time.Time
	//UpdatedAt time.Time
	Admin  bool              `json:"admin"`
	Role   constant.RoleType `json:"role"`
	Banned bool              `json:"banned"`

	Password string `json:"-"` // 这表示忽略 Password 字段
}

type JWTBlacklist struct {
	gorm.Model
	Jwt string `json:"jwt" gorm:"type:text"`
}
