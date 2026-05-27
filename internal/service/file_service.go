package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/repository"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
)

type FileService struct {
	uploadRepo  repository.UploadRepository
	videoRepo   repository.VideoRepository
	userService *UserService
}

func NewFileService(
	uploadRepo repository.UploadRepository,
	videoRepo repository.VideoRepository,
	userService *UserService,
) *FileService {
	return &FileService{
		uploadRepo:  uploadRepo,
		videoRepo:   videoRepo,
		userService: userService,
	}
}

// ======================== Upload ========================

func (fs *FileService) InitUpload(ctx context.Context, userId uint, fileName string, fileSize int64, mimeType string, chunkSize int) (*response.UploadInitResponse, error) {
	if chunkSize <= 0 {
		chunkSize = int(global.Config.Server.UploadChunkSize)
	}
	if chunkSize <= 0 {
		chunkSize = 10 * 1024 * 1024
	}

	totalChunks := int((fileSize + int64(chunkSize) - 1) / int64(chunkSize))
	uploadId := utils.GenerateUUID()

	session := &entity.UploadSession{
		UploadId:    uploadId,
		UserId:      userId,
		FileName:    fileName,
		FileSize:    fileSize,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		MimeType:    mimeType,
		Status:      entity.UploadPending,
		ExpiredAt:   time.Now().Add(24 * time.Hour),
	}
	if err := fs.uploadRepo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	sessionDir := fs.getSessionDir(uploadId)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, err
	}

	return &response.UploadInitResponse{
		UploadId:    uploadId,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
	}, nil
}

func (fs *FileService) UploadChunk(ctx context.Context, uploadId string, chunkIndex int, data io.Reader) (*response.UploadChunkResponse, error) {
	session, err := fs.uploadRepo.FindSessionByUploadId(ctx, uploadId)
	if err != nil {
		return nil, errors.New("upload session not found")
	}

	existing, _ := fs.uploadRepo.FindChunkByUploadIdAndIndex(ctx, uploadId, chunkIndex)
	if existing != nil {
		return &response.UploadChunkResponse{
			ChunkIndex: chunkIndex,
			Checksum:   existing.Checksum,
		}, nil
	}

	chunkData, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}

	checksum := md5.Sum(chunkData)
	checksumStr := hex.EncodeToString(checksum[:])

	chunkDir := fs.getSessionDir(uploadId)
	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("%d", chunkIndex))
	if err := os.WriteFile(chunkPath, chunkData, 0644); err != nil {
		return nil, err
	}

	chunk := &entity.UploadedChunk{
		UploadId:   uploadId,
		ChunkIndex: chunkIndex,
		ChunkSize:  len(chunkData),
		Checksum:   checksumStr,
	}
	if err := fs.uploadRepo.CreateChunk(ctx, chunk); err != nil {
		return nil, err
	}

	_ = fs.uploadRepo.UpdateSessionStatus(ctx, uploadId, entity.UploadUploading)

	_ = session // suppress unused warning

	return &response.UploadChunkResponse{
		ChunkIndex: chunkIndex,
		Checksum:   checksumStr,
	}, nil
}

func (fs *FileService) CompleteUpload(ctx context.Context, userId uint, uploadId string, title string, public bool) (*response.FileDetail, error) {
	session, err := fs.uploadRepo.FindSessionByUploadId(ctx, uploadId)
	if err != nil {
		return nil, err
	}

	chunkCount, err := fs.uploadRepo.CountChunksByUploadId(ctx, uploadId)
	if err != nil {
		return nil, err
	}
	if chunkCount != int64(session.TotalChunks) {
		return nil, errors.New("not all chunks uploaded")
	}

	chunks, err := fs.uploadRepo.ListChunksByUploadId(ctx, uploadId)
	if err != nil {
		return nil, err
	}

	chunkDir := fs.getSessionDir(uploadId)
	outPath := filepath.Join(constant_UploadFileBaseDir(), fmt.Sprintf("%s%s", uploadId, filepath.Ext(session.FileName)))
	outFile, err := os.Create(outPath)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

	for _, chunk := range chunks {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("%d", chunk.ChunkIndex))
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return nil, err
		}
		if _, err := outFile.Write(chunkData); err != nil {
			return nil, err
		}
	}

	video := &entity.Video{
		Title:   title,
		UserId:  userId,
		Size:    session.FileSize,
		Public:  public,
		Status:  "published",
	}
	if err := fs.videoRepo.Create(ctx, video); err != nil {
		return nil, err
	}

	_ = fs.uploadRepo.UpdateSessionStatus(ctx, uploadId, entity.UploadCompleted)

	go func() {
		os.RemoveAll(chunkDir)
	}()

	return fs.toFileDetail(video), nil
}

