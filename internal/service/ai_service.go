package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	tomori_tool "blog.alphazer01214.top/internal/agent/tool"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/ollama/api"
	"gorm.io/gorm"
)

type AIService struct {
	saveMu sync.Map // map[string]*sync.Mutex, keyed by chatId
}

func (ai *AIService) CreateAgent(ctx context.Context, agent *entity.Agent) (*entity.Agent, error) {
	if err := global.GetDB().WithContext(ctx).Create(agent).Error; err != nil {
		return nil, err
	}
	return agent, nil
}

func (ai *AIService) UpdateAgent(ctx context.Context, userId uint, agentId uint, newAgent *entity.Agent) (*entity.Agent, error) {
	agent, err := ai.queryAgentById(ctx, agentId)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, errors.New("agent not found")
	}
	if agent.UserId != userId {
		return nil, errors.New("unauthorized")
	}
	agent.Name = newAgent.Name
	agent.BaseUrl = newAgent.BaseUrl
	agent.ApiKey = newAgent.ApiKey
	agent.ModelName = newAgent.ModelName
	agent.Provider = newAgent.Provider
	agent.Activate = newAgent.Activate
	agent.Prompts = newAgent.Prompts
	agent.Memories = newAgent.Memories
	if err := global.GetDB().WithContext(ctx).Save(agent).Error; err != nil {
		return nil, err
	}
	return agent, nil
}

//func (ai *AIService) QueryAgentsByUser(ctx context.Context, userId uint) ([]response.AgentInfo, error) {
//	var agents []entity.Agent
//	var info []response.AgentInfo
//	if err := global.GetDB().WithContext(ctx).Where("user_id = ?", userId).Order("id desc").Find(&agents).Error; err != nil {
//		return nil, err
//	}
//
//	for _, agent := range agents {
//		info = append(info, response.AgentInfo{
//			AgentId: agent.ID,
//			UserId: agent.UserId,
//			Name:    agent.Name,
//			ModelName: agent.ModelName,
//			Provider:  agent.Provider,
//			Activate:  agent.Activate,
//			CreatedAt: agent.CreatedAt,
//			UpdatedAt: agent.UpdatedAt,
//			Temperature: agent.Temperature,
//			Thinking: agent.Thinking,
//			Prompts: agent.Prompts,
//			Memories: agent.Memories,
//		})
//	}
//	return info, nil
//}

func (ai *AIService) QueryAgentsByUser(ctx context.Context, userId uint) ([]entity.Agent, error) {
	var agents []entity.Agent
	if err := global.GetDB().WithContext(ctx).Where("user_id = ?", userId).Order("id desc").Find(&agents).Error; err != nil {
		return nil, err
	}

	return agents, nil
}

func (ai *AIService) QueryAgentById(ctx context.Context, userId uint, agentId uint) (*entity.Agent, error) {
	agent, err := ai.queryAgentById(ctx, agentId)
	if err != nil {
		return nil, err
	}
	if agent.UserId != userId {
		return nil, errors.New("unauthorized")
	}
	return agent, nil
}

func (ai *AIService) DeleteAgent(ctx context.Context, userId uint, agentId uint) error {
	agent, err := ai.queryAgentById(ctx, agentId)
	if err != nil {
		return err
	}
	if agent == nil {
		return errors.New("agent not found")
	}
	if agent.UserId != userId {
		return errors.New("unauthorized")
	}
	return global.GetDB().WithContext(ctx).Delete(agent).Error
}
func (ai *AIService) GetChatHistory(ctx context.Context, userId uint, chatId string) []*schema.Message {
	session, err := ai.loadChat(ctx, userId, chatId)
	if err != nil || session == nil {
		return []*schema.Message{}
	}
	var msgs []*schema.Message
	for _, m := range session.ChatMessages {
		role := schema.User
		if m.Role == "model" {
			role = schema.Assistant
		}
		msgs = append(msgs, &schema.Message{
			Role:    role,
			Content: m.Content,
		})
	}
	return msgs
}

