package request

import (
	"blog.alphazer01214.top/internal/constant"
	"github.com/golang-jwt/jwt/v5"
)

type BaseClaims struct {
	Id       uint
	Username string
	RoleType constant.RoleType
}

type AccessClaims struct {
	BaseClaims
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	Id uint
	jwt.RegisteredClaims
}
