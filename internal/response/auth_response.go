package response

import "blog.alphazer01214.top/internal/entity"

type Register struct {
	Env      *entity.EnvInfo `json:"env"`
	Username string          `json:"username"`
}

type Login struct {
	Env      *entity.EnvInfo `json:"env"`
	Token    *Token          `json:"token"`
	UserInfo UserInfo        `json:"user_info"`
}

type Token struct {
	AccessToken            string `json:"access_token"`
	AccessTokenExpireTime  int    `json:"access_token_expire_time"`
	RefreshToken           string `json:"refresh_token"`
	RefreshTokenExpireTime int    `json:"refresh_token_expire_time"`
}
