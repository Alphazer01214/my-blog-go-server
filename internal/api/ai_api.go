package api

import (
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/gin-gonic/gin"
)

type AiApi struct{}

func (ap *AiApi) Create(c *gin.Context) {
	userId := c.GetUint("user_id")
	var req request.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
}
