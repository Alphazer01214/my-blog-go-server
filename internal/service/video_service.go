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
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"

	"gorm.io/gorm"
)

type VideoService struct {
}

func (vs *VideoService) Create(ctx context.Context, userId uint, vReq request.VideoCreateRequest, videoSize int64) (entity.Video, error) {
	video := entity.Video{
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
	err := global.GetDB().Create(&video).Error
	return video, err
}

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
	err := vs.setSessionRedis(ctx, req.UploadId, session)
	if err != nil {
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

	if err := global.GetDB().Model(&entity.Video{}).Where("id = ?", video.ID).Updates(map[string]interface{}{
		"video_src_url":   streamUrl,
		"video_cover_url": coverUrl,
		"duration":        duration,
	}).Error; err != nil {
		vs.finishSessionRedis(ctx, uploadId, status)
		return err
	}

	status = constant.UploadCompleted
	vs.finishSessionRedis(ctx, uploadId, status)
	return nil
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

func (vs *VideoService) GetVideoById(id uint, viewerId uint) (*response.VideoDetail, error) {
	video, err := vs.getVideoEntityById(id)
	if err != nil {
		return nil, err
	}
	vs.visit(id)
	author, err := Service.UserService.GetUserInfoById(video.UserId, 0)
	if err != nil {
		return nil, err
	}
	return vs.toVideoDetail(video, &author, viewerId), nil
}

func (vs *VideoService) GetAllVideos(page, pageSize int, viewerId uint) (response.VideoList, error) {
	var videos []entity.Video
	var total int64
	db := global.GetDB().Model(&entity.Video{})
	if viewerId == 0 {
		db = db.Where("public = ?", true)
	}
	if err := db.Count(&total).Error; err != nil {
		return response.VideoList{}, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return response.VideoList{}, err
	}

	items := make([]response.VideoDetail, len(videos))
	for i, video := range videos {
		author, err := Service.UserService.GetUserInfoById(video.UserId, 0)
		if err != nil {
			author = response.UserInfo{}
		}
		items[i] = *vs.toVideoDetail(&video, &author, viewerId)
	}

	return response.VideoList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (vs *VideoService) GetVideosByUserId(userId uint, page, pageSize int, viewerId uint) (response.VideoList, error) {
	var videos []entity.Video
	var total int64
	db := global.GetDB().Model(&entity.Video{}).Where("user_id = ?", userId)
	if viewerId == 0 || viewerId != userId {
		db = db.Where("public = ?", true)
	}
	if err := db.Count(&total).Error; err != nil {
		return response.VideoList{}, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return response.VideoList{}, err
	}

	author, err := Service.UserService.GetUserInfoById(userId, 0)
	if err != nil {
		author = response.UserInfo{}
	}
	items := make([]response.VideoDetail, len(videos))
	for i, video := range videos {
		items[i] = *vs.toVideoDetail(&video, &author, viewerId)
	}

	return response.VideoList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (vs *VideoService) GetVideosByCategory(category string, page, pageSize int, viewerId uint) (response.VideoList, error) {
	var videos []entity.Video
	var total int64
	db := global.GetDB().Model(&entity.Video{}).Where("category = ?", category)
	if viewerId == 0 {
		db = db.Where("public = ?", true)
	}
	if err := db.Count(&total).Error; err != nil {
		return response.VideoList{}, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return response.VideoList{}, err
	}

	items := make([]response.VideoDetail, len(videos))
	for i, video := range videos {
		author, err := Service.UserService.GetUserInfoById(video.UserId, 0)
		if err != nil {
			author = response.UserInfo{}
		}
		items[i] = *vs.toVideoDetail(&video, &author, viewerId)
	}

	return response.VideoList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (vs *VideoService) SearchVideo(req *request.VideoSearchRequest, page, pageSize int, viewerId uint) (response.VideoList, error) {
	var videos []entity.Video
	var total int64
	qKeyword := "%" + req.Keyword + "%"

	db := global.GetDB().Model(&entity.Video{}).Where("title ILIKE ? OR description ILIKE ?", qKeyword, qKeyword)
	if viewerId == 0 {
		db = db.Where("public = ?", true)
	}
	if err := db.Count(&total).Error; err != nil {
		return response.VideoList{}, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return response.VideoList{}, err
	}

	items := make([]response.VideoDetail, 0, len(videos))
	for _, video := range videos {
		author, err := Service.UserService.GetUserInfoById(video.UserId, viewerId)
		if err != nil {
			continue
		}
		vd := vs.toVideoDetail(&video, &author, viewerId)
		items = append(items, *vd)
	}

	return response.VideoList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (vs *VideoService) Update(id uint, req request.VideoUpdateRequest) error {
	video, err := vs.getVideoEntityById(id)
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

	return global.GetDB().Save(video).Error
}

func (vs *VideoService) DeleteById(id uint) error {
	video, err := vs.getVideoEntityById(id)
	if err != nil {
		return err
	}

	if video.VideoSrcUrl != "" {
		os.Remove(video.VideoSrcUrl)
	}
	if video.VideoCoverUrl != "" {
		os.Remove(video.VideoCoverUrl)
	}

	return global.GetDB().Delete(&entity.Video{}, id).Error
}

func (vs *VideoService) Like(videoId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action
		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			videoId, userId, constant.TargetVideo, constant.ActionLike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Video{}).Where("id = ?", videoId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
		}

		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			videoId, userId, constant.TargetVideo, constant.ActionDislike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Video{}).Where("id = ?", videoId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:   userId,
			TargetId: videoId,
			ActType:  constant.ActionLike,
			TgtType:  constant.TargetVideo,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Video{}).Where("id = ?", videoId).
			Update("like_count", gorm.Expr("like_count + 1")).Error
	})
}

func (vs *VideoService) Dislike(videoId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action
		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			videoId, userId, constant.TargetVideo, constant.ActionDislike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Video{}).Where("id = ?", videoId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
		}

		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			videoId, userId, constant.TargetVideo, constant.ActionLike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Video{}).Where("id = ?", videoId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:   userId,
			TargetId: videoId,
			ActType:  constant.ActionDislike,
			TgtType:  constant.TargetVideo,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Video{}).Where("id = ?", videoId).
			Update("dislike_count", gorm.Expr("dislike_count + 1")).Error
	})
}