func (ai *AIService) ListChatSessions(ctx context.Context, userId uint, page, pageSize int) ([]entity.ChatSession, int64, error) {
	var sessions []entity.ChatSession
	var total int64
	db := global.GetDB().WithContext(ctx).Model(&entity.ChatSession{}).Where("user_id = ?", userId)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("updated_at desc").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

func (ai *AIService) GetChatSession(ctx context.Context, userId uint, chatId string) (*entity.ChatSession, error) {
	var chatSession entity.ChatSession
	err := global.GetDB().WithContext(ctx).Where("uuid = ? AND user_id = ?", chatId, userId).First(&chatSession).Error
	if err != nil {
		return nil, err
	}
	return &chatSession, nil
}

func (ai *AIService) InvokeAgent(ctx context.Context, userId uint, agentId uint, req *request.InvokeAgentRequest) (*response.StandardAiResponse, error) {
	agent, err := ai.queryAgentById(ctx, agentId)
	if err != nil {
		return &response.StandardAiResponse{
			AgentId: agentId,
			Status:  false,
			Message: err.Error(),
		}, err
	}
	if userId != agent.UserId {
		return &response.StandardAiResponse{
			AgentId: agentId,
			Status:  false,
			Message: "unauthorized: the agent's owner is not you",
		}, errors.New("unauthorized")
	}
	usrPrompt := req.UsrPrompt
	if usrPrompt == "" {
		usrPrompt = " "
	}
	message := []*schema.Message{
		schema.UserMessage(usrPrompt),
	}
	if req.SysPrompt != "" {
		message = append(message, schema.SystemMessage(req.SysPrompt))
	}

	chatModel, err := ai.getEinoChatModel(ctx, agent)
	if err != nil {
		return &response.StandardAiResponse{
			AgentId: agentId,
			Status:  false,
			Message: err.Error(),
		}, err
	}

	content, err := chatModel.Generate(ctx, message)
	if err != nil {
		return &response.StandardAiResponse{
			AgentId: agentId,
			Status:  false,
			Message: err.Error(),
		}, err
	}
	return &response.StandardAiResponse{
		AgentId: agentId,
		Status:  true,
		Content: content.Content,
		Message: "success",
	}, nil
}

// StreamChat receive a session id
func (ai *AIService) StreamChat(ctx context.Context, userId uint, agentId uint, chatId string, req *request.InvokeAgentRequest, callback func(string) error) error {
	fmt.Printf("[service] stream chat: userId: %d, agentId: %d, chatId: %s\n", userId, agentId, chatId)
	agent, err := ai.queryAgentById(ctx, agentId)
	if err != nil {
		return err
	}
	if agent.UserId != userId {
		return errors.New("service unauthorized: the agent's owner is not you")
	}
	chatModel, err := ai.getEinoChatModel(ctx, agent)
	if err != nil {
		return err
	}
	session, err := ai.loadChat(ctx, userId, chatId)
	fmt.Printf("history: %v\n", session)
	if err != nil {
		return err
	}
	ask := entity.ChatMessage{
		Role:    "user",
		Content: req.UsrPrompt,
		Time:    time.Now().Unix(),
	}
	session.ChatMessages = append(session.ChatMessages, ask)
	session.UserId = userId
	session.AgentId = agentId
	session.UpdateAt = time.Now().Unix()
	if session.CreateAt == 0 {
		session.CreateAt = time.Now().Unix()
	}
	ai.saveSessionAsync(userId, chatId, session)

	prompt, err := ai.buildPrompt(ctx, session, ask)
	if err != nil {
		return err
	}
	stream, err := chatModel.Stream(ctx, prompt)
	if err != nil {
		return err
	}
	defer stream.Close()
	res := ""
	for {
		if !agent.Activate {
			// 没激活，把这个错误信息当作消息不保存地发出去
			callback("agent not activated!")
			callback("")
			return nil
		}
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// stream 到结尾，跳出
			break
		}
		cm := entity.ChatMessage{
			Role:     "assistant",
			Content:  res,
			Time:     time.Now().Unix(),
			IsDone:   false,
			IsError:  false,
			ErrorMsg: "",
		}
		if err != nil {
			if res != "" {
				// 有未知错误，且有内容了，先保存再退出
				//session.ChatMessages = append(session.ChatMessages, entity.ChatMessage{
				//	Role: "model", Content: res, Time: time.Now().Unix(),
				//})
				//session.UpdateAt = time.Now().Unix()
				cm.IsError = true
				cm.ErrorMsg = err.Error()
				ai.cacheStreamChat(ctx, chatId, &cm)
				session.ChatMessages = append(session.ChatMessages, cm)
				ai.saveSessionAsync(userId, chatId, session)
				ai.removeChatFromRedis(ctx, chatId)
			}
			return err
		}
		if chunk.Content == "" {
			continue
		}
		res += chunk.Content
		cm.Content = res
		if err := callback(chunk.Content); err != nil {
			// 有未知错误，但是前端的，不管前端继续写入 redis，不过先保存
			//ai.saveSessionAsync()
		}
		ai.cacheStreamChat(ctx, chatId, &cm)
	}

	resChatMessage := entity.ChatMessage{
		Role:    "model",
		Content: res,
		Time:    time.Now().Unix(),
		IsDone:  true,
		IsError: false,
	}

	ai.cacheStreamChat(ctx, chatId, &resChatMessage)

	session.ChatMessages = append(session.ChatMessages, resChatMessage)
	session.UpdateAt = time.Now().Unix()
	ai.saveSessionAsync(userId, chatId, session)

	ai.removeChatFromRedis(ctx, chatId)
	return nil
}

