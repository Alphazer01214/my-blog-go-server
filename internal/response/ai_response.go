package response

import "github.com/cloudwego/eino/schema"

type StandardAiResponse struct {
	AgentId          uint   `json:"agent_id"`
	ReasoningContent string `json:"reasoning_content"`
	Content          string `json:"content"`
	Status           bool   `json:"status"`
	Message          string `json:"message"`
}

type OnlineStreamChatResponse struct {
	ChatId  string `json:"chat_id"`
	AgentId uint   `json:"agent_id"`
	UserId  uint   `json:"user_id"`
	History []*schema.Message
}
