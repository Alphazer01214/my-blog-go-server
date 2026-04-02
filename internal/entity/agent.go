package entity

import (
	"context"
	"errors"

	"blog.alphazer01214.top/internal/config"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

type Agent struct {
	Name      string `json:"agent_name"`
	ModelName string `json:"model_name"`
	Provider  string `json:"provider"`
	//	BaseUrl   string `json:"base_url"`
	//	ApiKey    string `json:"-"`
	ChatModel interface{}
	Ctx       context.Context
}

func NewAgent(agentName string, cfg *config.LLM) (*Agent, error) {
	var chatModel interface{}
	var err error
	apiKey := cfg.ApiKey
	modelName := cfg.ModelName
	baseUrl := cfg.BaseUrl
	provider := cfg.Provider
	ctx := context.Background()
	switch provider {
	case "openai":
		chatModel, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  apiKey,
			BaseURL: baseUrl,
			Model:   modelName,
		})
		if err != nil {
			return nil, err
		}

	case "ollama":
		chatModel, err = ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			BaseURL: baseUrl,
			Model:   modelName,
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unkown model provider")
	}
	return &Agent{
		Name:      agentName,
		Provider:  provider,
		ChatModel: chatModel,
		Ctx:       ctx,
	}, nil
}

func (ag *Agent) Ask(sysPrompt string, usrPrompt string) (interface{}, error) {
	var resp interface{}
	var err error
	messages := []*schema.Message{
		schema.SystemMessage(sysPrompt),
		schema.UserMessage(usrPrompt),
	}

	switch model := ag.ChatModel.(type) {
	case *openai.ChatModel:
		resp, err = model.Generate(ag.Ctx, messages)
	default:
		return nil, errors.New("unknown error")

	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}