// TmpToolCallingStreamChat 临时的工具调用流式聊天，所有数据存redis
func (ai *AIService) TmpToolCallingStreamChat(ctx context.Context, userId uint, chatId string, req *request.TmpChatRequest, callback func(string) error) error {
	session, err := ai.getSessionRedis(ctx, chatId)
	if err != nil {
		session = &entity.Session{
			UUID:     chatId,
			UserId:   userId,
			CreateAt: time.Now().Unix(),
			UpdateAt: time.Now().Unix(),
		}
	}
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: global.GetConfig().LLM.BaseUrl,
		APIKey:  global.GetConfig().LLM.ApiKey,
		Model:   global.GetConfig().LLM.ModelName,
	})
	if err != nil {
		return err
	}
	prompt := req.Prompt
	if prompt == "" {
		prompt = " "
	}
	ask := entity.ChatMessage{
		Role:    "user",
		Content: prompt,
		Time:    time.Now().Unix(),
	}
	answer := entity.ChatMessage{
		Role: "assistant",
		Time: time.Now().Unix(),
	}

	// register tools
	tools, toolInfos, err := ai.registerTools(ctx)
	if err != nil {
		return err
	}

	toolCallingChatModel, err := chatModel.WithTools(toolInfos)
	if err != nil {
		return err
	}

	// 构建包含历史的消息
	toolCallingMessage := []*schema.Message{
		schema.SystemMessage("根据用户问题，选择合适工具"),
	}
	// 加入 session 历史消息
	for _, msg := range session.ChatMessages {
		toolCallingMessage = append(toolCallingMessage, &schema.Message{
			Role:    ai.toSchemaRole(msg.Role),
			Content: msg.Content,
		})
	}
	toolCallingMessage = append(toolCallingMessage, schema.UserMessage(prompt))

	toolNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: tools,
	})

	if err != nil {
		return err
	}

	firstResp, err := toolCallingChatModel.Generate(ctx, toolCallingMessage)
	if err != nil {
		return err
	}
	callback(firstResp.Content)

	if len(firstResp.ToolCalls) > 0 {
		//fmt.Printf("模型决定调用工具: %s\n", firstResp.ToolCalls[0].Function.Name)
		//fmt.Printf("调用参数: %s\n\n", firstResp.ToolCalls[0].Function.Arguments)
		callback(fmt.Sprintf("模型决定调用工具: %s\n", firstResp.ToolCalls[0].Function.Name))
		callback(fmt.Sprintf("调用参数: %s\n\n", firstResp.ToolCalls[0].Function.Arguments))

		// 7. ToolsNode 执行工具
		toolResults, err := toolNode.Invoke(ctx, firstResp)
		if err != nil {
			return err
		}

		fmt.Printf("工具返回: %s\n\n", toolResults[0].Content)

		// 8. 把模型的工具调用请求和工具结果追加到消息历史
		toolCallingMessage = append(toolCallingMessage, firstResp)      // 模型的工具调用消息
		toolCallingMessage = append(toolCallingMessage, toolResults...) // 工具执行结果

		// 9. 第二轮：模型根据工具结果生成最终回答
		//finalResp, err := chatModel.Generate(ctx, toolCallingMessage)
		//if err != nil {
		//	log.Fatal(err)
		//}
		finalRespStream, err := chatModel.Stream(ctx, toolCallingMessage)
		if err != nil {
			return err
		}
		for {
			chunk, err := finalRespStream.Recv()
			if err != nil {
				break
			}
			if chunk.Content == "" {
				continue
			}
			callback(chunk.Content)
			answer.Content += chunk.Content
		}

		//fmt.Printf("助手: %s\n", finalResp.Content)
		//callback(finalResp.Content)
	} else {
		//fmt.Printf("助手: %s\n", firstResp.Content)
		callback(firstResp.Content)
	}
	session.ChatMessages = append(session.ChatMessages, ask, answer)
	if err := ai.saveSessionRedis(ctx, chatId, session); err != nil {
		return err
	}
	return nil
}

