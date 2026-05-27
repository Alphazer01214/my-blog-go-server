package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type UploadRepositoryImpl struct {
	*BaseRepository
}

func NewUploadRepository(db *gorm.DB) *UploadRepositoryImpl {
	return &UploadRepositoryImpl{BaseRepository: NewBaseRepository(db)}
}

func (r *UploadRepositoryImpl) WithTx(tx *gorm.DB) UploadRepository {
	return &UploadRepositoryImpl{BaseRepository: &BaseRepository{DB: tx}}
}

func (r *UploadRepositoryImpl) CreateSession(ctx context.Context, session *entity.UploadSession) error {
	return r.DB.WithContext(ctx).Create(session).Error
}

func (r *UploadRepositoryImpl) FindSessionByUploadId(ctx context.Context, uploadId string) (*entity.UploadSession, error) {
	var session entity.UploadSession
	err := r.DB.WithContext(ctx).Where("upload_id = ?", uploadId).First(&session).Error
	return &session, err
}

func (r *UploadRepositoryImpl) UpdateSessionStatus(ctx context.Context, uploadId string, status entity.UploadStatus) error {
	return r.DB.WithContext(ctx).
		Model(&entity.UploadSession{}).
		Where("upload_id = ?", uploadId).
		Update("status", status).Error
}

func (r *UploadRepositoryImpl) CreateChunk(ctx context.Context, chunk *entity.UploadedChunk) error {
	return r.DB.WithContext(ctx).Create(chunk).Error
}

func (r *UploadRepositoryImpl) FindChunkByUploadIdAndIndex(ctx context.Context, uploadId string, chunkIndex int) (*entity.UploadedChunk, error) {
	var chunk entity.UploadedChunk
	err := r.DB.WithContext(ctx).
		Where("upload_id = ? AND chunk_index = ?", uploadId, chunkIndex).
		First(&chunk).Error
	return &chunk, err
}

func (r *UploadRepositoryImpl) CountChunksByUploadId(ctx context.Context, uploadId string) (int64, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&entity.UploadedChunk{}).
		Where("upload_id = ?", uploadId).
		Count(&count).Error
	return count, err
}

func (r *UploadRepositoryImpl) ListChunksByUploadId(ctx context.Context, uploadId string) ([]entity.UploadedChunk, error) {
	var chunks []entity.UploadedChunk
	err := r.DB.WithContext(ctx).
		Where("upload_id = ?", uploadId).
		Order("chunk_index asc").
		Find(&chunks).Error
	return chunks, err
}

func (r *UploadRepositoryImpl) DeleteChunksByUploadId(ctx context.Context, uploadId string) error {
	return r.DB.WithContext(ctx).
		Where("upload_id = ?", uploadId).
		Delete(&entity.UploadedChunk{}).Error
}
