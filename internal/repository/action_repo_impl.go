package repository

import (
	"context"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type ActionRepositoryImpl struct {
	*BaseRepository
}

func NewActionRepository(db *gorm.DB) *ActionRepositoryImpl {
	return &ActionRepositoryImpl{BaseRepository: NewBaseRepository(db)}
}

func (r *ActionRepositoryImpl) WithTx(tx *gorm.DB) ActionRepository {
	return &ActionRepositoryImpl{BaseRepository: &BaseRepository{DB: tx}}
}

func (r *ActionRepositoryImpl) Find(ctx context.Context, userId, targetId uint, actType constant.ActionType, tgtType constant.TargetType) (*entity.Action, error) {
	var action entity.Action
	err := r.DB.WithContext(ctx).
		Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			targetId, userId, tgtType, actType).
		First(&action).Error
	return &action, err
}

func (r *ActionRepositoryImpl) Exists(ctx context.Context, userId, targetId uint, actType constant.ActionType, tgtType constant.TargetType) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&entity.Action{}).
		Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			targetId, userId, tgtType, actType).
		Count(&count).Error
	return count > 0, err
}

func (r *ActionRepositoryImpl) Create(ctx context.Context, action *entity.Action) error {
	return r.DB.WithContext(ctx).Create(action).Error
}

func (r *ActionRepositoryImpl) Delete(ctx context.Context, userId, targetId uint, actType constant.ActionType, tgtType constant.TargetType) (int64, error) {
	result := r.DB.WithContext(ctx).
		Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			targetId, userId, tgtType, actType).
		Delete(&entity.Action{})
	return result.RowsAffected, result.Error
}