// registerTools 注册工具并返回工具列表和工具信息
func (ai *AIService) registerTools(ctx context.Context) ([]tool.BaseTool, []*schema.ToolInfo, error) {
	webSearchTool := tomori_tool.NewWebSearchTool(global.GetConfig().Tools.WebSearch.ApiKey)
	searchPostTool := tomori_tool.NewSearchPostTool()
	tushareTool := tomori_tool.NewTushareTool(global.GetConfig().Tools.Tushare.DataDir)

	webSearchToolInfo, err := webSearchTool.Info(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("web search tool info error: %w", err)
	}
	searchPostToolInfo, err := searchPostTool.Info(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("search post tool info error: %w", err)
	}
	tushareToolInfo, err := tushareTool.Info(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("tushare tool info error: %w", err)
	}

	tools := []tool.BaseTool{webSearchTool, searchPostTool, tushareTool}
	toolInfos := []*schema.ToolInfo{webSearchToolInfo, searchPostToolInfo, tushareToolInfo}
	return tools, toolInfos, nil
}

// ToolCallingStreamChat 工具调用的流式chat，需要持久化存储
func (ai *AIService) ToolCallingStreamChat(ctx context.Context, userId uint, chatId string, req *request.InvokeAgentRequest, callback func(string) error) error {
	// read req first
	//maxRetries := 3
	agentId := req.AgentId
	sysPrompt := req.SysPrompt
	usrPrompt := req.UsrPrompt

	// 初始化流式缓存
	ai.SetStreamStatus(ctx, chatId, "streaming")
	ai.ClearStreamChunks(ctx, chatId)

	// streamCallback 同时推送到前端和写入 redis
	streamCallback := func(content string, isError bool, errMsg string) error {
		chunk := entity.StreamChunk{
			ChatId:   chatId,
			Content:  content,
			IsError:  isError,
			ErrorMsg: errMsg,
		}
		ai.AppendStreamChunk(ctx, chatId, chunk)
		return callback(content)
	}

	session, err := ai.loadChat(ctx, userId, chatId)

	if err != nil {
		ai.SetStreamStatus(ctx, chatId, "error")
		ai.SetStreamError(ctx, chatId, err.Error())
		return err
	}
	session.UserId = userId
	session.AgentId = agentId
	session.UpdateAt = time.Now().Unix()
	session.UUID = chatId
	agent, err := ai.QueryAgentById(ctx, userId, agentId)
	if err != nil {
		ai.SetStreamStatus(ctx, chatId, "error")
		ai.SetStreamError(ctx, chatId, err.Error())
		return err
	}
	chatModel, err := ai.getEinoChatModel(ctx, agent)
	if err != nil {
		ai.SetStreamStatus(ctx, chatId, "error")
		ai.SetStreamError(ctx, chatId, err.Error())
		return err
	}

	// register tools
	tools, toolInfos, err := ai.registerTools(ctx)
	if err != nil {
		ai.SetStreamStatus(ctx, chatId, "error")
		ai.SetStreamError(ctx, chatId, err.Error())
		return err
	}
	toolNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: tools,
	})

	toolCallingChatModel, err := chatModel.WithTools(toolInfos)

	firstUsrChatMessage := entity.ChatMessage{
		Role:    "user",
		Content: usrPrompt,
		Time:    time.Now().Unix(),
	}
	firstSysChatMessage := entity.ChatMessage{
		Role:    "system",
		Content: sysPrompt,
		Time:    time.Now().Unix(),
	}
	prompt, err := ai.buildPrompt(ctx, session, firstSysChatMessage, firstUsrChatMessage)
	if err != nil {
		ai.SetStreamStatus(ctx, chatId, "error")
		ai.SetStreamError(ctx, chatId, err.Error())
		return err
	}

	// ReAct 循环：支持多轮工具调用
	// maxToolRounds 最大工具调用轮数，防止无限循环
	const maxToolRounds = 10
	reactMessages := make([]*schema.Message, len(prompt))
	copy(reactMessages, prompt)
	round := 0

	// 累积所有内容到一条消息
	assistantMessage := entity.ChatMessage{
		Role:    "assistant",
		Time:    time.Now().Unix(),
		Content: "",
	}

	for round < maxToolRounds {
		// 使用 Stream 进行流式调用
		streamReader, err := toolCallingChatModel.Stream(ctx, reactMessages)
		if err != nil {
			ai.SetStreamStatus(ctx, chatId, "error")
			ai.SetStreamError(ctx, chatId, err.Error())
			return err
		}

		// 完整的响应，用于判断是否有 ToolCalls
		fullMessage := &schema.Message{Role: schema.Assistant}
		// 累积 ToolCalls 的临时结构
		toolCallsAccumulator := make(map[int]*schema.ToolCall)

		// 读取流式响应
		for {
			chunk, err := streamReader.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				ai.SetStreamStatus(ctx, chatId, "error")
				ai.SetStreamError(ctx, chatId, err.Error())
				return err
			}

			// 流式输出文本内容给前端，并累积到 assistantMessage
			if chunk.Content != "" {
				streamCallback(chunk.Content, false, "")
				fullMessage.Content += chunk.Content
				assistantMessage.Content += chunk.Content
			}

			// 累积 ToolCalls 的增量数据
			for _, tc := range chunk.ToolCalls {
				if _, ok := toolCallsAccumulator[*tc.Index]; !ok {
					toolCallsAccumulator[*tc.Index] = &schema.ToolCall{
						Index:    tc.Index,
						ID:       tc.ID,
						Type:     tc.Type,
						Function: schema.FunctionCall{},
					}
				}
				accTC := toolCallsAccumulator[*tc.Index]
				if tc.ID != "" {
					accTC.ID = tc.ID
				}
				if tc.Type != "" {
					accTC.Type = tc.Type
				}
				if tc.Function.Name != "" {
					accTC.Function.Name += tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					accTC.Function.Arguments += tc.Function.Arguments
				}
			}
		}
		streamReader.Close()

		// 将累积的 ToolCalls 合并到 fullMessage
		if len(toolCallsAccumulator) > 0 {
			fullMessage.ToolCalls = make([]schema.ToolCall, 0, len(toolCallsAccumulator))
			for i := 0; i < len(toolCallsAccumulator); i++ {
				if tc, ok := toolCallsAccumulator[i]; ok {
					fullMessage.ToolCalls = append(fullMessage.ToolCalls, *tc)
				}
			}
		}

		// 如果没有工具调用，结束循环
		if len(fullMessage.ToolCalls) == 0 {
			break
		}

		// 有工具调用，将工具调用信息追加到 assistantMessage
		toolCallInfo := ""
		for _, toolCall := range fullMessage.ToolCalls {
			toolCallInfo += fmt.Sprintf("\n模型决定调用工具%v, 参数为%v\n", toolCall.Function.Name, toolCall.Function.Arguments)
		}
		streamCallback(toolCallInfo, false, "")
		assistantMessage.Content += toolCallInfo

		// 执行工具
		toolResp, err := toolNode.Invoke(ctx, fullMessage)
		if err != nil {
			ai.SetStreamStatus(ctx, chatId, "error")
			ai.SetStreamError(ctx, chatId, err.Error())
			return err
		}

		// 将工具调用和结果追加到消息历史，继续循环
		reactMessages = append(reactMessages, fullMessage)
		reactMessages = append(reactMessages, toolResp...)

		round++
	}

	// 如果循环结束仍未生成最终回答（达到最大轮数）
	if round >= maxToolRounds {
		streamCallback("\n已达到最大工具调用轮数，生成最终回答...\n", false, "")
		assistantMessage.Content += "\n已达到最大工具调用轮数，生成最终回答...\n"
		finalRespStream, err := chatModel.Stream(ctx, reactMessages)
		if err != nil {
			ai.SetStreamStatus(ctx, chatId, "error")
			ai.SetStreamError(ctx, chatId, err.Error())
			return err
		}
		defer finalRespStream.Close()
		for {
			chunk, err := finalRespStream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				ai.SetStreamStatus(ctx, chatId, "error")
				ai.SetStreamError(ctx, chatId, err.Error())
				return err
			}
			if chunk.Content == "" {
				continue
			}
			streamCallback(chunk.Content, false, "")
			assistantMessage.Content += chunk.Content
		}
	}

	// 流式完成
	ai.SetStreamStatus(ctx, chatId, "done")

	// 保存用户消息和助手消息到 session
	session.ChatMessages = append(session.ChatMessages, firstUsrChatMessage, assistantMessage)

	ai.saveSessionAsync(userId, chatId, session)
	return nil
}

