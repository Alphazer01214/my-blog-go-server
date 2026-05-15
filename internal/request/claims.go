package request

import (
	"blog.alphazer01214.top/internal/entity"
	"github.com/golang-jwt/jwt/v5"
)

type BaseClaims struct {
	UserId   uint
	Username string
	RoleType entity.RoleType
}

type AccessClaims struct {
	BaseClaims
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	Id uint
	jwt.RegisteredClaims
}
