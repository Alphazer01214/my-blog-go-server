package api

import "blog.alphazer01214.top/internal/service"

type Apis struct {
	UserApi
	PostApi
	AiApi
}

var Api = new(Apis)

var (
	userService = service.Service.UserService
	postService = service.Service.PostService
	aiService   = service.Service.AIService
)