func (ai *AIService) InstantAsk(ctx context.Context, sysPrompt string, usrPrompt string) (string, error) {
	if usrPrompt == "" {
		usrPrompt = " "
	}
	if sysPrompt == "" {
		sysPrompt = " "
	}
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: global.GetConfig().LLM.BaseUrl,
		APIKey:  global.GetConfig().LLM.ApiKey,
		Model:   global.GetConfig().LLM.ModelName,
	})
	if err != nil {
		return "", err
	}
	msg := []*schema.Message{
		schema.SystemMessage(sysPrompt),
		schema.UserMessage(usrPrompt),
	}

	res, err := chatModel.Generate(ctx, msg)
	if err != nil {
		return "", err
	}
	return res.Content, nil
}

func (ai *AIService) saveSessionRedis(ctx context.Context, chatId string, session *entity.Session) error {
	sessionJsonString, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return global.GetRedis().Set(ctx, "session_"+chatId, sessionJsonString, 1145*time.Second).Err()

}

func (ai *AIService) getSessionRedis(ctx context.Context, chatId string) (*entity.Session, error) {
	var Session entity.Session
	sessionJsonString, err := global.GetRedis().Get(ctx, "session_"+chatId).Bytes()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(sessionJsonString, &Session); err != nil {
		return nil, err
	}
	return &Session, nil
}

