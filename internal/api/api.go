package api

import (
	"errors"

	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/service"
	"github.com/gin-gonic/gin"
)

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

func Authorize(c *gin.Context) (request.AccessClaims, error) {
	claims, exist := c.Get("claims")
	if !exist {
		return request.AccessClaims{}, errors.New("unauthorized")
	}

	if cl, ok := claims.(request.AccessClaims); ok {
		return cl, nil
	}
	return request.AccessClaims{}, errors.New("wrong claims format")
}
