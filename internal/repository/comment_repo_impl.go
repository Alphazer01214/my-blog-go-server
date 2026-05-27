package repository

import (
	"context"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type CommentRepositoryImpl struct {
	*BaseRepository
}

func NewCommentRepository(db *gorm.DB) *CommentRepositoryImpl {
	return &CommentRepositoryImpl{BaseRepository: NewBaseRepository(db)}
}

func (r *CommentRepositoryImpl) WithTx(tx *gorm.DB) CommentRepository {
	return &CommentRepositoryImpl{BaseRepository: &BaseRepository{DB: tx}}
}

func (r *CommentRepositoryImpl) Create(ctx context.Context, comment *entity.Comment) error {
	return r.DB.WithContext(ctx).Create(comment).Error
}

func (r *CommentRepositoryImpl) FindById(ctx context.Context, id uint) (*entity.Comment, error) {
	var comment entity.Comment
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&comment).Error
	return &comment, err
}

func (r *CommentRepositoryImpl) DeleteById(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Delete(&entity.Comment{}, id).Error
}

func (r *CommentRepositoryImpl) ListRootComments(ctx context.Context, targetType constant.TargetType, targetId uint, page, pageSize int) (Pagination[entity.Comment], error) {
	db := r.DB.WithContext(ctx).
		Model(&entity.Comment{}).
		Where("target_id = ? AND target_type = ? AND root_comment_id = 0", targetId, targetType)
	return Paginate[entity.Comment](db, page, pageSize, "created_at desc")
}

func (r *CommentRepositoryImpl) ListCommentsByUserId(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.Comment], error) {
	db := r.DB.WithContext(ctx).
		Model(&entity.Comment{}).
		Where("user_id = ?", userId)
	return Paginate[entity.Comment](db, page, pageSize, "created_at desc")
}

func (r *CommentRepositoryImpl) FindRepliesByRootId(ctx context.Context, rootId, parentId uint) ([]entity.Comment, error) {
	var children []entity.Comment
	err := r.DB.WithContext(ctx).
		Where("root_comment_id = ? AND parent_comment_id = ?", rootId, parentId).
		Find(&children).Error
	return children, err
}

func (r *CommentRepositoryImpl) DeleteByRootCommentId(ctx context.Context, rootCommentId uint) (int64, error) {
	result := r.DB.WithContext(ctx).
		Where("root_comment_id = ?", rootCommentId).
		Delete(&entity.Comment{})
	return result.RowsAffected, result.Error
}

func (r *CommentRepositoryImpl) DeleteByParentCommentId(ctx context.Context, parentCommentId uint) (int64, error) {
	result := r.DB.WithContext(ctx).
		Where("parent_comment_id = ?", parentCommentId).
		Delete(&entity.Comment{})
	return result.RowsAffected, result.Error
}

func (r *CommentRepositoryImpl) IncrementReplyCount(ctx context.Context, commentId uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Comment{}).Where("id = ?", commentId).
		Update("reply_count", gorm.Expr("reply_count + 1")).Error
}

func (r *CommentRepositoryImpl) DecrementReplyCount(ctx context.Context, commentId uint, delta int) error {
	return r.DB.WithContext(ctx).Model(&entity.Comment{}).Where("id = ?", commentId).
		Update("reply_count", gorm.Expr("GREATEST(reply_count - ?, 0)", delta)).Error
}

func (r *CommentRepositoryImpl) IncrementLikeCount(ctx context.Context, commentId uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Comment{}).Where("id = ?", commentId).
		Update("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *CommentRepositoryImpl) DecrementLikeCount(ctx context.Context, commentId uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Comment{}).Where("id = ?", commentId).
		Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

func (r *CommentRepositoryImpl) IncrementDislikeCount(ctx context.Context, commentId uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Comment{}).Where("id = ?", commentId).
		Update("dislike_count", gorm.Expr("dislike_count + 1")).Error
}

func (r *CommentRepositoryImpl) DecrementDislikeCount(ctx context.Context, commentId uint) error {
	return r.DB.WithContext(ctx).Model(&entity.Comment{}).Where("id = ?", commentId).
		Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
}
