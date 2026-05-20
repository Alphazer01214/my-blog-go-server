package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/gin-gonic/gin"
)

type FileApi struct{}

func (fa *FileApi) InitUpload(c *gin.Context) {
	var req request.UploadInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	rp, err := fileService.InitUpload(cl.UserId, req.FileName, req.FileSize, req.MimeType, req.ChunkSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "upload session created")
}

func (fa *FileApi) UploadChunk(c *gin.Context) {
	uploadId := c.GetHeader("X-Upload-Id")
	if uploadId == "" {
		response.ErrorWithMsg(c, "missing X-Upload-Id header")
		return
	}

	chunkIndexStr := c.GetHeader("X-Chunk-Index")
	if chunkIndexStr == "" {
		response.ErrorWithMsg(c, "missing X-Chunk-Index header")
		return
	}
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		response.ErrorWithMsg(c, "invalid X-Chunk-Index")
		return
	}

	rp, err := fileService.UploadChunk(uploadId, chunkIndex, c.Request.Body)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "chunk uploaded")
}

func (fa *FileApi) CompleteUpload(c *gin.Context) {
	var req request.UploadCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	rp, err := fileService.CompleteUpload(cl.UserId, req.UploadId, req.Title, req.Public)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "upload completed")
}

func (fa *FileApi) GetUploadStatus(c *gin.Context) {
	uploadId := c.Query("upload_id")
	if uploadId == "" {
		response.ErrorWithMsg(c, "missing upload_id")
		return
	}

	rp, err := fileService.GetUploadStatus(uploadId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "upload status")
}

func (fa *FileApi) AbortUpload(c *gin.Context) {
	uploadId := c.Query("upload_id")
	if uploadId == "" {
		response.ErrorWithMsg(c, "missing upload_id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := fileService.AbortUpload(cl.UserId, uploadId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "upload aborted")
}

func (fa *FileApi) GetFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid id")
		return
	}

	rp, err := fileService.GetFileById(uint(id), GetUserId(c))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "query success")
}

func (fa *FileApi) DownloadFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid id")
		return
	}

	video, err := fileService.GetVideoById(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, "file not found")
		return
	}
	if !video.Public && GetUserId(c) != video.UserId {
		response.ErrorWithMsg(c, "file is private")
		return
	}

	localPath := filepath.Join(global.Config.Server.UploadDir, video.VideoSrcUrl)
	f, err := os.Open(localPath)
	if err != nil {
		response.ErrorWithMsg(c, "file not found on disk")
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		response.ErrorWithMsg(c, "file stat failed")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", video.Title))
	c.Header("Accept-Ranges", "bytes")
	http.ServeContent(c.Writer, c.Request, video.Title, info.ModTime(), f)
}

func (fa *FileApi) ListFiles(c *gin.Context) {
	page, pageSize := parsePagination(c)
	rp, err := fileService.ListFiles(page, pageSize, GetUserId(c))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "query success")
}

func (fa *FileApi) GetFilesByUserId(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}
	page, pageSize := parsePagination(c)
	rp, err := fileService.GetFilesByUserId(uint(userId), page, pageSize, GetUserId(c))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "query success")
}

func (fa *FileApi) DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := fileService.DeleteFile(uint(id), cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "delete success")
}
