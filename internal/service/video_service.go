package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/repository"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"

	"gorm.io/gorm"
)

type VideoService struct {
	videoRepo   repository.VideoRepository
	actionRepo  repository.ActionRepository
	userRepo    repository.UserRepository
	userService *UserService
}

func NewVideoService(
	videoRepo repository.VideoRepository,
	actionRepo repository.ActionRepository,
	userRepo repository.UserRepository,
	userService *UserService,
) *VideoService {
	return &VideoService{
		videoRepo:   videoRepo,
		actionRepo:  actionRepo,
		userRepo:    userRepo,
		userService: userService,
	}
}

func (vs *VideoService) Create(ctx context.Context, userId uint, vReq request.VideoCreateRequest, videoSize int64) (*entity.Video, error) {
	video := &entity.Video{
		UserId:        userId,
		Title:         vReq.Title,
		Description:   vReq.Description,
		Public:        vReq.Public,
		Tags:          vReq.Tags,
		Category:      vReq.Category,
		VideoSrcUrl:   vReq.VideoSrcUrl,
		VideoCoverUrl: vReq.VideoCoverUrl,
		Size:          videoSize,
		MimeType:      "video/mp4",
		ForbidComment: vReq.ForbidComment,
		ForbidShare:   vReq.ForbidShare,
		Status:        "published",
	}
	if err := vs.videoRepo.Create(ctx, video); err != nil {
		return nil, err
	}
	return video, nil
}

func (vs *VideoService) GetVideoById(ctx context.Context, id uint, viewerId uint) (*response.VideoDetail, error) {
	video, err := vs.videoRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = vs.videoRepo.IncrementViewCount(ctx, id)
	author, _ := vs.userService.GetUserById(ctx, video.UserId, 0)
	return vs.toVideoDetail(ctx, video, author, viewerId), nil
}

