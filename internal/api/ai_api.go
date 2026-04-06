package api

import (
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/gin-gonic/gin"
)

type AiApi struct{}

func (ap *AiApi) Create(c *gin.Context) {
	ctx := c.Request.Context()
	userId := c.GetUint("user_id")
	var req *request.CreateAgentRequest
	if err := c.ShouldBindJSON(req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	if _, err := aiService.CreateAgent(ctx, userId, req); err != nil {
		response.ErrorWithMsg(c, err.Error())
	}

	response.SuccessWithMsg(c, "create agent success")
}

func (ap *AiApi) Update(c *gin.Context) {
	ctx := c.Request.Context()
	userId := c.GetUint("user_id")

}
