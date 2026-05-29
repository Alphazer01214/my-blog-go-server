package service

import (
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
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
	"gorm.io/gorm"
)

type FileService struct{}

func (fs *FileService) InitUpload(userId uint, fileName string, fileSize int64, mimeType string, chunkSize int) (*response.UploadInitResponse, error) {
	if chunkSize <= 0 {
		chunkSize = int(global.Config.Server.UploadChunkSize)
	}
	if chunkSize <= 0 {
		chunkSize = 10 * 1024 * 1024 // 10MB default
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
	if err := global.GetDB().Create(session).Error; err != nil {
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

func (fs *FileService) UploadChunk(uploadId string, chunkIndex int, data io.Reader) (*response.UploadChunkResponse, error) {
	var session entity.UploadSession
	if err := global.GetDB().Where("upload_id = ?", uploadId).First(&session).Error; err != nil {
		return nil, errors.New("upload session not found")
	}
	if session.Status == entity.UploadCompleted {
		return nil, errors.New("upload already completed")
	}
	if session.Status == entity.UploadAborted {
		return nil, errors.New("upload was aborted")
	}
	if time.Now().After(session.ExpiredAt) {
		return nil, errors.New("upload session expired")
	}
	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return nil, errors.New("invalid chunk index")
	}

	var existing entity.UploadedChunk
	result := global.GetDB().Where("upload_id = ? AND chunk_index = ?", uploadId, chunkIndex).First(&existing)
	if result.RowsAffected > 0 {
		hash := md5.New()
		io.Copy(hash, data)
		return &response.UploadChunkResponse{
			ChunkIndex: chunkIndex,
			Checksum:   hex.EncodeToString(hash.Sum(nil)),
		}, nil
	}

	chunkPath := fs.getChunkPath(uploadId, chunkIndex)
	if err := os.MkdirAll(filepath.Dir(chunkPath), 0755); err != nil {
		return nil, err
	}

	file, err := os.Create(chunkPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := md5.New()
	written, err := io.Copy(file, io.TeeReader(data, hash))
	if err != nil {
		os.Remove(chunkPath)
		return nil, err
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	chunk := &entity.UploadedChunk{
		UploadId:   uploadId,
		ChunkIndex: chunkIndex,
		ChunkSize:  int(written),
		Checksum:   checksum,
	}
	if err := global.GetDB().Create(chunk).Error; err != nil {
		os.Remove(chunkPath)
		return nil, err
	}

	if session.Status == entity.UploadPending {
		global.GetDB().Model(&session).Update("status", entity.UploadUploading)
	}

	return &response.UploadChunkResponse{
		ChunkIndex: chunkIndex,
		Checksum:   checksum,
	}, nil
}

func (fs *FileService) CompleteUpload(userId uint, uploadId string, title string, public bool) (*response.FileDetail, error) {
	var session entity.UploadSession
	if err := global.GetDB().Where("upload_id = ?", uploadId).First(&session).Error; err != nil {
		return nil, errors.New("upload session not found")
	}
	if session.UserId != userId {
		return nil, errors.New("unauthorized")
	}

	var count int64
	global.GetDB().Model(&entity.UploadedChunk{}).Where("upload_id = ?", uploadId).Count(&count)
	if int(count) < session.TotalChunks {
		return nil, fmt.Errorf("missing chunks: uploaded %d, expected %d", count, session.TotalChunks)
	}

	destDir := filepath.Join(global.Config.Server.UploadDir, "files")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, err
	}

	ext := filepath.Ext(session.FileName)
	finalName := utils.GenerateUUID() + ext
	finalPath := filepath.Join(destDir, finalName)

	destFile, err := os.Create(finalPath)
	if err != nil {
		return nil, err
	}
	defer destFile.Close()

	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := fs.getChunkPath(uploadId, i)
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			os.Remove(finalPath)
			return nil, fmt.Errorf("failed to open chunk %d: %v", i, err)
		}
		io.Copy(destFile, chunkFile)
		chunkFile.Close()
		os.Remove(chunkPath)
	}

	os.RemoveAll(fs.getSessionDir(uploadId))

	video := &entity.Video{
		UserId:      userId,
		Title:       title,
		Description: "",
		VideoSrcUrl: "/uploads/files/" + finalName,
		Size:        session.FileSize,
		MimeType:    session.MimeType,
		Public:      public,
		Type:        "file",
	}
	if err := global.GetDB().Create(video).Error; err != nil {
		os.Remove(finalPath)
		return nil, err
	}

	global.GetDB().Where("upload_id = ?", uploadId).Delete(&entity.UploadedChunk{})
	global.GetDB().Model(&session).Update("status", entity.UploadCompleted)

	return fs.toFileDetail(video), nil
}

func (fs *FileService) GetUploadStatus(uploadId string) (*response.UploadStatusResponse, error) {
	var session entity.UploadSession
	if err := global.GetDB().Where("upload_id = ?", uploadId).First(&session).Error; err != nil {
		return nil, errors.New("upload session not found")
	}

	var chunks []entity.UploadedChunk
	global.GetDB().Where("upload_id = ?", uploadId).Find(&chunks)

	uploadedIndices := make([]int, len(chunks))
	for i, c := range chunks {
		uploadedIndices[i] = c.ChunkIndex
	}

	statusStr := "pending"
	switch session.Status {
	case entity.UploadUploading:
		statusStr = "uploading"
	case entity.UploadCompleted:
		statusStr = "completed"
	case entity.UploadAborted:
		statusStr = "aborted"
	}

	return &response.UploadStatusResponse{
		UploadId:       uploadId,
		Status:         statusStr,
		UploadedCount:  len(chunks),
		TotalChunks:    session.TotalChunks,
		UploadedChunks: uploadedIndices,
	}, nil
}

func (fs *FileService) AbortUpload(userId uint, uploadId string) error {
	var session entity.UploadSession
	if err := global.GetDB().Where("upload_id = ?", uploadId).First(&session).Error; err != nil {
		return errors.New("upload session not found")
	}
	if session.UserId != userId {
		return errors.New("unauthorized")
	}

	os.RemoveAll(fs.getSessionDir(uploadId))
	global.GetDB().Where("upload_id = ?", uploadId).Delete(&entity.UploadedChunk{})
	global.GetDB().Model(&session).Update("status", entity.UploadAborted)
	return nil
}

func (fs *FileService) GetFileById(id uint, viewerId uint) (*response.FileDetail, error) {
	var video entity.Video
	if err := global.GetDB().Where("id = ?", id).First(&video).Error; err != nil {
		return nil, err
	}
	if !video.Public && video.UserId != viewerId {
		return nil, errors.New("file is private")
	}
	global.GetDB().Model(&entity.Video{}).Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1"))
	return fs.toFileDetail(&video), nil
}

func (fs *FileService) GetVideoById(id uint) (*entity.Video, error) {
	var video entity.Video
	if err := global.GetDB().Where("id = ?", id).First(&video).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

func (fs *FileService) ListFiles(page, pageSize int, viewerId uint) (response.FileList, error) {
	var videos []entity.Video
	var total int64
	db := global.GetDB().Model(&entity.Video{}).Where("type = ?", "file")
	if viewerId == 0 {
		db = db.Where("public = ?", true)
	}
	if err := db.Count(&total).Error; err != nil {
		return response.FileList{}, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return response.FileList{}, err
	}

	items := make([]response.FileDetail, len(videos))
	for i, v := range videos {
		items[i] = *fs.toFileDetail(&v)
	}

	return response.FileList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (fs *FileService) GetFilesByUserId(userId uint, page, pageSize int, viewerId uint) (response.FileList, error) {
	var videos []entity.Video
	var total int64
	db := global.GetDB().Model(&entity.Video{}).Where("user_id = ? AND type = ?", userId, "file")
	if viewerId != 0 && viewerId != userId {
		db = db.Where("public = ?", true)
	}
	if err := db.Count(&total).Error; err != nil {
		return response.FileList{}, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return response.FileList{}, err
	}

	items := make([]response.FileDetail, len(videos))
	for i, v := range videos {
		items[i] = *fs.toFileDetail(&v)
	}

	return response.FileList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (fs *FileService) DeleteFile(id uint, userId uint) error {
	var video entity.Video
	if err := global.GetDB().Where("id = ?", id).First(&video).Error; err != nil {
		return err
	}
	if video.UserId != userId {
		return errors.New("you can't delete others' files")
	}

	if video.VideoSrcUrl != "" {
		localPath := filepath.Join(global.Config.Server.UploadDir, video.VideoSrcUrl)
		os.Remove(localPath)
	}
	if video.VideoCoverUrl != "" {
		localPath := filepath.Join(global.Config.Server.UploadDir, video.VideoCoverUrl)
		os.Remove(localPath)
	}

	return global.GetDB().Delete(&video).Error
}

func (fs *FileService) toFileDetail(video *entity.Video) *response.FileDetail {
	return &response.FileDetail{
		ID:           video.ID,
		CreatedAt:    video.CreatedAt,
		UpdatedAt:    video.UpdatedAt,
		UserId:       video.UserId,
		Name:         video.Title,
		Path:         video.VideoSrcUrl,
		Size:         video.Size,
		MimeType:     video.MimeType,
		Public:       video.Public,
		Duration:     video.Duration,
		Width:        0,
		Height:       0,
		CoverUrl:     video.VideoCoverUrl,
		ViewCount:    video.ViewCount,
		LikeCount:    video.LikeCount,
		DislikeCount: video.DislikeCount,
		FavoriteCount: video.FavoriteCount,
		ShareCount:   video.ShareCount,
	}
}

func (fs *FileService) getSessionDir(uploadId string) string {
	return filepath.Join(global.Config.Server.UploadDir, "sessions", uploadId)
}

func (fs *FileService) getChunkPath(uploadId string, chunkIndex int) string {
	return filepath.Join(fs.getSessionDir(uploadId), fmt.Sprintf("%d.chunk", chunkIndex))
}