func (vs *VideoService) ListVideos(ctx context.Context, page, pageSize int, viewerId uint) (*response.VideoList, error) {
	publicOnly := viewerId == 0
	pagination, err := vs.videoRepo.ListVideos(ctx, publicOnly, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.VideoDetail, len(pagination.Items))
	for i, video := range pagination.Items {
		author, _ := vs.userService.GetUserById(ctx, video.UserId, 0)
		items[i] = *vs.toVideoDetail(ctx, &video, author, viewerId)
	}
	return &response.VideoList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (vs *VideoService) ListVideosByUserId(ctx context.Context, userId uint, page, pageSize int, viewerId uint) (*response.VideoList, error) {
	publicOnly := viewerId == 0 || viewerId != userId
	pagination, err := vs.videoRepo.ListVideosByUserId(ctx, userId, publicOnly, page, pageSize)
	if err != nil {
		return nil, err
	}
	author, _ := vs.userService.GetUserById(ctx, userId, 0)
	items := make([]response.VideoDetail, len(pagination.Items))
	for i, video := range pagination.Items {
		items[i] = *vs.toVideoDetail(ctx, &video, author, viewerId)
	}
	return &response.VideoList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (vs *VideoService) ListVideosByCategory(ctx context.Context, category string, page, pageSize int, viewerId uint) (*response.VideoList, error) {
	publicOnly := viewerId == 0
	pagination, err := vs.videoRepo.ListVideosByCategory(ctx, category, publicOnly, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.VideoDetail, len(pagination.Items))
	for i, video := range pagination.Items {
		author, _ := vs.userService.GetUserById(ctx, video.UserId, 0)
		items[i] = *vs.toVideoDetail(ctx, &video, author, viewerId)
	}
	return &response.VideoList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (vs *VideoService) SearchVideos(ctx context.Context, keyword string, page, pageSize int, viewerId uint) (*response.VideoList, error) {
	publicOnly := viewerId == 0
	pagination, err := vs.videoRepo.SearchByKeyword(ctx, keyword, publicOnly, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.VideoDetail, len(pagination.Items))
	for i, video := range pagination.Items {
		author, _ := vs.userService.GetUserById(ctx, video.UserId, 0)
		items[i] = *vs.toVideoDetail(ctx, &video, author, viewerId)
	}
	return &response.VideoList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (vs *VideoService) UpdateVideo(ctx context.Context, id uint, req request.VideoUpdateRequest) error {
	video, err := vs.videoRepo.FindById(ctx, id)
	if err != nil {
		return err
	}
	if req.Title != "" {
		video.Title = req.Title
	}
	if req.Description != "" {
		video.Description = req.Description
	}
	if req.VideoCoverUrl != "" {
		video.VideoCoverUrl = req.VideoCoverUrl
	}
	if req.Category != "" {
		video.Category = req.Category
	}
	if req.Tags != nil {
		video.Tags = req.Tags
	}
	video.Public = req.Public
	video.ForbidComment = req.ForbidComment
	video.ForbidShare = req.ForbidShare
	return vs.videoRepo.Save(ctx, video)
}

func (vs *VideoService) DeleteVideoById(ctx context.Context, id uint) error {
	video, err := vs.videoRepo.FindById(ctx, id)
	if err != nil {
		return err
	}
	if video.VideoSrcUrl != "" {
		os.Remove(video.VideoSrcUrl)
	}
	if video.VideoCoverUrl != "" {
		os.Remove(video.VideoCoverUrl)
	}
	return vs.videoRepo.DeleteById(ctx, id)
}

// ======================== Actions ========================

func (vs *VideoService) ToggleLike(ctx context.Context, videoId, userId uint) error {
	return vs.videoRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := vs.actionRepo.WithTx(tx)
		txVideo := vs.videoRepo.WithTx(tx)

		exists, _ := txAction.Exists(ctx, userId, videoId, constant.ActionLike, constant.TargetVideo)
		if exists {
			txAction.Delete(ctx, userId, videoId, constant.ActionLike, constant.TargetVideo)
			return txVideo.DecrementLikeCount(ctx, videoId)
		}

		removed, _ := txAction.Delete(ctx, userId, videoId, constant.ActionDislike, constant.TargetVideo)
		if removed > 0 {
			txVideo.DecrementDislikeCount(ctx, videoId)
		}

		txAction.Create(ctx, &entity.Action{
			UserId:   userId,
			TargetId: videoId,
			ActType:  constant.ActionLike,
			TgtType:  constant.TargetVideo,
		})
		return txVideo.IncrementLikeCount(ctx, videoId)
	})
}

func (vs *VideoService) ToggleDislike(ctx context.Context, videoId, userId uint) error {
	return vs.videoRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := vs.actionRepo.WithTx(tx)
		txVideo := vs.videoRepo.WithTx(tx)

		exists, _ := txAction.Exists(ctx, userId, videoId, constant.ActionDislike, constant.TargetVideo)
		if exists {
			txAction.Delete(ctx, userId, videoId, constant.ActionDislike, constant.TargetVideo)
			return txVideo.DecrementDislikeCount(ctx, videoId)
		}

		removed, _ := txAction.Delete(ctx, userId, videoId, constant.ActionLike, constant.TargetVideo)
		if removed > 0 {
			txVideo.DecrementLikeCount(ctx, videoId)
		}

		txAction.Create(ctx, &entity.Action{
			UserId:   userId,
			TargetId: videoId,
			ActType:  constant.ActionDislike,
			TgtType:  constant.TargetVideo,
		})
		return txVideo.IncrementDislikeCount(ctx, videoId)
	})
}

func (vs *VideoService) ToggleFavorite(ctx context.Context, videoId, userId uint) error {
	return vs.videoRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := vs.actionRepo.WithTx(tx)
		txVideo := vs.videoRepo.WithTx(tx)

		exists, _ := txAction.Exists(ctx, userId, videoId, constant.ActionFavorite, constant.TargetVideo)
		if exists {
			txAction.Delete(ctx, userId, videoId, constant.ActionFavorite, constant.TargetVideo)
			return txVideo.DecrementFavoriteCount(ctx, videoId)
		}

		txAction.Create(ctx, &entity.Action{
			UserId:   userId,
			TargetId: videoId,
			ActType:  constant.ActionFavorite,
			TgtType:  constant.TargetVideo,
		})
		return txVideo.IncrementFavoriteCount(ctx, videoId)
	})
}

