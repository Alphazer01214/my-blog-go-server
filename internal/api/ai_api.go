package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

func (ap *AiApi) Delete(c *gin.Context) {
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
	if err := aiService.DeleteAgent(c.Request.Context(), cl.UserId, uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "delete agent success")
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
	// 巨坑：如果使用 c.Request.Context()，那么前端 refresh 就会取消这个 context 导致后端 stream 直接退出
	//ctx := c.Request.Context()
	chatId := c.Query("chat_id")

	// standard response
	rp := response.StandardAiResponse{
		AgentId: 0,
		ChatId:  chatId,
		Status:  false,
		Content: "",
	}

	ctx := context.Background()
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
	// 这个非常 fuck up, 如果任何一个 prompt 为空都会导致错误， fuck
	req.SysPrompt = "你是一个在交易论坛的 assistant，请回答用户的以下问题："
	agentId := req.AgentId

	if chatId == "" {
		chatId = utils.GenerateUUID()
	}
	rp.ChatId = chatId
	rp.AgentId = agentId

	// set head when start streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Connection", "keep-alive")
	c.Header("Cache-Control", "no-cache")
	// The Flusher interface is implemented by ResponseWriters that allow an HTTP handler to flush buffered data to the client.
	// writeSSE: push sse event
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

	// 检查是否有进行中的流式响应（前端重连场景）
	streamStatus := aiService.GetStreamStatus(ctx, chatId)
	if streamStatus == "streaming" {
		// 有进行中的流式响应，从 redis 恢复已有 chunks
		sentChunks := 0
		for {
			// 检查连接是否断开
			if c.Writer.Status() != http.StatusOK {
				return
			}

			chunks := aiService.GetStreamChunks(ctx, chatId)
			// 推送新增的 chunks
			for i := sentChunks; i < len(chunks); i++ {
				chunk := chunks[i]
				rp.Content = chunk.Content
				rp.Status = true
				if chunk.IsError {
					rp.Status = false
					rp.Message = chunk.ErrorMsg
				}
				if err := writeSSE(rp); err != nil {
					return
				}
			}
			sentChunks = len(chunks)

			// 检查流式是否完成
			currentStatus := aiService.GetStreamStatus(ctx, chatId)
			if currentStatus == "done" {
				// 推送剩余的 chunks
				chunks = aiService.GetStreamChunks(ctx, chatId)
				for i := sentChunks; i < len(chunks); i++ {
					chunk := chunks[i]
					rp.Content = chunk.Content
					rp.Status = true
					if err := writeSSE(rp); err != nil {
						return
					}
				}
				rp.Message = "done"
				rp.Status = true
				_ = writeSSE(rp)
				return
			} else if currentStatus == "error" {
				// 推送错误信息
				errMsg := aiService.GetStreamError(ctx, chatId)
				rp.Status = false
				rp.Message = errMsg
				_ = writeSSE(rp)
				return
			}

			// 轮询间隔
			time.Sleep(100 * time.Millisecond)
		}
	} else if streamStatus == "error" {
		// 流式已结束且有错误
		errMsg := aiService.GetStreamError(ctx, chatId)
		rp.Status = false
		rp.Message = errMsg
		_ = writeSSE(rp)
		return
	}

	// 正常开始新的流式响应
	pushStream := func(data string) error {
		if c.Writer.Status() != http.StatusOK {
			return errors.New("internet interrupted")
		}
		rp.Content = data
		rp.Status = true
		return writeSSE(rp)
	}

	if err := aiService.ToolCallingStreamChat(ctx, userId, chatId, &req, pushStream); err != nil {
		rp.Status = false
		rp.Message = err.Error()
		_ = writeSSE(rp)
		return
	}
	rp.Message = "done"
	_ = writeSSE(rp)
}

func (ap *AiApi) DeleteChatSession(c *gin.Context) {
	var req request.DeleteChatSessionRequest
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
	chatId := req.ChatId

	if err := aiService.DeleteChatSession(c.Request.Context(), userId, chatId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "delete chat session success")
}

func (ap *AiApi) UpdateChatSession(c *gin.Context) {
	var req request.UpdateChatSessionRequest
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
	chatId := req.ChatId
	title := req.Title

	if err := aiService.UpdateChatSession(c.Request.Context(), userId, chatId, title); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "update chat session success")
}
