package request

import "blog.alphazer01214.top/internal/entity"

type UserRegisterRequest struct {
	Env      *entity.EnvInfo `json:"env"`
	Username string          `json:"username" binding:"required"`
	// Password is not encrypted
	Password string `json:"password" binding:"required"`
}

type UserUpdateRequest struct {
	Env         *entity.EnvInfo `json:"env"`
	Id          uint            `json:"id"`
	NewUsername string          `json:"new_username"`
	NewEmail    string          `json:"new_email"`
	NewPhone    string          `json:"new_phone"`
	NewBio      string          `json:"new_bio"`
	NewAvatar   string          `json:"new_avatar"`
}

type UserUpdatePasswordRequest struct {
	Env         *entity.EnvInfo `json:"env"`
	Id          uint            `json:"id"`
	OldPassword string          `json:"old_password"`
	NewPassword string          `json:"new_password"`
}

type UserLoginRequest struct {
	Env *entity.EnvInfo `json:"env"`
	//Info     *entity.UserEnvInfo `json:"info"`
	//Id       uint                `json:"id"`
	Username string `json:"username" binding:"required"`
	//LoginType  constant.LoginType `json:"login_type"`
	//OS         string             `json:"os"`
	//IPv4       string             `json:"ipv4"`
	//IPv6       string             `json:"ipv6"`
	//DeviceInfo string             `json:"device_info"`
	Password string `json:"password" binding:"required"`
}

//type UserQueryRequest struct {
//	Id      uint   `json:"id"`
//	Keyword string `json:"keyword"`
//}
