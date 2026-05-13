package api

import (
	"errors"
	"fmt"

	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/service"
	"github.com/gin-gonic/gin"
)

type Apis struct {
	UserApi
	PostApi
	CommentApi
	AiApi
}

var Api = new(Apis)

var (
	userService    = service.Service.UserService
	postService    = service.Service.PostService
	commentService = service.Service.CommentService
	aiService      = service.Service.AIService
)

func Authorize(c *gin.Context) (request.AccessClaims, error) {
	claims, exist := c.Get("claims")
	if !exist {
		fmt.Println("[api] no claims")
		return request.AccessClaims{}, errors.New("unauthorized")
	}

	if cl, ok := claims.(request.AccessClaims); ok {
		fmt.Printf("[api] claim: %v \n", cl)
		return cl, nil
	}

	return request.AccessClaims{}, errors.New("wrong claims format")
}

// GetUserId 从 context 中获取当前登录用户 ID，未登录返回 0
func GetUserId(c *gin.Context) uint {
	claims, exist := c.Get("claims")
	if !exist {
		return 0
	}
	if cl, ok := claims.(request.AccessClaims); ok {
		return cl.UserId
	}
	return 0
}
