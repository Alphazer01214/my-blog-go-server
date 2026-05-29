package repository

import (
	"context"

	"blog.alphazer01214.top/internal/constant"
)

type ActionRepository interface {
	DoAction(ctx context.Context, userId uint, targetId uint, targetType constant.TargetType, actionType constant.ActionType) error
	Like(ctx context.Context, userId uint, targetId uint, targetType constant.TargetType) error
	Dislike(ctx context.Context, userId uint, targetId uint, targetType constant.TargetType) error
	Favorite(ctx context.Context, userId uint, targetId uint, targetType constant.TargetType) error
	IsLiked(ctx context.Context, userId uint, targetId uint, targetType constant.TargetType) error
	IsDisliked(ctx context.Context, userId uint, targetId uint, targetType constant.TargetType) error

	CalculateLikeCount(ctx context.Context, userId uint, targetId uint, targetType constant.TargetType) error
}
