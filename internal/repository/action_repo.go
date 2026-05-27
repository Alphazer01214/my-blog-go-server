package repository

import (
	"context"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type ActionRepository interface {
	WithTx(tx *gorm.DB) ActionRepository
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	Find(ctx context.Context, userId, targetId uint, actType constant.ActionType, tgtType constant.TargetType) (*entity.Action, error)
	Exists(ctx context.Context, userId, targetId uint, actType constant.ActionType, tgtType constant.TargetType) (bool, error)
	Create(ctx context.Context, action *entity.Action) error
	Delete(ctx context.Context, userId, targetId uint, actType constant.ActionType, tgtType constant.TargetType) (int64, error)
}
