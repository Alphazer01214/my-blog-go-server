package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type AIRepositoryImpl struct {
	*BaseRepository
}

func NewAIRepository(db *gorm.DB) *AIRepositoryImpl {
	return &AIRepositoryImpl{BaseRepository: NewBaseRepository(db)}
}

func (r *AIRepositoryImpl) WithTx(tx *gorm.DB) AIRepository {
	return &AIRepositoryImpl{BaseRepository: &BaseRepository{DB: tx}}
}

func (r *AIRepositoryImpl) CreateAgent(ctx context.Context, agent *entity.Agent) error {
	return r.DB.WithContext(ctx).Create(agent).Error
}

func (r *AIRepositoryImpl) FindAgentById(ctx context.Context, id uint) (*entity.Agent, error) {
	var agent entity.Agent
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&agent).Error
	return &agent, err
}

func (r *AIRepositoryImpl) SaveAgent(ctx context.Context, agent *entity.Agent) error {
	return r.DB.WithContext(ctx).Save(agent).Error
}

func (r *AIRepositoryImpl) ListAgentsByUserId(ctx context.Context, userId uint) ([]entity.Agent, error) {
	var agents []entity.Agent
	err := r.DB.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("id desc").
		Find(&agents).Error
	return agents, err
}

func (r *AIRepositoryImpl) DeleteAgentById(ctx context.Context, id uint) (int64, error) {
	result := r.DB.WithContext(ctx).Delete(&entity.Agent{}, id)
	return result.RowsAffected, result.Error
}

func (r *AIRepositoryImpl) CreateChatSession(ctx context.Context, session *entity.ChatSession) error {
	return r.DB.WithContext(ctx).Create(session).Error
}

func (r *AIRepositoryImpl) FindChatSessionByUuidAndUserId(ctx context.Context, uuid string, userId uint) (*entity.ChatSession, error) {
	var session entity.ChatSession
	err := r.DB.WithContext(ctx).
		Where("uuid = ? AND user_id = ?", uuid, userId).
		First(&session).Error
	return &session, err
}

func (r *AIRepositoryImpl) UpdateChatSessionFields(ctx context.Context, uuid string, userId uint, updates map[string]interface{}) error {
	return r.DB.WithContext(ctx).
		Model(&entity.ChatSession{}).
		Where("uuid = ? AND user_id = ?", uuid, userId).
		Updates(updates).Error
}

func (r *AIRepositoryImpl) ListChatSessionsByUserId(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.ChatSession], error) {
	db := r.DB.WithContext(ctx).
		Model(&entity.ChatSession{}).
		Where("user_id = ?", userId)
	return Paginate[entity.ChatSession](db, page, pageSize, "updated_at desc")
}

func (r *AIRepositoryImpl) DeleteChatSessionByUuid(ctx context.Context, uuid string) (int64, error) {
	result := r.DB.WithContext(ctx).
		Where("uuid = ?", uuid).
		Delete(&entity.ChatSession{})
	return result.RowsAffected, result.Error
}