func (vs *VideoService) Share(ctx context.Context, videoId, userId uint, shareInfo entity.ShareInfo) error {
	extraJSON, err := json.Marshal(shareInfo)
	if err != nil {
		return err
	}
	return vs.videoRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := vs.actionRepo.WithTx(tx)
		txVideo := vs.videoRepo.WithTx(tx)

		if err := txAction.Create(ctx, &entity.Action{
			UserId:    userId,
			TargetId:  videoId,
			ActType:   constant.ActionShare,
			TgtType:   constant.TargetVideo,
			ExtraInfo: extraJSON,
		}); err != nil {
			return err
		}
		return txVideo.IncrementShareCount(ctx, videoId)
	})
}

// ======================== Upload (Redis-based) ========================

func (vs *VideoService) InitUpload(ctx context.Context, userId uint, req request.VideoInitUpload, vreq request.VideoCreateRequest) (entity.VideoUploadSession, error) {
	session := entity.VideoUploadSession{
		ExpireAt:       time.Now().Add(constant.UploadSessionExpireTime),
		UploadId:       req.UploadId,
		UserId:         userId,
		VideoName:      req.VideoName,
		VideoSize:      req.VideoSize,
		ChunkSize:      req.ChunkSize,
		TotalChunks:    (req.VideoSize + req.ChunkSize - 1) / req.ChunkSize,
		Status:         constant.UploadPending,
		UploadedChunks: make([]bool, (req.VideoSize+req.ChunkSize-1)/req.ChunkSize),
	}
	if err := vs.setSessionRedis(ctx, req.UploadId, session); err != nil {
		return entity.VideoUploadSession{}, err
	}
	if err := vs.setCreateVideoRedis(ctx, req.UploadId, vreq); err != nil {
		return entity.VideoUploadSession{}, err
	}
	return session, nil
}

func (vs *VideoService) SaveChunk(ctx context.Context, data []byte, req request.VideoChunkUpload) error {
	session, err := vs.getSessionRedis(ctx, req.UploadId)
	if err != nil {
		return errors.New("upload session not found")
	}
	if req.ChunkIndex < 0 || int64(req.ChunkIndex) >= session.TotalChunks {
		return errors.New("chunk index out of range")
	}
	dir, err := vs.getUploadChunkDir(req.UploadId)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, strconv.Itoa(req.ChunkIndex))
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	session.UploadedChunks[req.ChunkIndex] = true
	return vs.setSessionRedis(ctx, req.UploadId, session)
}

