package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"github.com/redis/go-redis/v9"
)

type AIService struct{}

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

func (ai *AIService) InvokeAgent(ctx context.Context, userId uint, agentId uint, req *request.InvokeAgentRequest) (*response.StandardAiResponse, error) {
	var chatModel interface{}
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
			Message: err.Error(),
		}, errors.New("unauthorized")
	}
	message := []*schema.Message{
		schema.UserMessage(req.UsrPrompt),
		schema.SystemMessage(req.SysPrompt),
	}

	chatModel, err = ai.getEinoChatModel(ctx, agent)
	if err != nil {
		return &response.StandardAiResponse{
			AgentId: agentId,
			Status:  false,
			Message: err.Error(),
		}, err
	}

	switch t := chatModel.(type) {
	//case *openai.ChatModel:
	//	content, err := t.Generate(ctx, message)
	//	if err != nil {
	//		return &response.StandardAiResponse{
	//			AgentId: agentId,
	//			Status:  false,
	//			Message: err.Error(),
	//		}, err
	//	}
	//
	//	return &response.StandardAiResponse{
	//		AgentId: agentId,
	//		Status:  true,
	//		Content: content.Content,
	//		ReasoningContent: content.ReasoningContent,
	//		Message: "",
	//	}, nil
	//
	//case *ollama.ChatModel:
	//	content, err := t.Generate(ctx, message)
	//	if err != nil {
	//		return &response.StandardAiResponse{
	//			AgentId: agentId,
	//			Status:  false,
	//			Message: err.Error(),
	//		}, err
	//	}
	//	return &response.StandardAiResponse{
	//		AgentId:          agentId,
	//		Status:           true,
	//		Content:          content.Content,
	//		ReasoningContent: content.ReasoningContent,
	//		Message:          "",
	//	}, nil

	case model.ToolCallingChatModel:
		content, err := t.Generate(ctx, message)
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
			//ReasoningContent: content.ReasoningContent,
			Message: "success",
		}, nil

	default:
		return &response.StandardAiResponse{
			AgentId: agentId,
			Status:  false,
			Message: "unsupported model",
		}, errors.New("unsupported model")
	}
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
	session, err := ai.loadChat(ctx, chatId)
	if err != nil {
		return err
	}
	ask := entity.ChatMessage{
		Role:    "user",
		Content: req.UsrPrompt,
		Time:    time.Now().Unix(),
	}
	prompt, err := ai.buildPrompt(ctx, session, ask)
	if err != nil {
		return err
	}
	stream, err := chatModel.Stream(ctx, prompt)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if err := callback(chunk.Content); err != nil {
			return err
		}
	}

	return nil
}

func (ai *AIService) queryAgentById(ctx context.Context, id uint) (*entity.Agent, error) {
	var agent entity.Agent
	if err := global.GetDB().WithContext(ctx).First(&agent, id).Error; err != nil {
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
			APIKey:  apiKey,
			BaseURL: baseUrl,
			Model:   modelName,
		})

	case "ollama":
		m, err = ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			BaseURL:  baseUrl,
			Model:    modelName,
			Thinking: &api.ThinkValue{Value: false},
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

func (ai *AIService) loadChat(ctx context.Context, chatId string) (*entity.Session, error) {
	var chat entity.Session
	data, err := global.GetRedis().Get(ctx, chatId).Bytes()
	if errors.Is(err, redis.Nil) {
		return &chat, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &chat); err != nil {
		return nil, err
	}

	return &chat, nil
}

func (ai *AIService) saveChat(ctx context.Context, chatId string, chat *entity.Session) error {
	data, err := json.Marshal(chat)
	if err != nil {
		return err
	}
	return global.GetRedis().Set(ctx, chatId, data, 0).Err()
}