func (vs *VideoService) Favorite(videoId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action
		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			videoId, userId, constant.TargetVideo, constant.ActionFavorite).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Video{}).Where("id = ?", videoId).
				Update("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error
		}

		if err := tx.Create(&entity.Action{
			UserId:   userId,
			TargetId: videoId,
			ActType:  constant.ActionFavorite,
			TgtType:  constant.TargetVideo,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Video{}).Where("id = ?", videoId).
			Update("favorite_count", gorm.Expr("favorite_count + 1")).Error
	})
}

func (vs *VideoService) Share(videoId, userId uint, shareInfo entity.ShareInfo) error {
	extraJSON, err := json.Marshal(shareInfo)
	if err != nil {
		return err
	}
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&entity.Action{
			UserId:    userId,
			TargetId:  videoId,
			ActType:   constant.ActionShare,
			TgtType:   constant.TargetVideo,
			ExtraInfo: extraJSON,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Video{}).Where("id = ?", videoId).
			Update("share_count", gorm.Expr("share_count + 1")).Error
	})
}

func (vs *VideoService) GetVideoEntityById(id uint) (*entity.Video, error) {
	return vs.getVideoEntityById(id)
}

func (vs *VideoService) getVideoEntityById(id uint) (*entity.Video, error) {
	var video entity.Video
	err := global.GetDB().Where("id = ?", id).First(&video).Error
	return &video, err
}

func (vs *VideoService) visit(videoId uint) {
	global.GetDB().Model(&entity.Video{}).Where("id = ?", videoId).
		Update("view_count", gorm.Expr("view_count + 1"))
}

func (vs *VideoService) IsVideoForbidComment(videoId uint) bool {
	var res bool
	err := global.GetDB().Model(&entity.Video{}).Where("ID = ?", videoId).Pluck("forbid_comment", &res).Error
	return err == nil && res
}

func (vs *VideoService) toVideoDetail(video *entity.Video, author *response.UserInfo, viewerId uint) *response.VideoDetail {
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
		var like entity.Action
		if global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			video.ID, viewerId, constant.TargetVideo, constant.ActionLike).First(&like).Error == nil {
			vd.IsLiked = true
		}
		var dislike entity.Action
		if global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			video.ID, viewerId, constant.TargetVideo, constant.ActionDislike).First(&dislike).Error == nil {
			vd.IsDisliked = true
		}
		var fav entity.Action
		if global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			video.ID, viewerId, constant.TargetVideo, constant.ActionFavorite).First(&fav).Error == nil {
			vd.IsFavorited = true
		}
	}

	return vd
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
			return response.VideoUploadSessions{
				Sessions: sessions,
			}
		}
	}
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