// cacheStreamChat 流式传输的流程是： 后端->redis->前端，这是为了防止前端连接中断导致数据丢失
// redis 只存最后一条消息
func (ai *AIService) cacheStreamChat(ctx context.Context, chatId string, chatMessage *entity.ChatMessage) {
	data, _ := json.Marshal(chatMessage)
	if err := global.GetRedis().Set(ctx, chatId, data, 114*time.Second).Err(); err != nil {
		return
	}
}

// getChatFromRedis 从 Redis 获取聊天消息，返回消息和是否存在
func (ai *AIService) getChatFromRedis(ctx context.Context, chatId string) (*entity.ChatMessage, bool) {
	res, err := global.GetRedis().Get(ctx, chatId).Bytes()
	if err != nil {
		return nil, false
	}

	var chatMessage entity.ChatMessage
	if err := json.Unmarshal(res, &chatMessage); err != nil {
		return nil, false
	}

	return &chatMessage, true
}

func (ai *AIService) toSchemaRole(role string) schema.RoleType {
	switch role {
	case "user":
		return schema.User
	case "system":
		return schema.System
	case "assistant":
		return schema.Assistant
	case "tool":
		return schema.Tool
	default:
		return "user"
	}
}

func (ai *AIService) getLatestChatFromRedis(ctx context.Context, chatId string, callback func(string) error) *entity.ChatMessage {
	var chatMessage entity.ChatMessage
	res, err := global.GetRedis().Get(ctx, chatId).Bytes()
	if err != nil {
		return nil
	}
	err = json.Unmarshal(res, &chatMessage)
	if err != nil {
		return nil
	}
	return &chatMessage
}

func (ai *AIService) removeChatFromRedis(ctx context.Context, chatId string) {
	if err := global.GetRedis().Del(ctx, chatId).Err(); err != nil {
		fmt.Printf("[service] ERROR remove chat from redis: %s\n", err.Error())
		return
	}
	fmt.Printf("[service] remove chat from redis: %s\n", chatId)
}

// isStreaming check if the chat is streaming, redis key: chatId value: chatMessage
func (ai *AIService) isStreaming(ctx context.Context, chatId string) bool {
	var chatMessage entity.ChatMessage
	res, err := global.GetRedis().Get(ctx, chatId).Bytes()
	if err != nil {
		return false
	}
	err = json.Unmarshal(res, &chatMessage)
	if err == nil {
		return false
	}
	return chatMessage.IsDone
}

// ========== 流式缓存相关方法 ==========

const (
	streamStatusTTL = 5 * time.Minute
	streamChunkTTL  = 5 * time.Minute
)

// streamStatusKey 获取流式状态的 redis key
func streamStatusKey(chatId string) string {
	return fmt.Sprintf("stream:%s:status", chatId)
}

// streamChunksKey 获取流式 chunks 的 redis key
func streamChunksKey(chatId string) string {
	return fmt.Sprintf("stream:%s:chunks", chatId)
}

// streamErrorKey 获取流式错误的 redis key
func streamErrorKey(chatId string) string {
	return fmt.Sprintf("stream:%s:error", chatId)
}

