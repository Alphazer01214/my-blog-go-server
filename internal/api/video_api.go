package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/gin-gonic/gin"
)

type VideoApi struct {
}

func (va *VideoApi) InitUpload(c *gin.Context) {
	ctx := c.Request.Context()
	var req request.VideoInitUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, "invalid token")
		return
	}
	userId := cl.UserId

	initReq := request.VideoInitUpload{
		UploadId:  req.UploadId,
		VideoName: req.VideoName,
		VideoSize: req.VideoSize,
		ChunkSize: req.ChunkSize,
	}
	vReq := request.VideoCreateRequest{
		Env:           req.Env,
		Title:         req.Title,
		Description:   req.Description,
		VideoSrcUrl:   req.VideoSrcUrl,
		VideoCoverUrl: req.VideoCoverUrl,
		Category:      req.Category,
		Tags:          req.Tags,
		Public:        req.Public,
		ForbidComment: req.ForbidComment,
		ForbidShare:   req.ForbidShare,
	}

	session, err := videoService.InitUpload(ctx, userId, initReq, vReq)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, session, "upload session created")
}

func (va *VideoApi) UploadChunk(c *gin.Context) {
	ctx := c.Request.Context()

	uploadId := c.PostForm("upload_id")
	chunkIndexStr := c.PostForm("chunk_index")
	chunkHash := c.PostForm("chunk_hash")

	if uploadId == "" || chunkIndexStr == "" || chunkHash == "" {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		response.ErrorWithMsg(c, "invalid chunk_index")
		return
	}

	req := request.VideoChunkUpload{
		UploadId:   uploadId,
		ChunkIndex: chunkIndex,
		ChunkHash:  chunkHash,
	}

	f, err := c.FormFile("file")
	if err != nil {
		response.ErrorWithMsg(c, "file error")
		return
	}

	maxRetries := 3
	for retry := 0; retry < maxRetries; retry++ {
		chunkFile, err := f.Open()
		if err != nil {
			if retry < maxRetries-1 {
				time.Sleep(time.Duration(retry+1) * 500 * time.Millisecond)
				continue
			}
			response.ErrorWithMsg(c, "can't open file")
			return
		}

		data, err := io.ReadAll(chunkFile)
		chunkFile.Close()
		if err != nil {
			if retry < maxRetries-1 {
				time.Sleep(time.Duration(retry+1) * 500 * time.Millisecond)
				continue
			}
			response.ErrorWithMsg(c, "can't read file")
			return
		}

		hash := sha256.Sum256(data)
		sha := fmt.Sprintf("%x", hash)
		if sha != chunkHash {
			if retry < maxRetries-1 {
				time.Sleep(time.Duration(retry+1) * 500 * time.Millisecond)
				continue
			}
			response.ErrorWithMsg(c, "chunk md5 not match")
			return
		}

		if err := videoService.SaveChunk(ctx, data, req); err == nil {
			response.SuccessWithMsg(c, "upload "+strconv.Itoa(req.ChunkIndex)+" success")
			return
		}

		if retry < maxRetries-1 {
			time.Sleep(time.Duration(retry+1) * 500 * time.Millisecond)
		}
	}

	response.ErrorWithMsg(c, "failed after "+strconv.Itoa(maxRetries)+" retries")
}

func (va *VideoApi) Merge(c *gin.Context) {
	ctx := c.Request.Context()
	var req request.VideoMergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	if err := videoService.Merge(ctx, req.UploadId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "video upload completed")
}

func (va *VideoApi) Stream(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid video id")
		return
	}

	video, err := videoService.GetVideoEntityById(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, "video not found")
		return
	}
	if !video.Public && GetUserId(c) != video.UserId {
		response.ErrorWithMsg(c, "video is private")
		return
	}

	videoPath := filepath.Join(constant.UploadVideoBaseDir, fmt.Sprintf("%d.mp4", video.ID))
	f, err := os.Open(videoPath)
	if err != nil {
		response.ErrorWithMsg(c, "video file not found")
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		response.ErrorWithMsg(c, "file stat failed")
		return
	}

	c.Header("Content-Type", video.MimeType)
	c.Header("Accept-Ranges", "bytes")
	http.ServeContent(c.Writer, c.Request, video.Title, info.ModTime(), f)
}

func (va *VideoApi) Cover(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid video id")
		return
	}

	coverPath := filepath.Join(constant.UploadVideoBaseDir, fmt.Sprintf("%d_cover.jpg", id))
	f, err := os.Open(coverPath)
	if err != nil {
		response.ErrorWithMsg(c, "cover not found")
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		response.ErrorWithMsg(c, "cover stat failed")
		return
	}

	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=86400")
	http.ServeContent(c.Writer, c.Request, fmt.Sprintf("cover_%d.jpg", id), info.ModTime(), f)
}

func (va *VideoApi) Create(c *gin.Context) {
	var req request.VideoCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if _, err := videoService.Create(c.Request.Context(), cl.UserId, req, 0); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "video create success")
}

func (va *VideoApi) Query(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid video id")
		return
	}

	vd, err := videoService.GetVideoById(uint(id), GetUserId(c))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, vd, "query success")
}

func (va *VideoApi) QueryAll(c *gin.Context) {
	category := c.Query("category")
	keyword := c.Query("keyword")
	page, pageSize := parsePagination(c)

	var list response.VideoList
	var err error

	if keyword != "" {
		list, err = videoService.SearchVideo(&request.VideoSearchRequest{
			Keyword: keyword,
		}, page, pageSize, GetUserId(c))
	} else if category != "" {
		list, err = videoService.GetVideosByCategory(category, page, pageSize, GetUserId(c))
	} else {
		list, err = videoService.GetAllVideos(page, pageSize, GetUserId(c))
	}

	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, list, "query success")
}

func (va *VideoApi) ListByUserId(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}
	page, pageSize := parsePagination(c)

	list, err := videoService.GetVideosByUserId(uint(userId), page, pageSize, GetUserId(c))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, list, "query success")
}

func (va *VideoApi) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid video id")
		return
	}

	var req request.VideoUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}
	req.VideoId = uint(id)

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	video, err := videoService.GetVideoEntityById(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, "video not found")
		return
	}
	if video.UserId != cl.UserId {
		response.ErrorWithMsg(c, "unauthorized")
		return
	}

	if err := videoService.Update(uint(id), req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "update success")
}

func (va *VideoApi) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid video id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	video, err := videoService.GetVideoEntityById(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, "video not found")
		return
	}
	if video.UserId != cl.UserId {
		response.ErrorWithMsg(c, "unauthorized")
		return
	}

	if err := videoService.DeleteById(uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "delete success")
}

func (va *VideoApi) Like(c *gin.Context) {
	var req request.VideoActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := videoService.Like(req.VideoId, cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "like success")
}

func (va *VideoApi) Dislike(c *gin.Context) {
	var req request.VideoActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := videoService.Dislike(req.VideoId, cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "dislike success")
}

func (va *VideoApi) Favorite(c *gin.Context) {
	var req request.VideoActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := videoService.Favorite(req.VideoId, cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "favorite success")
}

func (va *VideoApi) Share(c *gin.Context) {
	var req request.VideoShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	shareInfo := entity.ShareInfo{
		ShareTo:      req.ShareTo,
		ShareMessage: req.ShareMessage,
	}
	infoJSON, _ := json.Marshal(shareInfo)

	var shareInfoEntity entity.ShareInfo
	err = json.Unmarshal(infoJSON, &shareInfoEntity)
	if err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	if err := videoService.Share(req.VideoId, cl.UserId, shareInfoEntity); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "share success")
}
