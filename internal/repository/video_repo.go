package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type VideoRepository interface {
	WithTx(tx *gorm.DB) VideoRepository
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	Create(ctx context.Context, video *entity.Video) error
	FindById(ctx context.Context, id uint) (*entity.Video, error)
	Save(ctx context.Context, video *entity.Video) error
	DeleteById(ctx context.Context, id uint) error

	ListVideos(ctx context.Context, publicOnly bool, page, pageSize int) (Pagination[entity.Video], error)
	ListVideosByUserId(ctx context.Context, userId uint, publicOnly bool, page, pageSize int) (Pagination[entity.Video], error)
	ListVideosByCategory(ctx context.Context, category string, publicOnly bool, page, pageSize int) (Pagination[entity.Video], error)
	SearchByKeyword(ctx context.Context, keyword string, publicOnly bool, page, pageSize int) (Pagination[entity.Video], error)

	UpdateUrls(ctx context.Context, id uint, srcUrl, coverUrl string, duration int) error

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
