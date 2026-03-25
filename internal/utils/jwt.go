package utils

import (
	"time"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateBaseClaims(id uint, username string, role constant.RoleType) request.BaseClaims {
	return request.BaseClaims{
		Id:       id,
		Username: username,
		RoleType: role,
	}
}

func GenerateAccessClaims(base request.BaseClaims) request.AccessClaims {
	expireAt := time.Now().Add(
		time.Duration(global.GetConfig().JWT.AccessTokenExpireTime) * time.Second,
	)

	return request.AccessClaims{
		BaseClaims: base,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    global.GetConfig().JWT.Issuer,
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}
}

func GenerateRefreshClaims(id uint) request.RefreshClaims {
	expireAt := time.Now().Add(
		time.Duration(global.GetConfig().JWT.AccessTokenExpireTime) * time.Second,
	)

	return request.RefreshClaims{
		Id: id,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    global.GetConfig().JWT.Issuer,
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}
}

func GenerateAccessTokenFromClaims(claims request.AccessClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(global.GetConfig().JWT.AccessTokenSecret))
	return tokenString
}

func GenerateRefreshTokenFromClaims(claims request.RefreshClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(global.GetConfig().JWT.RefreshTokenSecret))
	return tokenString
}
