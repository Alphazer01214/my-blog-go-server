package entity

import (
	"gorm.io/gorm"
)

type Agent struct {
	gorm.Model
	// Uid: owner's id
	UserId uint `gorm:"uniqueKey" json:"user_id"`

	// LLM configs

	Name      string `gorm:"uniqueKey" json:"name"`
	BaseUrl   string `json:"base_url"`
	ApiKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
	Provider  string `json:"provider"`
	Activate  bool   `gorm:"default:false" json:"activate"`

	// options

	Temperature float64 `json:"temperature"`
	Thinking    bool    `json:"thinking"`

	// User personalize
	//
	Prompts  map[string]string `json:"prompts"`
	Memories map[string]string `json:"memories"`
}

type Chat struct {
	UUID         string        `json:"uuid"`
	ChatMessages []ChatMessage `json:"chat_messages"`
	CreateAt     int64         `json:"create_at"`
	UpdateAt     int64         `json:"update_at"`
}

type ChatMessage struct {
	// Role: user or model
	Role    string `json:"role"`
	Content string `json:"content"`
	Time    int64  `json:"time"`
}