func (vs *VideoService) Merge(ctx context.Context, uploadId string) error {
	status := constant.UploadFailed
	chunkDir, err := vs.getUploadChunkDir(uploadId)
	if err != nil {
		vs.finishSessionRedis(ctx, uploadId, status)
		vs.removeCreateVideoRedis(ctx, uploadId)
		return err
	}
	if !vs.isAllChunksUploaded(ctx, uploadId) {
		return errors.New("not all chunks uploaded")
	}
	session, err := vs.getSessionRedis(ctx, uploadId)
	if err != nil {
		vs.finishSessionRedis(ctx, uploadId, status)
		return err
	}

	vReq, err := vs.getCreateVideoRedis(ctx, uploadId)
	if err != nil {
		vs.finishSessionRedis(ctx, uploadId, status)
		return err
	}

	var videoData []byte
	for i := int64(0); i < session.TotalChunks; i++ {
		chunkData, err := os.ReadFile(filepath.Join(chunkDir, strconv.Itoa(int(i))))
		if err != nil {
			vs.finishSessionRedis(ctx, uploadId, status)
			return err
		}
		videoData = append(videoData, chunkData...)
	}

	if err := os.MkdirAll(constant.UploadVideoBaseDir, os.ModePerm); err != nil {
		vs.finishSessionRedis(ctx, uploadId, status)
		return err
	}

	video, err := vs.Create(ctx, session.UserId, vReq, session.VideoSize)
	if err != nil {
		vs.finishSessionRedis(ctx, uploadId, status)
		return err
	}

	videoPath := filepath.Join(constant.UploadVideoBaseDir, fmt.Sprintf("%d.mp4", video.ID))
	if err := os.WriteFile(videoPath, videoData, 0644); err != nil {
		vs.finishSessionRedis(ctx, uploadId, status)
		return err
	}

	streamUrl := fmt.Sprintf("/api/video/%d/stream", video.ID)
	coverUrl := ""
	duration := 0

	coverPath := filepath.Join(constant.UploadVideoBaseDir, fmt.Sprintf("%d_cover.jpg", video.ID))
	if d, err := vs.extractCover(videoPath, coverPath); err == nil {
		coverUrl = fmt.Sprintf("/api/video/%d/cover", video.ID)
		duration = d
	} else {
		fmt.Printf("[video] extract cover failed for video %d: %v\n", video.ID, err)
	}

	if err := vs.videoRepo.UpdateUrls(ctx, video.ID, streamUrl, coverUrl, duration); err != nil {
		vs.finishSessionRedis(ctx, uploadId, status)
		return err
	}

	status = constant.UploadCompleted
	vs.finishSessionRedis(ctx, uploadId, status)
	return nil
}

func (vs *VideoService) GetAllVideoUploadSessions(ctx context.Context, maxQuery int64) response.VideoUploadSessions {
	var cursor uint64
	var sessions []entity.VideoUploadSession
	for {
		keys, nextCursor, err := global.GetRedis().Scan(ctx, cursor, "upload_id:*", maxQuery).Result()
		if err != nil {
			return response.VideoUploadSessions{}
		}
		for _, key := range keys {
			data := global.GetRedis().Get(ctx, key).Val()
			var session entity.VideoUploadSession
			err := json.Unmarshal([]byte(data), &session)
			if err != nil {
				continue
			}
			sessions = append(sessions, session)
		}
		cursor = nextCursor
		if cursor == 0 {
			return response.VideoUploadSessions{Sessions: sessions}
		}
	}
}

// ======================== Internal ========================

func (vs *VideoService) toVideoDetail(ctx context.Context, video *entity.Video, author *response.UserInfo, viewerId uint) *response.VideoDetail {
	vd := &response.VideoDetail{
		ID:            video.ID,
		CreatedAt:     video.CreatedAt,
		UpdatedAt:     video.UpdatedAt,
		Title:         video.Title,
		Description:   video.Description,
		VideoSrcUrl:   video.VideoSrcUrl,
		VideoCoverUrl: video.VideoCoverUrl,
		Duration:      video.Duration,
		Size:          video.Size,
		MimeType:      video.MimeType,
		UserId:        video.UserId,
		Category:      video.Category,
		Tags:          video.Tags,
		ViewCount:     video.ViewCount,
		LikeCount:     video.LikeCount,
		DislikeCount:  video.DislikeCount,
		CommentCount:  video.CommentCount,
		FavoriteCount: video.FavoriteCount,
		ShareCount:    video.ShareCount,
		Public:        video.Public,
		ForbidComment: video.ForbidComment,
		ForbidShare:   video.ForbidShare,
		Status:        video.Status,
	}
	if author != nil {
		vd.Author = *author
	}
	if viewerId > 0 {
		if exists, _ := vs.actionRepo.Exists(ctx, viewerId, video.ID, constant.ActionLike, constant.TargetVideo); exists {
			vd.IsLiked = true
		}
		if exists, _ := vs.actionRepo.Exists(ctx, viewerId, video.ID, constant.ActionDislike, constant.TargetVideo); exists {
			vd.IsDisliked = true
		}
		if exists, _ := vs.actionRepo.Exists(ctx, viewerId, video.ID, constant.ActionFavorite, constant.TargetVideo); exists {
			vd.IsFavorited = true
		}
	}
	return vd
}