func (fs *FileService) GetUploadStatus(ctx context.Context, uploadId string) (*response.UploadStatusResponse, error) {
	session, err := fs.uploadRepo.FindSessionByUploadId(ctx, uploadId)
	if err != nil {
		return nil, err
	}
	chunks, err := fs.uploadRepo.ListChunksByUploadId(ctx, uploadId)
	if err != nil {
		return nil, err
	}
	uploadedIndices := make([]int, len(chunks))
	for i, c := range chunks {
		uploadedIndices[i] = c.ChunkIndex
	}
	return &response.UploadStatusResponse{
		UploadId:       uploadId,
		Status:         string(session.Status),
		UploadedCount:  len(chunks),
		TotalChunks:    session.TotalChunks,
		UploadedChunks: uploadedIndices,
	}, nil
}

func (fs *FileService) AbortUpload(ctx context.Context, userId uint, uploadId string) error {
	session, err := fs.uploadRepo.FindSessionByUploadId(ctx, uploadId)
	if err != nil {
		return err
	}
	if session.UserId != userId {
		return errors.New("unauthorized")
	}

	os.RemoveAll(fs.getSessionDir(uploadId))
	_ = fs.uploadRepo.UpdateSessionStatus(ctx, uploadId, entity.UploadAborted)
	return nil
}

// ======================== File CRUD ========================

func (fs *FileService) GetFileById(ctx context.Context, id uint, viewerId uint) (*response.FileDetail, error) {
	video, err := fs.videoRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = fs.videoRepo.IncrementViewCount(ctx, id)
	return fs.toFileDetail(video), nil
}

func (fs *FileService) ListFiles(ctx context.Context, page, pageSize int, viewerId uint) (*response.FileList, error) {
	publicOnly := viewerId == 0
	pagination, err := fs.videoRepo.ListVideos(ctx, publicOnly, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.FileDetail, len(pagination.Items))
	for i, v := range pagination.Items {
		items[i] = *fs.toFileDetail(&v)
	}
	return &response.FileList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (fs *FileService) ListFilesByUserId(ctx context.Context, userId uint, page, pageSize int, viewerId uint) (*response.FileList, error) {
	publicOnly := viewerId == 0 || viewerId != userId
	pagination, err := fs.videoRepo.ListVideosByUserId(ctx, userId, publicOnly, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.FileDetail, len(pagination.Items))
	for i, v := range pagination.Items {
		items[i] = *fs.toFileDetail(&v)
	}
	return &response.FileList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (fs *FileService) DeleteFile(ctx context.Context, id uint, userId uint) error {
	video, err := fs.videoRepo.FindById(ctx, id)
	if err != nil {
		return err
	}
	if userId != 0 && video.UserId != userId {
		return errors.New("unauthorized")
	}
	if video.VideoSrcUrl != "" {
		os.Remove(video.VideoSrcUrl)
	}
	if video.VideoCoverUrl != "" {
		os.Remove(video.VideoCoverUrl)
	}
	return fs.videoRepo.DeleteById(ctx, id)
}

// ======================== Internal ========================

func (fs *FileService) getSessionDir(uploadId string) string {
	uploadDir := global.Config.Server.UploadDir
	if uploadDir == "" {
		uploadDir = "./upload"
	}
	return filepath.Join(uploadDir, "chunk", uploadId)
}

func (fs *FileService) toFileDetail(video *entity.Video) *response.FileDetail {
	return &response.FileDetail{
		ID:            video.ID,
		CreatedAt:     video.CreatedAt,
		UpdatedAt:     video.UpdatedAt,
		UserId:        video.UserId,
		Name:          video.Title,
		Size:          video.Size,
		MimeType:      video.MimeType,
		Public:        video.Public,
		Duration:      video.Duration,
		CoverUrl:      video.VideoCoverUrl,
		ViewCount:     video.ViewCount,
		LikeCount:     video.LikeCount,
		DislikeCount:  video.DislikeCount,
		FavoriteCount: video.FavoriteCount,
		ShareCount:    video.ShareCount,
	}
}

func constant_UploadFileBaseDir() string {
	uploadDir := global.Config.Server.UploadDir
	if uploadDir == "" {
		uploadDir = "./upload"
	}
	return filepath.Join(uploadDir, "files")
}
