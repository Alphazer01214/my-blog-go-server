package request

type UploadInitRequest struct {
	FileName   string `json:"file_name" binding:"required"`
	FileSize   int64  `json:"file_size" binding:"required"`
	MimeType   string `json:"mime_type" binding:"required"`
	ChunkSize  int    `json:"chunk_size"`
}

type UploadChunkRequest struct {
	UploadId   string `form:"upload_id" binding:"required"`
	ChunkIndex int    `form:"chunk_index" binding:"required"`
}

type UploadCompleteRequest struct {
	UploadId string `json:"upload_id" binding:"required"`
	Title    string `json:"title"`
	Public   bool   `json:"public"`
}

type FileListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
