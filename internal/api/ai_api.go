package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
	userId := cl.UserId
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
		ApiKey:    req.ApiKey,
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
	userId := cl.UserId
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
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.UserId
	agentId := req.AgentId
	ctx := c.Request.Context()

	rp, err := aiService.InvokeAgent(ctx, userId, agentId, &req)
	if err != nil {
		response.ErrorWithDetail(c, rp, "error")
		return
	}

	response.SuccessWithDetail(c, rp, "success")
}

func (ap *AiApi) QueryAgents(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	agents, err := aiService.QueryAgentsByUser(c.Request.Context(), cl.UserId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, agents, "query agents success")
}

func (ap *AiApi) QueryAgentById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid id")
		return
	}
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	agent, err := aiService.QueryAgentById(c.Request.Context(), cl.UserId, uint(id))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, agent, "query agent success")
}

func (ap *AiApi) ListChats(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	page, pageSize := parsePagination(c)
	sessions, total, err := aiService.ListChatSessions(c.Request.Context(), cl.UserId, page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, gin.H{
		"items":     sessions,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}, "query chats success")
}

func (ap *AiApi) GetChatSession(c *gin.Context) {
	chatId := c.Param("chat_id")
	if chatId == "" {
		response.ErrorWithMsg(c, "missing chat_id")
		return
	}
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	session, err := aiService.GetChatSession(c.Request.Context(), cl.UserId, chatId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, session, "query chat session success")
}

// OnlineStreamChat should handle chat in api layer
func (ap *AiApi) OnlineStreamChat(c *gin.Context) {
	ctx := c.Request.Context()

	chatId := c.Query("chat_id")
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.UserId
	var req request.InvokeAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	agentId := req.AgentId

	if chatId == "" {
		chatId = utils.GenerateUUID()
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Connection", "keep-alive")
	c.Header("Cache-Control", "no-cache")

	rp := response.StandardAiResponse{
		AgentId: agentId,
		ChatId:  chatId,
		Status:  true,
		Content: "",
	}

	writeSSE := func(payload interface{}) error {
		j, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", j); err != nil {
			return err
		}
		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
		return nil
	}

	history := aiService.GetChatHistory(ctx, userId, chatId)
	if err := writeSSE(response.OnlineStreamChatResponse{
		ChatId:  chatId,
		AgentId: agentId,
		UserId:  userId,
		History: history,
	}); err != nil {
		return
	}

	pushStream := func(data string) error {
		if c.Writer.Status() != http.StatusOK {
			return errors.New("internet interrupted")
		}

		rp.Content = data
		return writeSSE(rp)
	}
	if err := aiService.StreamChat(ctx, userId, agentId, chatId, &req, pushStream); err != nil {
		rp.Status = false
		rp.Message = err.Error()
		_ = writeSSE(rp)
		return
	}
	rp.Message = "done"
	_ = writeSSE(rp)
}
