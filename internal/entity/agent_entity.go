package entity

import (
	"gorm.io/datatypes"
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
	Prompts  datatypes.JSONMap `json:"prompts"`
	Memories datatypes.JSONMap `json:"memories"`
}

type Session struct {
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
