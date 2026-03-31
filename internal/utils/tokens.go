package utils

import (
	"errors"
	"net"
	"strconv"
	"time"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
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

func GenerateRefreshClaims(base request.BaseClaims) request.RefreshClaims {
	expireAt := time.Now().Add(
		time.Duration(global.GetConfig().JWT.RefreshTokenExpireTime) * time.Second,
	)

	return request.RefreshClaims{
		Id: base.Id,
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

func SetRefreshTokenCookie(c *gin.Context, token string, age int) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}

	setCookies(c, constant.CookieRefreshToken, token, age, host)
}

func RemoveRefreshTokenCookie(c *gin.Context) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}

	setCookies(c, constant.CookieRefreshToken, "", -1, host)
}

func GetRefreshTokenCookie(c *gin.Context) string {
	token := c.Request.Header.Get(constant.CookieRefreshToken)
	return token
}

func GetAccessTokenCookie(c *gin.Context) string {
	token := c.Request.Header.Get(constant.CookieAccessToken)
	return token
}

func SetRefreshTokenRedis(id uint, token string) error {
	expire := global.GetConfig().JWT.RefreshTokenExpireTime
	dur := time.Duration(expire) * time.Second
	idstr := strconv.Itoa(int(id))
	return global.GetRedis().Set(idstr, token, dur).Err()
}

func GetRefreshTokenRedis(id uint) (string, error) {
	idstr := strconv.Itoa(int(id))
	return global.GetRedis().Get(idstr).Result()
}

func TokenJoinBlacklist(token string) error {
	return global.GetDB().Create(&entity.TokenBlacklist{
		Token: token,
	}).Error
}

func IsTokenBlacklisted(token string) (bool, error) {
	if err := global.GetDB().Where("token = ?", token).First(&entity.TokenBlacklist{}).Error; err == nil {
		return true, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else {
		return false, err
	}
}

// ParseAccessToken 接收token返回claims
func ParseAccessToken(token string) (*request.AccessClaims, error) {
	claims, err := parseToken(token, &request.AccessClaims{}, global.GetConfig().JWT.AccessTokenSecret)
	if err != nil {
		return nil, err
	}
	if accessClaims, ok := claims.(*request.AccessClaims); ok {
		return accessClaims, nil
	}
	return nil, errors.New("invalid access claims")
}

func setCookies(c *gin.Context, name string, value string, age int, host string) {
	if net.ParseIP(host) == nil {
		c.SetCookie(name, value, age, "/", "", false, true)
	} else {
		c.SetCookie(name, value, age, "/", host, false, false)
	}
}

func parseToken(input string, claims jwt.Claims, secret interface{}) (interface{}, error) {
	token, err := jwt.ParseWithClaims(input, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	// secret 可以判断 token 和 claims 是否相对应
	if err != nil {
		return nil, err
	}
	if token.Valid {
		return token.Claims, nil
	}
	return nil, errors.New("invalid token")
}
