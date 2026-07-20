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
	"blog.alphazer01214.top/internal/global"
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
	userId := cl.UserId
	ctx := c.Request.Context()

	session, err := aiService.GetChatSession(ctx, userId, chatId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	// 检查是否有进行中的流式响应
	streamStatus := aiService.GetStreamStatus(ctx, chatId)
	if streamStatus == "streaming" || streamStatus == "done" {
		// 从 redis 读取已有的 chunks
		chunks := aiService.GetStreamChunks(ctx, chatId, 0)
		if len(chunks) > 0 {
			// 将 chunks 转换为 ChatMessage 格式
			streamContent := ""
			for _, chunk := range chunks {
				if !chunk.IsError {
					streamContent += chunk.Content
				}
			}
			if streamContent != "" {
				// 添加到 session 的消息列表中
				streamMessage := entity.ChatMessage{
					Role:    "assistant",
					Content: streamContent,
					Time:    time.Now().Unix(),
					IsDone:  streamStatus == "done",
				}
				// 将 Messages 从 JSON 解析为 []ChatMessage
				var messages []entity.ChatMessage
				if session.Messages != nil {
					json.Unmarshal(session.Messages, &messages)
				}
				messages = append(messages, streamMessage)
				// 重新序列化为 JSON
				session.Messages, _ = json.Marshal(messages)
			}
		}
	}

	response.SuccessWithDetail(c, session, "query chat session success")
}

// OnlineStreamChat should handle chat in api layer
// 支持断线重连：如果 chat 还在 streaming，前端可以从 redis 继续接收数据
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

	if chatId == "" {
		chatId = utils.GenerateUUID()
	}

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

	// 检查是否有进行中的流式响应（优先检查，支持断线重连）
	streamStatus := aiService.GetStreamStatus(ctx, chatId)

	if streamStatus == "streaming" {
		// 有进行中的流式响应，先发送 history
		history := aiService.GetChatHistory(ctx, userId, chatId)
		rp.ChatId = chatId
		rp.Status = true
		if err := writeSSE(response.OnlineStreamChatResponse{
			ChatId:  chatId,
			AgentId: 0,
			UserId:  userId,
			History: history,
		}); err != nil {
			return
		}

		// 从 redis 订阅并推送 chunks
		subCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		streamCh := aiService.SubscribeStream(subCtx, chatId)
		for data := range streamCh {
			// 检查连接是否断开
			if c.Writer.Status() != http.StatusOK {
				return
			}

			chunk := data.Chunk
			rp.Content = chunk.Content
			rp.Status = true
			rp.Message = ""
			if chunk.IsError {
				rp.Status = false
				rp.Message = chunk.ErrorMsg
			}
			if err := writeSSE(rp); err != nil {
				return
			}
		}

		// 流式结束
		finalStatus := aiService.GetStreamStatus(ctx, chatId)
		if finalStatus == "error" {
			errMsg := aiService.GetStreamError(ctx, chatId)
			rp.Status = false
			rp.Message = errMsg
			_ = writeSSE(rp)
		} else {
			rp.Message = "done"
			rp.Status = true
			_ = writeSSE(rp)
		}
		return
	} else if streamStatus == "error" {
		// 流式已结束且有错误
		errMsg := aiService.GetStreamError(ctx, chatId)
		rp.ChatId = chatId
		rp.Status = false
		rp.Message = errMsg
		_ = writeSSE(rp)
		return
	} else if streamStatus == "done" {
		// 流式已完成，返回 history
		history := aiService.GetChatHistory(ctx, userId, chatId)
		rp.ChatId = chatId
		rp.Status = true
		_ = writeSSE(response.OnlineStreamChatResponse{
			ChatId:  chatId,
			AgentId: 0,
			UserId:  userId,
			History: history,
		})
		rp.Message = "done"
		_ = writeSSE(rp)
		return
	}

	// 没有进行中的流式响应，正常开始新的流式响应
	var req request.InvokeAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	// 这个非常 fuck up, 如果任何一个 prompt 为空都会导致错误， fuck
	req.SysPrompt = "你是一个在交易论坛的 assistant，请回答用户的以下问题："
	agentId := req.AgentId

	rp.ChatId = chatId
	rp.AgentId = agentId

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

// GetPresetAgents 获取预设 Agent 列表
func (ap *AiApi) GetPresetAgents(c *gin.Context) {
	presets := []gin.H{
		{
			"name":        "股票分析师",
			"description": "专业的 A 股分析师，提供技术分析、基本面分析、行情解读等服务",
			"tools":       []string{"tushare", "fundamental_analysis", "news_search", "web_search"},
		},
		{
			"name":        "论坛助手",
			"description": "帮助用户查找论坛帖子、回答问题、总结内容",
			"tools":       []string{"search_post", "web_search"},
		},
		{
			"name":        "量化分析师",
			"description": "专注于量化交易策略、技术指标计算和回测分析",
			"tools":       []string{"tushare", "web_search"},
		},
	}
	response.SuccessWithDetail(c, presets, "preset agents")
}

// CreatePresetAgent 从预设创建 Agent
func (ap *AiApi) CreatePresetAgent(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	var req struct {
		PresetName string `json:"preset_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	// 查找预设
	var sysPrompt string
	var tools []string
	switch req.PresetName {
	case "股票分析师":
		sysPrompt = "你是一位专业的 A 股市场分析师，具备技术分析、基本面分析、行情解读能力。"
		tools = []string{"tushare", "fundamental_analysis", "news_search", "web_search"}
	case "论坛助手":
		sysPrompt = "你是交易论坛的 AI 助手，帮助用户查找帖子、回答问题。"
		tools = []string{"search_post", "web_search"}
	case "量化分析师":
		sysPrompt = "你是一位量化交易分析师，专注于技术指标计算和策略分析。"
		tools = []string{"tushare", "web_search"}
	default:
		response.ErrorWithMsg(c, "unknown preset")
		return
	}

	agent := &entity.Agent{
		UserId:   cl.UserId,
		Name:     req.PresetName,
		Provider: "openai",
		BaseUrl:  global.GetConfig().LLM.BaseUrl,
		ApiKey:   global.GetConfig().LLM.ApiKey,
		ModelName: global.GetConfig().LLM.ModelName,
		Activate: true,
		Prompts: map[string]interface{}{
			"system": sysPrompt,
		},
		Tools: map[string]interface{}{
			"enabled": tools,
		},
	}

	if _, err := aiService.CreateAgent(c.Request.Context(), agent); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, agent, "preset agent created")
}
