package request

import (
	"blog.alphazer01214.top/internal/entity"
)

type StandardAiRequest struct {
	User      *entity.User
	Env       *entity.EnvInfo
	SysPrompt string
	UsrPrompt string
}
