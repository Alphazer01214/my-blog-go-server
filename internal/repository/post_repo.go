package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type PostRepository interface {
	WithTx(tx *gorm.DB) PostRepository
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	Create(ctx context.Context, post *entity.Post) error
	FindById(ctx context.Context, id uint) (*entity.Post, error)
	Save(ctx context.Context, post *entity.Post) error
	DeleteById(ctx context.Context, id uint) error

	ListPosts(ctx context.Context, page, pageSize int) (Pagination[entity.Post], error)
	ListPostsByUserId(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.Post], error)
	SearchByKeyword(ctx context.Context, keyword string, page, pageSize int) (Pagination[entity.Post], error)

	IncrementViewCount(ctx context.Context, id uint) error
	IncrementLikeCount(ctx context.Context, id uint) error
	DecrementLikeCount(ctx context.Context, id uint) error
	IncrementDislikeCount(ctx context.Context, id uint) error
	DecrementDislikeCount(ctx context.Context, id uint) error
	IncrementFavoriteCount(ctx context.Context, id uint) error
	DecrementFavoriteCount(ctx context.Context, id uint) error
	IncrementShareCount(ctx context.Context, id uint) error
	IncrementCommentCount(ctx context.Context, id uint) error
	DecrementCommentCount(ctx context.Context, id uint, delta int) error

	GetForbidComment(ctx context.Context, id uint) (bool, error)
}
