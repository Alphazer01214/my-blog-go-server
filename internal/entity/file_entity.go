package entity

import (
	"time"

	"gorm.io/gorm"
)

type UploadStatus int

const (
	UploadPending   UploadStatus = iota + 1
	UploadUploading
	UploadCompleted
	UploadAborted
)

type UploadSession struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	UploadId    string       `gorm:"uniqueIndex;size:64" json:"upload_id"`
	UserId      uint         `gorm:"index" json:"user_id"`
	FileName    string       `gorm:"size:255" json:"file_name"`
	FileSize    int64        `json:"file_size"`
	ChunkSize   int          `json:"chunk_size"`
	TotalChunks int          `json:"total_chunks"`
	MimeType    string       `gorm:"size:128" json:"mime_type"`
	Status      UploadStatus `gorm:"default:1" json:"status"`
	ExpiredAt   time.Time    `json:"expired_at"`
}

type UploadedChunk struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UploadId   string    `gorm:"index;size:64" json:"upload_id"`
	ChunkIndex int       `json:"chunk_index"`
	ChunkSize  int       `json:"chunk_size"`
	Checksum   string    `gorm:"size:64" json:"checksum"`
}
