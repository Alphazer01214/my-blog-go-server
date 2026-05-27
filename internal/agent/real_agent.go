package agent

import (
	"github.com/cloudwego/eino/components/model"
)

type ChatRole string

const (
	User      ChatRole = "user"
	Assistant ChatRole = "assistant"
)

type ChatMode string

const (
	Text      ChatMode = "text"
	Reasoning ChatMode = "reasoning"
	ToolCall  ChatMode = "tool_call"
)

type Session struct {
	UUID         string        `json:"uuid"`
	ChatMessages []ChatMessage `json:"chat_messages"`
}

type ChatMessage struct {
	//ChatId   int    `json:"chat_id"`
	//ParentId int    `json:"parent_id"`
	Role    string `json:"role"`
	Content string `json:"content"`

	ToolCallId string `json:"tool_call_id"`
}

type Agent struct {
	Model model.ToolCallingChatModel
}
