package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
	"github.com/gin-gonic/gin"
)

type AiApi struct{}

func (ap *AiApi) Create(c *gin.Context) {
	ctx := c.Request.Context()
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.Id
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
		return
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
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.Id
	agentId := req.AgentId
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
		return
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
	agentId := req.AgentId
	ctx := c.Request.Context()

	rp, err := aiService.InvokeAgent(ctx, userId, agentId, &req)
	if err != nil {
		response.ErrorWithDetail(c, rp, "error")
		return
	}

	response.SuccessWithDetail(c, rp, "success")
}

// OnlineStreamChat should handle chat in api layer
func (ap *AiApi) OnlineStreamChat(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Connection", "keep-alive")
	c.Header("Cache-Control", "")
	ctx := c.Request.Context()

	chatId := c.Query("chat_id")
	//userId := c.GetUint("user_id")
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.Id
	var req request.InvokeAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	agentId := req.AgentId

	if chatId == "" {
		chatId = utils.GenerateUUID()
	}

	rp := response.StandardAiResponse{
		AgentId: agentId,
		Status:  true,
		Content: "",
	}

	pushStream := func(data string) error {
		if c.Writer.Status() != http.StatusOK {
			//response.ErrorWithMsg(c, "internet interrupted")
			return errors.New("internet interrupted")
		}

		rp.Content += data
		j, _ := json.Marshal(rp)
		_, err := fmt.Fprintf(c.Writer, "data: %s\n\n", j)
		if err != nil {
			return err
		}

		return nil
	}
	if err := aiService.
		StreamChat(ctx, userId, agentId, chatId, &req, pushStream); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, response.OnlineStreamChatResponse{
		ChatId:  chatId,
		AgentId: agentId,
		UserId:  userId,
	}, "success")
}
