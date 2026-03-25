package service

import (
	"strconv"
	"time"

	"blog.alphazer01214.top/internal/global"
)

type JWTService struct {
}

func (js *JWTService) SetRedisJwt(id uint, jwt string) error {
	expire := global.GetConfig().JWT.RefreshTokenExpireTime
	dur := time.Duration(expire) * time.Second
	idstr := strconv.Itoa(int(id))
	return global.GetRedis().Set(idstr, jwt, dur).Err()
}

func (js *JWTService) GetRedisJwt(id uint) (string, error) {
	idstr := strconv.Itoa(int(id))
	return global.GetRedis().Get(idstr).Result()
}

func (js *JWTService) JoinBlacklist(jwt string) error {
	global.JWTBlacklist[jwt] = true
	return nil
}

func (js *JWTService) IsBlacklisted(jwt string) bool {
	if _, exist := global.JWTBlacklist[jwt]; exist {
		return true
	}
	return false
}
