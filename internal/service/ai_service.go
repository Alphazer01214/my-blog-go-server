package service

import (
	"context"
	"errors"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
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
	case *openai.ChatModel:
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
			Message: "",
		}, nil

	case *ollama.ChatModel:
		content, err := t.Generate(ctx, message)
		if err != nil {
			return &response.StandardAiResponse{
				AgentId: agentId,
				Status:  false,
				Message: err.Error(),
			}, err
		}
		return &response.StandardAiResponse{
			AgentId:          agentId,
			Status:           true,
			Content:          content.Content,
			ReasoningContent: content.ReasoningContent,
			Message:          "",
		}, nil

	default:
		return &response.StandardAiResponse{
			AgentId: agentId,
			Status:  false,
			Message: "unsupported model",
		}, errors.New("unsupported model")
	}
}

func (ai *AIService) StreamChat(ctx context.Context) {
}

func (ai *AIService) queryAgentById(ctx context.Context, id uint) (*entity.Agent, error) {
	var agent *entity.Agent
	if err := global.GetDB().WithContext(ctx).First(agent, id).Error; err != nil {
		return nil, err
	}
	return agent, nil
}

func (ai *AIService) getEinoChatModel(ctx context.Context, agent *entity.Agent) (interface{}, error) {
	var model interface{}
	var err error
	provider := agent.Provider
	apiKey := agent.ApiKey
	baseUrl := agent.BaseUrl
	modelName := agent.ModelName

	switch provider {
	case "openai":
		model, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  apiKey,
			BaseURL: baseUrl,
			Model:   modelName,
		})

	case "ollama":
		model, err = ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			BaseURL: baseUrl,
			Model:   modelName,
		})
	default:
		return nil, errors.New("unsupported provider")
	}

	if err != nil {
		return nil, err
	}
	return model, nil
}
