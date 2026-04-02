package request

import (
	"blog.alphazer01214.top/internal/entity"
)

type StandardAiRequest struct {
	// Uid       uint            `json:"uid"`
	AgentName string          `json:"agent_name"`
	Env       *entity.EnvInfo `json:"env"`
	SysPrompt string          `json:"sys_prompt`
	UsrPrompt string          `json:"usr_prompt"`
}

type CreateAgentRequest struct {
	// uid must fetch in context, otherwise the request may be faked
	//	Uid       uint   `json:"uid"`
	AgentName string `json:"agent_name"`
	BaseUrl   string `json:"base_url"`
	ApiKey    string `json:"api_key"`
	Provider  string `json:"provider"`
	ModelName string `json:"model_name"`
}

type UpdateAgentRequest struct {
	AgentName string            `json:"agent_name"`
	BaseUrl   string            `json:"base_url"`
	ApiKey    string            `json:"api_key"`
	Provider  string            `json:"provider"`
	ModelName string            `json:"model_name"`
	Prompts   map[string]string `json:"prompts"`
	Memories  map[string]string `json:"memories"`
	Activete  bool              `json:"activate"`
}
