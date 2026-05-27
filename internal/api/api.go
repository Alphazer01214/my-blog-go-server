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
	MarketApi
	FileApi
	TomoriApi
	VideoApi
}

func NewApis(svc *service.Services) *Apis {
	return &Apis{
		UserApi:    UserApi{userService: svc.UserService},
		PostApi:    PostApi{postService: svc.PostService, aiService: svc.AIService, userService: svc.UserService},
		CommentApi: CommentApi{commentService: svc.CommentService},
		AiApi:      AiApi{aiService: svc.AIService},
		MarketApi:  MarketApi{marketService: svc.MarketService},
		FileApi:    FileApi{fileService: svc.FileService},
		TomoriApi:  TomoriApi{tomoriService: svc.TomoriService},
		VideoApi:   VideoApi{videoService: svc.VideoService, userService: svc.UserService},
	}
}

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

func GetUserId(c *gin.Context) uint {
	claims, exist := c.Get("claims")
	if !exist {
		return 0
	}
	if cl, ok := claims.(request.AccessClaims); ok {
		fmt.Printf("[Auth] current user id: %v\n", cl.UserId)
		return cl.UserId
	}
	fmt.Println("[Auth] no claims")
	return 0
}
