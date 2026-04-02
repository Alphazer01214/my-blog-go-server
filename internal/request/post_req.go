package request

import "blog.alphazer01214.top/internal/entity"

type PostCreateRequest struct {
	Env      entity.EnvInfo `json:"env"`
	UserId   uint           `json:"user_id"`
	Title    string         `json:"title"`
	Cover    string         `json:"cover"`
	Category string         `json:"category"`
	Keywords string         `json:"keywords"`
	Content  string         `json:"content"`
	Public   bool           `json:"public"`
}

type PostUpdateRequest struct {
	Id       uint           `json:"post_id"`
	Env      entity.EnvInfo `json:"env"`
	Title    string         `json:"title"`
	Cover    string         `json:"cover"`
	Category string         `json:"category"`
	Keywords string         `json:"keywords"`
	Content  string         `json:"content"`
	Public   bool           `json:"public"`
}
