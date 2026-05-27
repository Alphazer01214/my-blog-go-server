package repository

import (
	"context"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type CommentRepository interface {
	WithTx(tx *gorm.DB) CommentRepository
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	Create(ctx context.Context, comment *entity.Comment) error
	FindById(ctx context.Context, id uint) (*entity.Comment, error)
	DeleteById(ctx context.Context, id uint) error

	ListRootComments(ctx context.Context, targetType constant.TargetType, targetId uint, page, pageSize int) (Pagination[entity.Comment], error)
	ListCommentsByUserId(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.Comment], error)
	FindRepliesByRootId(ctx context.Context, rootId, parentId uint) ([]entity.Comment, error)

	DeleteByRootCommentId(ctx context.Context, rootCommentId uint) (int64, error)
	DeleteByParentCommentId(ctx context.Context, parentCommentId uint) (int64, error)

	IncrementReplyCount(ctx context.Context, commentId uint) error
	DecrementReplyCount(ctx context.Context, commentId uint, delta int) error
	IncrementLikeCount(ctx context.Context, commentId uint) error
	DecrementLikeCount(ctx context.Context, commentId uint) error
	IncrementDislikeCount(ctx context.Context, commentId uint) error
	DecrementDislikeCount(ctx context.Context, commentId uint) error
}
