package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
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
func (ai *AIService) GetChatHistory(ctx context.Context, userId uint, chatId string) []*schema.Message {
	session, err := ai.loadChat(ctx, userId, chatId)
	if err != nil || session == nil {
		return nil
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
	message := []*schema.Message{
		schema.UserMessage(req.UsrPrompt),
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
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// stream 到结尾，跳出
			break
		}
		cm := entity.ChatMessage{
			Role:     "model",
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

// cacheStreamChat 流式传输的流程是： 后端->redis->前端，这是为了防止前端连接中断导致数据丢失
// redis 只存最后一条消息
func (ai *AIService) cacheStreamChat(ctx context.Context, chatId string, chatMessage *entity.ChatMessage) {
	data, _ := json.Marshal(chatMessage)
	if err := global.GetRedis().Set(ctx, chatId, data, 114*time.Second).Err(); err != nil {
		return
	}
}

// GetChatFromRedis 从 Redis 获取聊天消息，返回消息和是否存在
func (ai *AIService) GetChatFromRedis(ctx context.Context, chatId string) (*entity.ChatMessage, bool) {
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

// WaitForStreamComplete 等待流式完成，期间持续推送更新给前端
func (ai *AIService) WaitForStreamComplete(ctx context.Context, userId uint, chatId string, callback func(string) error) error {
	fmt.Printf("[service] waiting for stream complete: chatId=%s\n", chatId)

	lastContent := ""
	timeout := time.After(120 * time.Second)         // 最多等待 2 分钟
	ticker := time.NewTicker(500 * time.Millisecond) // 每 500ms 检查一次
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			fmt.Printf("[service] wait for stream timeout: chatId=%s\n", chatId)
			return errors.New("wait for stream timeout")

		case <-ticker.C:
			// 检查 Redis 中的最新内容
			chatMsg, exists := ai.GetChatFromRedis(ctx, chatId)
			if !exists || chatMsg == nil {
				// Redis 中没有数据，可能已经被清理
				fmt.Printf("[service] redis data cleared, checking DB: chatId=%s\n", chatId)
				return nil
			}

			// 如果有新内容，推送给前端
			if chatMsg.Content != lastContent && len(chatMsg.Content) > len(lastContent) {
				// 只推送新增的部分
				newContent := chatMsg.Content[len(lastContent):]
				if err := callback(newContent); err != nil {
					fmt.Printf("[service] callback error while waiting: %v\n", err)
					return err
				}
				lastContent = chatMsg.Content
			}

			// 检查是否完成
			if chatMsg.IsDone {
				fmt.Printf("[service] stream completed while waiting: chatId=%s, length=%d\n", chatId, len(chatMsg.Content))
				if chatMsg.IsError {
					return fmt.Errorf("stream completed with error: %s", chatMsg.ErrorMsg)
				}
				return nil
			}
		}
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

func (ai *AIService) buildPrompt(ctx context.Context, session *entity.Session, newMessage entity.ChatMessage) ([]*schema.Message, error) {
	var msg []*schema.Message
	history := session.ChatMessages
	for _, h := range history {
		var rt schema.RoleType
		role := h.Role
		if role == "model" {
			rt = schema.Assistant
		} else {
			rt = schema.User
		}
		msg = append(msg, &schema.Message{
			Role:    rt,
			Content: h.Content,
		})
	}
	msg = append(msg, &schema.Message{
		Role:    schema.User,
		Content: newMessage.Content,
	})

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
		return nil, err
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

func (ai *AIService) saveChat(ctx context.Context, userId uint, chatId string, session *entity.Session) error {
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
		if err := ai.saveChat(bgCtx, userId, chatId, sessCopy); err != nil {
			fmt.Printf("[service] async save chat failed: %v\n", err)
		}
	}()
}

func float32Ptr(v float32) *float32 {
	return &v
}
