package api

import (
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
	"github.com/gin-gonic/gin"
)

type AiApi struct{}

func (ap *AiApi) Create(c *gin.Context) {
	ctx := c.Request.Context()
	userId := c.GetUint("user_id")
	var req request.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	agent := &entity.Agent{
		UserId:    userId,
		Name:      req.AgentName,
		Provider:  req.Provider,
		BaseUrl:   req.BaseUrl,
		ModelName: req.ModelName,
	}
	if _, err := aiService.CreateAgent(ctx, agent); err != nil {
		response.ErrorWithMsg(c, err.Error())
	}

	response.SuccessWithMsg(c, "create agent success")
}

func (ap *AiApi) Update(c *gin.Context) {
	var req request.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	ctx := c.Request.Context()
	userId := c.GetUint("user_id")
	agentId := c.GetUint("agent_id")
	newAgent := &entity.Agent{
		ApiKey:    req.ApiKey,
		Name:      req.AgentName,
		Provider:  req.Provider,
		BaseUrl:   req.BaseUrl,
		ModelName: req.ModelName,
		Activate:  req.Activate,
		Prompts:   req.Prompts,
		Memories:  req.Memories,
	}

	if _, err := aiService.UpdateAgent(ctx, userId, agentId, newAgent); err != nil {
		response.ErrorWithMsg(c, err.Error())
	} else {
		response.SuccessWithMsg(c, "update agent success")
	}
}

func (ap *AiApi) Invoke(c *gin.Context) {
	var req request.InvokeAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := c.GetUint("user_id")
	agentId := c.GetUint("agent_id")
	ctx := c.Request.Context()

	rp, err := aiService.InvokeAgent(ctx, userId, agentId, &req)
	if err != nil {
		response.ErrorWithDetail(c, rp, "error")
	}

	response.SuccessWithDetail(c, rp, "success")
}

// OnlineStreamChat should handle chat in api layer
func (ap *AiApi) OnlineStreamChat(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Connection", "keep-alive")
	c.Header("Cache-Control", "")

	chatId := c.Query("chat_id")
	userId := c.GetUint("user_id")
	agentId := c.GetUint("agent_id")
	if chatId == "" {
		chatId = utils.GenerateUUID()
	}
}
