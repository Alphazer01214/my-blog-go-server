package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type AIRepository interface {
	WithTx(tx *gorm.DB) AIRepository
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	// Agent
	CreateAgent(ctx context.Context, agent *entity.Agent) error
	FindAgentById(ctx context.Context, id uint) (*entity.Agent, error)
	SaveAgent(ctx context.Context, agent *entity.Agent) error
	ListAgentsByUserId(ctx context.Context, userId uint) ([]entity.Agent, error)
	DeleteAgentById(ctx context.Context, id uint) (int64, error)

	// ChatSession
	CreateChatSession(ctx context.Context, session *entity.ChatSession) error
	FindChatSessionByUuidAndUserId(ctx context.Context, uuid string, userId uint) (*entity.ChatSession, error)
	UpdateChatSessionFields(ctx context.Context, uuid string, userId uint, updates map[string]interface{}) error
	ListChatSessionsByUserId(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.ChatSession], error)
	DeleteChatSessionByUuid(ctx context.Context, uuid string) (int64, error)
}
