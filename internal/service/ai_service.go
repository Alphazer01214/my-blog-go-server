package service

import (
	"context"
	"errors"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
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
	if err := global.GetDB().Create(agent).Error; err != nil {
		return nil, err
	}
	return agent, nil
}

func (ai *AIService) UpdateAgent(userId uint, agentId uint, req *request.UpdateAgentRequest) (*entity.Agent, error) {
	agent, err := ai.queryAgentById(agentId)
	if err != nil {
		return nil, err
	}
	if agent.UserId != userId {
		return nil, errors.New("unauthorized")
	}
	return agent, nil
}

func (ai *AIService) queryAgentById(id uint) (*entity.Agent, error) {
	var agent *entity.Agent
	if err := global.GetDB().First(agent, id).Error; err != nil {
		return nil, err
	}
	return agent, nil
}