// SetStreamStatus 设置流式状态
func (ai *AIService) SetStreamStatus(ctx context.Context, chatId string, status string) error {
	return global.GetRedis().Set(ctx, streamStatusKey(chatId), status, streamStatusTTL).Err()
}

// GetStreamStatus 获取流式状态
func (ai *AIService) GetStreamStatus(ctx context.Context, chatId string) string {
	status, err := global.GetRedis().Get(ctx, streamStatusKey(chatId)).Result()
	if err != nil {
		return ""
	}
	return status
}

// ClearStreamChunks 清空流式 chunks
func (ai *AIService) ClearStreamChunks(ctx context.Context, chatId string) error {
	return global.GetRedis().Del(ctx, streamChunksKey(chatId)).Err()
}

// AppendStreamChunk 追加一个流式 chunk
func (ai *AIService) AppendStreamChunk(ctx context.Context, chatId string, chunk entity.StreamChunk) error {
	data, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	pipe := global.GetRedis().Pipeline()
	pipe.RPush(ctx, streamChunksKey(chatId), data)
	pipe.Expire(ctx, streamChunksKey(chatId), streamChunkTTL)
	_, err = pipe.Exec(ctx)
	return err
}

// GetStreamChunks 获取所有流式 chunks
func (ai *AIService) GetStreamChunks(ctx context.Context, chatId string) []entity.StreamChunk {
	results, err := global.GetRedis().LRange(ctx, streamChunksKey(chatId), 0, -1).Result()
	if err != nil {
		return nil
	}

	chunks := make([]entity.StreamChunk, 0, len(results))
	for _, r := range results {
		var chunk entity.StreamChunk
		if err := json.Unmarshal([]byte(r), &chunk); err != nil {
			continue
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

// SetStreamError 设置流式错误信息
func (ai *AIService) SetStreamError(ctx context.Context, chatId string, errMsg string) error {
	return global.GetRedis().Set(ctx, streamErrorKey(chatId), errMsg, streamStatusTTL).Err()
}

// GetStreamError 获取流式错误信息
func (ai *AIService) GetStreamError(ctx context.Context, chatId string) string {
	errMsg, err := global.GetRedis().Get(ctx, streamErrorKey(chatId)).Result()
	if err != nil {
		return ""
	}
	return errMsg
}

// ClearStreamCache 清空所有流式缓存
func (ai *AIService) ClearStreamCache(ctx context.Context, chatId string) {
	pipe := global.GetRedis().Pipeline()
	pipe.Del(ctx, streamStatusKey(chatId))
	pipe.Del(ctx, streamChunksKey(chatId))
	pipe.Del(ctx, streamErrorKey(chatId))
	pipe.Exec(ctx)
}

// IsStreamActive 检查流式是否正在进行
func (ai *AIService) IsStreamActive(ctx context.Context, chatId string) bool {
	status := ai.GetStreamStatus(ctx, chatId)
	return status == "streaming"
}

// ========== 流式缓存相关方法 END ==========

func (ai *AIService) queryAgentById(ctx context.Context, id uint) (*entity.Agent, error) {
	var agent entity.Agent
	if err := global.GetDB().WithContext(ctx).Where("id = ?", id).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (ai *AIService) getEinoChatModel(ctx context.Context, agent *entity.Agent) (model.ToolCallingChatModel, error) {
	var m model.ToolCallingChatModel
	var err error
	provider := agent.Provider
	apiKey := agent.ApiKey
	baseUrl := agent.BaseUrl
	modelName := agent.ModelName

	switch provider {
	case "openai":
		m, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:      apiKey,
			BaseURL:     baseUrl,
			Model:       modelName,
			Temperature: float32Ptr(float32(agent.Temperature)),
		})

	case "ollama":
		m, err = ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			BaseURL:  baseUrl,
			Model:    modelName,
			Thinking: &api.ThinkValue{Value: agent.Thinking},
		})
	default:
		return nil, errors.New("unsupported provider")
	}

	if err != nil {
		return nil, err
	}
	return m, nil
}

func (ai *AIService) buildPrompt(ctx context.Context, session *entity.Session, newMessages ...entity.ChatMessage) ([]*schema.Message, error) {
	var msg []*schema.Message
	history := session.ChatMessages
	for _, h := range history {
		content := h.Content
		if content == "" {
			content = " "
		}
		msg = append(msg, &schema.Message{
			Role:    ai.toSchemaRole(h.Role),
			Content: content,
		})
	}

	for _, newMessage := range newMessages {
		newContent := newMessage.Content
		if newContent == "" {
			newContent = " "
		}
		msg = append(msg, &schema.Message{
			Role:    ai.toSchemaRole(newMessage.Role),
			Content: newContent,
		})
	}

	return msg, nil
}

