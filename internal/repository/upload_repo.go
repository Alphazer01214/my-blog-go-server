package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type UploadRepository interface {
	WithTx(tx *gorm.DB) UploadRepository
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	// UploadSession
	CreateSession(ctx context.Context, session *entity.UploadSession) error
	FindSessionByUploadId(ctx context.Context, uploadId string) (*entity.UploadSession, error)
	UpdateSessionStatus(ctx context.Context, uploadId string, status entity.UploadStatus) error

	// UploadedChunk
	CreateChunk(ctx context.Context, chunk *entity.UploadedChunk) error
	FindChunkByUploadIdAndIndex(ctx context.Context, uploadId string, chunkIndex int) (*entity.UploadedChunk, error)
	CountChunksByUploadId(ctx context.Context, uploadId string) (int64, error)
	ListChunksByUploadId(ctx context.Context, uploadId string) ([]entity.UploadedChunk, error)
	DeleteChunksByUploadId(ctx context.Context, uploadId string) error
}
