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

	// User personalize
	//
	Prompts  map[string]string `json:"prompts"`
	Memories map[string]string `json:"memories"`
}
