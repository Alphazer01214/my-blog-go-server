package response

import "blog.alphazer01214.top/internal/entity"

type Register struct {
	Env      *entity.EnvInfo `json:"env"`
	Username string          `json:"username"`
}

// Login 登录响应结构
type Login struct {
	Env *entity.EnvInfo `json:"env"`
	// Token access token
	Token *Token `json:"token"`
	// UserInfo is entity.User without password
	UserInfo interface{} `json:"user_info"`
}

type Token struct {
	AccessToken            string `json:"access_token"`
	AccessTokenExpireTime  int    `json:"access_token_expire_time"`
	RefreshToken           string `json:"refresh_token"`
	RefreshTokenExpireTime int    `json:"refresh_token_expire_time"`
}
