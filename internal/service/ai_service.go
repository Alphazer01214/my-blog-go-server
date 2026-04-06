package service

import (
	"context"
	"errors"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

type AIService struct{}

func (ai *AIService) CreateAgent(ctx context.Context, userId uint, req *request.CreateAgentRequest) (*entity.Agent, error) {
	agent := &entity.Agent{
		UserId:    userId,
		Name:      req.AgentName,
		Provider:  req.Provider,
		BaseUrl:   req.BaseUrl,
		ModelName: req.ModelName,
	}
	if err := global.GetDB().WithContext(ctx).Create(agent).Error; err != nil {
		return nil, err
	}
	return agent, nil
}

func (ai *AIService) UpdateAgent(ctx context.Context, userId uint, agentId uint, req *request.UpdateAgentRequest) (*entity.Agent, error) {
	agent, err := ai.queryAgentById(ctx, agentId)
	if err != nil {
		return nil, err
	}
	if agent.UserId != userId {
		return nil, errors.New("unauthorized")
	}
	return agent, nil
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