func (ai *AIService) getLastKChatMessages(chat *entity.Session, k int) []entity.ChatMessage {
	if chat == nil || k < 0 {
		return nil
	}
	cm := chat.ChatMessages
	if len(cm) <= k {
		return cm
	}

	return cm[len(cm)-k:]
}

func (ai *AIService) loadChat(ctx context.Context, userId uint, chatId string) (*entity.Session, error) {
	var chatSession entity.ChatSession
	err := global.GetDB().WithContext(ctx).Where("uuid = ? AND user_id = ?", chatId, userId).First(&chatSession).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &entity.Session{}, nil
	}
	if err != nil {
		return &entity.Session{}, err
	}

	var messages []entity.ChatMessage
	if chatSession.Messages != nil {
		if err := json.Unmarshal(chatSession.Messages, &messages); err != nil {
			return nil, err
		}
	}

	return &entity.Session{
		UUID:         chatSession.UUID,
		UserId:       chatSession.UserId,
		AgentId:      chatSession.AgentId,
		ChatMessages: messages,
		CreateAt:     chatSession.CreatedAt.Unix(),
		UpdateAt:     chatSession.UpdatedAt.Unix(),
	}, nil
}

func (ai *AIService) saveChatSession(ctx context.Context, userId uint, chatId string, session *entity.Session) error {
	messagesJSON, err := json.Marshal(session.ChatMessages)
	if err != nil {
		return err
	}

	var chatSession entity.ChatSession
	dbResult := global.GetDB().WithContext(ctx).Where("uuid = ? AND user_id = ?", chatId, userId).First(&chatSession)
	title := ""
	if len(session.ChatMessages) > 0 {
		title = session.ChatMessages[0].Content
		if len(title) > 100 {
			title = title[:100]
		}
	}

	if errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
		now := time.Now()
		return global.GetDB().WithContext(ctx).Create(&entity.ChatSession{
			UUID:      chatId,
			UserId:    userId,
			AgentId:   session.AgentId,
			Title:     title,
			Messages:  messagesJSON,
			CreatedAt: time.Unix(session.CreateAt, 0),
			UpdatedAt: now,
		}).Error
	}

	updates := map[string]interface{}{
		"agent_id":   session.AgentId,
		"title":      title,
		"messages":   messagesJSON,
		"updated_at": time.Now(),
	}
	if session.CreateAt > 0 {
		updates["created_at"] = time.Unix(session.CreateAt, 0)
	}
	return global.GetDB().WithContext(ctx).Model(&chatSession).Updates(updates).Error
}

func (ai *AIService) saveSessionAsync(userId uint, chatId string, session *entity.Session) {
	mu, _ := ai.saveMu.LoadOrStore(chatId, &sync.Mutex{})
	go func() {
		mu.(*sync.Mutex).Lock()
		defer mu.(*sync.Mutex).Unlock()
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		messages := make([]entity.ChatMessage, len(session.ChatMessages))
		copy(messages, session.ChatMessages)
		sessCopy := &entity.Session{
			UUID:         session.UUID,
			UserId:       session.UserId,
			AgentId:      session.AgentId,
			ChatMessages: messages,
			CreateAt:     session.CreateAt,
			UpdateAt:     session.UpdateAt,
		}
		if err := ai.saveChatSession(bgCtx, userId, chatId, sessCopy); err != nil {
			fmt.Printf("[service] async save chat failed: %v\n", err)
		}
	}()
}

func float32Ptr(v float32) *float32 {
	return &v
}

func (ai *AIService) DeleteChatSession(ctx context.Context, userId uint, chatId string) error {
	result := global.GetDB().WithContext(ctx).Where("uuid = ? AND user_id = ?", chatId, userId).Delete(&entity.ChatSession{})
	if result.RowsAffected == 0 {
		return errors.New("chat session not found")
	}
	return result.Error
}

func (ai *AIService) UpdateChatSession(ctx context.Context, userId uint, chatId string, title string) error {
	result := global.GetDB().WithContext(ctx).Model(&entity.ChatSession{}).Where("uuid = ? AND user_id = ?", chatId, userId).Update("title", title)
	if result.RowsAffected == 0 {
		return errors.New("chat session not found")
	}
	return result.Error
}
