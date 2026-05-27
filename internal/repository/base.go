package repository

import (
	"context"

	"gorm.io/gorm"
)

type Pagination[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type BaseRepository struct {
	DB *gorm.DB
}

func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{DB: db}
}

func (b *BaseRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return b.DB.WithContext(ctx).Transaction(fn)
}

func Paginate[T any](db *gorm.DB, page, pageSize int, order string) (Pagination[T], error) {
	var items []T
	var total int64

	if err := db.Count(&total).Error; err != nil {
		return Pagination[T]{}, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order(order).Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return Pagination[T]{}, err
	}

	return Pagination[T]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
