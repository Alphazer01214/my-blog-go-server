package response

import "blog.alphazer01214.top/internal/entity"

type UserInfo struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Bio      string `json:"bio"`
}

type UserUpdate struct {
	Env      *entity.EnvInfo `json:"env"`
	UserInfo interface{}     `json:"user_info"`
}

type UserQueryOne struct {
	UserInfo interface{} `json:"user_info"`
}

type UserQueryMultiple struct {
	Users interface{} `json:"users"`
}
