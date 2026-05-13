package response

import (
	"time"

	"github.com/cloudwego/eino/schema"
	"gorm.io/datatypes"
)

type StandardAiResponse struct {
	AgentId          uint   `json:"agent_id"`
	ChatId           string `json:"chat_id,omitempty"`
	ReasoningContent string `json:"reasoning_content"`
	Content          string `json:"content"`
	Status           bool   `json:"status"`
	Message          string `json:"message"`
}

type OnlineStreamChatResponse struct {
	ChatId  string            `json:"chat_id"`
	AgentId uint              `json:"agent_id"`
	UserId  uint              `json:"user_id"`
	History []*schema.Message `json:"history"`
}

type AgentInfo struct {
	AgentId   uint      `json:"agent_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserId    uint      `json:"user_id"`

	// LLM configs
	Name string `json:"name"`
	//BaseUrl   string `json:"base_url"`
	//ApiKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
	Provider  string `json:"provider"`
	Activate  bool   `gorm:"default:false" json:"activate"`

	// options

	Temperature float64 `json:"temperature"`
	Thinking    bool    `json:"thinking"`

	// User personalize
	//
	Prompts  datatypes.JSONMap `json:"prompts"`
	Memories datatypes.JSONMap `json:"memories"`
}
