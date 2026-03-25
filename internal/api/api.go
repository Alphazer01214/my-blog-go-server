package api

import "blog.alphazer01214.top/internal/service"

type Apis struct {
	UserApi
	PostApi
}

var Api = new(Apis)

var userService = service.Service.UserService
var postService = service.Service.PostService
var jwtService = service.Service.JWTService
