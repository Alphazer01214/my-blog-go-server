package utils

import (
	"errors"
	"net"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func SetRefreshToken(c *gin.Context, token string, age int) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}

	setCookies(c, constant.CookieRefreshToken, token, age, host)
}

func RemoveRefreshToken(c *gin.Context) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}

	setCookies(c, constant.CookieRefreshToken, "", -1, host)
}

func GetRefreshToken(c *gin.Context) string {
	token := c.Request.Header.Get(constant.CookieRefreshToken)
	return token
}

func GetAccessToken(c *gin.Context) string {
	token := c.Request.Header.Get(constant.CookieAccessToken)
	return token
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
