package request

import "blog.alphazer01214.top/internal/entity"

type PostCreateRequest struct {
	Env      entity.EnvInfo `json:"env"`
	Creator  entity.User    `json:"creator"`
	Title    string         `json:"title"`
	Cover    string         `json:"cover"`
	Category string         `json:"category"`
	Keywords string         `json:"keywords"`
	Content  string         `json:"content"`
	Public   bool           `json:"public"`
}
