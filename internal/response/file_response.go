package response

import "time"

type UploadInitResponse struct {
	UploadId    string `json:"upload_id"`
	ChunkSize   int    `json:"chunk_size"`
	TotalChunks int    `json:"total_chunks"`
}

type UploadStatusResponse struct {
	UploadId      string `json:"upload_id"`
	Status        string `json:"status"`
	UploadedCount int    `json:"uploaded_count"`
	TotalChunks   int    `json:"total_chunks"`
	UploadedChunks []int `json:"uploaded_chunks"`
}

type UploadChunkResponse struct {
	ChunkIndex int    `json:"chunk_index"`
	Checksum   string `json:"checksum"`
}

type FileDetail struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserId      uint      `json:"user_id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mime_type"`
	Public      bool      `json:"public"`
	Duration    int       `json:"duration"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	CoverUrl    string    `json:"cover_url"`
	ViewCount   int       `json:"view_count"`
	LikeCount   int       `json:"like_count"`
	DislikeCount int      `json:"dislike_count"`
	FavoriteCount int     `json:"favorite_count"`
	ShareCount  int       `json:"share_count"`
}

type FileList struct {
	Items    []FileDetail `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int64        `json:"total"`
}
