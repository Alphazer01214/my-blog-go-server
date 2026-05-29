package request

import "gorm.io/datatypes"

type StandardAiRequest struct {
	// Uid       uint            `json:"uid"`
	//Env       *entity.EnvInfo `json:"env"`
	SysPrompt string `json:"sys_prompt"`
	UsrPrompt string `json:"usr_prompt"`
}

type AiOptions struct {
	Temperature    float64 `json:"temperature"`
	Thinking       bool    `json:"thinking"`
	SearchInternet bool    `json:"search_internet"`
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
	AgentId   uint              `json:"agent_id"`
	AgentName string            `json:"agent_name"`
	BaseUrl   string            `json:"base_url"`
	ApiKey    string            `json:"api_key"`
	Provider  string            `json:"provider"`
	ModelName string            `json:"model_name"`
	Prompts   datatypes.JSONMap `json:"prompts"`
	Memories  datatypes.JSONMap `json:"memories"`
	Activate  bool              `json:"activate"`
}

type InvokeAgentRequest struct {
	AgentId   uint `json:"agent_id"`
	AiOptions `json:"ai_options"`
	SysPrompt string `json:"sys_prompt"`
	UsrPrompt string `json:"usr_prompt"`
}

type TmpChatRequest struct {
	Provider string `json:"provider"`
	BaseUrl  string `json:"base_url"`
	ApiKey   string `json:"api_key"`
	Model    string `json:"model"`

	Prompt string `json:"prompt"`
}

type PostAskRequest struct {
	ChatId       string `json:"chat_id"`       // 前端生成的 UUID，追问时传相同值以延续会话
	PostId       uint   `json:"post_id" binding:"required"`
	Prompt       string `json:"prompt"`
	SelectedText string `json:"selected_text"`
	Mode         string `json:"mode"` // "summarize" | "ask" | "selected"

	BaseUrl string `json:"base_url"`
	ApiKey  string `json:"api_key"`
	Model   string `json:"model"`
}

type DeleteChatSessionRequest struct {
	ChatId string `json:"chat_id" binding:"required"`
}

type UpdateChatSessionRequest struct {
	ChatId string `json:"chat_id" binding:"required"`
	Title  string `json:"title"`
}