func (vs *VideoService) extractCover(videoPath string, coverPath string) (int, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return 0, fmt.Errorf("ffmpeg not found")
	}
	duration, err := vs.getVideoDuration(videoPath)
	if err != nil {
		return 0, err
	}
	if duration <= 0 {
		return 0, fmt.Errorf("invalid video duration")
	}
	seekTime := 1.0
	if duration > 3 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		minSec := float64(duration) * 0.1
		maxSec := float64(duration) * 0.9
		seekTime = minSec + rng.Float64()*(maxSec-minSec)
	}
	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%.2f", seekTime),
		"-i", videoPath,
		"-vframes", "1",
		"-q:v", "2",
		"-y",
		coverPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffmpeg failed: %v, output: %s", err, string(output))
	}
	return duration, nil
}

func (vs *VideoService) getVideoDuration(videoPath string) (int, error) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return 0, fmt.Errorf("ffprobe not found")
	}
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %v", err)
	}
	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration failed: %v", err)
	}
	return int(durationFloat), nil
}

func (vs *VideoService) getUploadChunkDir(uploadId string) (string, error) {
	b := constant.UploadChunkBaseDir
	chunkDir := filepath.Join(b, uploadId)
	err := os.MkdirAll(chunkDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	return chunkDir, nil
}

func (vs *VideoService) setSessionRedis(ctx context.Context, uploadId string, session entity.VideoUploadSession) error {
	s, err := json.Marshal(session)
	if err != nil {
		return err
	}
	global.GetRedis().Set(ctx, "upload_session:"+uploadId, s, constant.UploadSessionExpireTime)
	return nil
}

func (vs *VideoService) getSessionRedis(ctx context.Context, uploadId string) (entity.VideoUploadSession, error) {
	data := global.GetRedis().Get(ctx, "upload_session:"+uploadId).Val()
	var session entity.VideoUploadSession
	err := json.Unmarshal([]byte(data), &session)
	if err != nil {
		return entity.VideoUploadSession{}, err
	}
	return session, nil
}

func (vs *VideoService) finishSessionRedis(ctx context.Context, uploadId string, status constant.UploadStatus) {
	session, _ := vs.getSessionRedis(ctx, uploadId)
	session.Status = status
	global.GetRedis().Set(ctx, "upload_session:"+uploadId, session, 1145*time.Second)
	vs.removeCreateVideoRedis(ctx, uploadId)
}

func (vs *VideoService) isAllChunksUploaded(ctx context.Context, uploadId string) bool {
	session, err := vs.getSessionRedis(ctx, uploadId)
	if err != nil {
		return false
	}
	for _, uploaded := range session.UploadedChunks {
		if !uploaded {
			return false
		}
	}
	return true
}

func (vs *VideoService) setCreateVideoRedis(ctx context.Context, uploadId string, req request.VideoCreateRequest) error {
	s, err := json.Marshal(req)
	if err != nil {
		return err
	}
	global.GetRedis().Set(ctx, "create_video:"+uploadId, s, 1145*time.Second)
	return nil
}

func (vs *VideoService) getCreateVideoRedis(ctx context.Context, uploadId string) (request.VideoCreateRequest, error) {
	data := global.GetRedis().Get(ctx, "create_video:"+uploadId).Val()
	var req request.VideoCreateRequest
	err := json.Unmarshal([]byte(data), &req)
	if err != nil {
		return request.VideoCreateRequest{}, err
	}
	return req, nil
}

func (vs *VideoService) removeCreateVideoRedis(ctx context.Context, uploadId string) {
	global.GetRedis().Del(ctx, "create_video:"+uploadId)
}
