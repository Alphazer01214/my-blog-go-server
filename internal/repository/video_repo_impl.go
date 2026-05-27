package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type VideoRepositoryImpl struct {
	*BaseRepository
}

func NewVideoRepository(db *gorm.DB) *VideoRepositoryImpl {
	return &VideoRepositoryImpl{BaseRepository: NewBaseRepository(db)}
}

func (r *VideoRepositoryImpl) WithTx(tx *gorm.DB) VideoRepository {
	return &VideoRepositoryImpl{BaseRepository: &BaseRepository{DB: tx}}
}

func (r *VideoRepositoryImpl) Create(ctx context.Context, video *entity.Video) error {
	return r.DB.WithContext(ctx).Create(video).Error
}

func (r *VideoRepositoryImpl) FindById(ctx context.Context, id uint) (*entity.Video, error) {
	var video entity.Video
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&video).Error
	return &video, err
}

func (r *VideoRepositoryImpl) Save(ctx context.Context, video *entity.Video) error {
	return r.DB.WithContext(ctx).Save(video).Error
}

func (r *VideoRepositoryImpl) DeleteById(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Delete(&entity.Video{}, id).Error
}

func (r *VideoRepositoryImpl) ListVideos(ctx context.Context, publicOnly bool, page, pageSize int) (Pagination[entity.Video], error) {
	db := r.DB.WithContext(ctx).Model(&entity.Video{})
	if publicOnly {
		db = db.Where("public = ?", true)
	}
	return Paginate[entity.Video](db, page, pageSize, "created_at desc")
}

func (r *VideoRepositoryImpl) ListVideosByUserId(ctx context.Context, userId uint, publicOnly bool, page, pageSize int) (Pagination[entity.Video], error) {
	db := r.DB.WithContext(ctx).Model(&entity.Video{}).Where("user_id = ?", userId)
	if publicOnly {
		db = db.Where("public = ?", true)
	}
	return Paginate[entity.Video](db, page, pageSize, "created_at desc")
}

func (r *VideoRepositoryImpl) ListVideosByCategory(ctx context.Context, category string, publicOnly bool, page, pageSize int) (Pagination[entity.Video], error) {
	db := r.DB.WithContext(ctx).Model(&entity.Video{}).Where("category = ?", category)
	if publicOnly {
		db = db.Where("public = ?", true)
	}
	return Paginate[entity.Video](db, page, pageSize, "created_at desc")
}

func (r *VideoRepositoryImpl) SearchByKeyword(ctx context.Context, keyword string, publicOnly bool, page, pageSize int) (Pagination[entity.Video], error) {
	qKeyword := "%" + keyword + "%"
	db := r.DB.WithContext(ctx).Model(&entity.Video{}).Where("title ILIKE ? OR description ILIKE ?", qKeyword, qKeyword)
	if publicOnly {
		db = db.Where("public = ?", true)
	}
	return Paginate[entity.Video](db, page, pageSize, "created_at desc")
}

func (r *VideoRepositoryImpl) UpdateUrls(ctx context.Context, id uint, srcUrl, coverUrl string, duration int) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"video_src_url":   srcUrl,
			"video_cover_url": coverUrl,
			"duration":        duration,
		}).Error
}

func (r *VideoRepositoryImpl) IncrementViewCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *VideoRepositoryImpl) IncrementLikeCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *VideoRepositoryImpl) DecrementLikeCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

func (r *VideoRepositoryImpl) IncrementDislikeCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("dislike_count", gorm.Expr("dislike_count + 1")).Error
}

func (r *VideoRepositoryImpl) DecrementDislikeCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
}

func (r *VideoRepositoryImpl) IncrementFavoriteCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("favorite_count", gorm.Expr("favorite_count + 1")).Error
}

func (r *VideoRepositoryImpl) DecrementFavoriteCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error
}

func (r *VideoRepositoryImpl) IncrementShareCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *VideoRepositoryImpl) IncrementCommentCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("comment_count", gorm.Expr("comment_count + 1")).Error
}

func (r *VideoRepositoryImpl) DecrementCommentCount(ctx context.Context, id uint, delta int) error {
	return r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).
		Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delta)).Error
}

func (r *VideoRepositoryImpl) GetForbidComment(ctx context.Context, id uint) (bool, error) {
	var res bool
	err := r.DB.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).Pluck("forbid_comment", &res).Error
	return res, err
}
