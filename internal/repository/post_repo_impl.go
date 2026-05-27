package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type PostRepositoryImpl struct {
	*BaseRepository
}

func NewPostRepository(db *gorm.DB) *PostRepositoryImpl {
	return &PostRepositoryImpl{BaseRepository: NewBaseRepository(db)}
}

func (r *PostRepositoryImpl) WithTx(tx *gorm.DB) PostRepository {
	return &PostRepositoryImpl{BaseRepository: &BaseRepository{DB: tx}}
}

func (r *PostRepositoryImpl) Create(ctx context.Context, post *entity.Post) error {
	return r.DB.WithContext(ctx).Create(post).Error
}

func (r *PostRepositoryImpl) FindById(ctx context.Context, id uint) (*entity.Post, error) {
	var post entity.Post
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&post).Error
	return &post, err
}

func (r *PostRepositoryImpl) Save(ctx context.Context, post *entity.Post) error {
	return r.DB.WithContext(ctx).Save(post).Error
}

func (r *PostRepositoryImpl) DeleteById(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Delete(&entity.Post{}, id).Error
}

func (r *PostRepositoryImpl) ListPosts(ctx context.Context, page, pageSize int) (Pagination[entity.Post], error) {
	db := r.DB.WithContext(ctx).Model(&entity.Post{})
	return Paginate[entity.Post](db, page, pageSize, "created_at desc")
}

func (r *PostRepositoryImpl) ListPostsByUserId(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.Post], error) {
	db := r.DB.WithContext(ctx).Model(&entity.Post{}).Where("user_id = ?", userId)
	return Paginate[entity.Post](db, page, pageSize, "created_at desc")
}

func (r *PostRepositoryImpl) SearchByKeyword(ctx context.Context, keyword string, page, pageSize int) (Pagination[entity.Post], error) {
	qKeyword := "%" + keyword + "%"
	db := r.DB.WithContext(ctx).Model(&entity.Post{}).Where("title ILIKE ? OR content ILIKE ?", qKeyword, qKeyword)
	return Paginate[entity.Post](db, page, pageSize, "created_at desc")
}

func (r *PostRepositoryImpl) IncrementViewCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *PostRepositoryImpl) IncrementLikeCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *PostRepositoryImpl) DecrementLikeCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

func (r *PostRepositoryImpl) IncrementDislikeCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("dislike_count", gorm.Expr("dislike_count + 1")).Error
}

func (r *PostRepositoryImpl) DecrementDislikeCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
}

func (r *PostRepositoryImpl) IncrementFavoriteCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("favorite_count", gorm.Expr("favorite_count + 1")).Error
}

func (r *PostRepositoryImpl) DecrementFavoriteCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error
}

func (r *PostRepositoryImpl) IncrementShareCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *PostRepositoryImpl) IncrementCommentCount(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("comment_count", gorm.Expr("comment_count + 1")).Error
}

func (r *PostRepositoryImpl) DecrementCommentCount(ctx context.Context, id uint, delta int) error {
	return r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).
		Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delta)).Error
}

func (r *PostRepositoryImpl) GetForbidComment(ctx context.Context, id uint) (bool, error) {
	var res bool
	err := r.DB.WithContext(ctx).Model(&entity.Post{}).Where("id = ?", id).Pluck("forbid_comment", &res).Error
	return res, err
}
