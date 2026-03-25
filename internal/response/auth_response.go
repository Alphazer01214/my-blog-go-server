package response

import "blog.alphazer01214.top/internal/entity"

type Register struct {
	Env      *entity.EnvInfo `json:"env"`
	Username string          `json:"username"`
}

// Login 登录响应结构
type Login struct {
	Env       *entity.EnvInfo `json:"env"`
	Token     string          `json:"token"`
	ExpiresIn int64           `json:"expires_in"` // 过期时间（秒）
	UserInfo  interface{}     `json:"user_info"`  // 用户信息
}
