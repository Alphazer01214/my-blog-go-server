package api

import (
	"strconv"

	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/service"
	"github.com/gin-gonic/gin"
)

type SignalApi struct{}

var signalService = &service.Service.SignalService

// Create 创建交易信号
func (sa *SignalApi) Create(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	var req request.SignalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	result, err := signalService.Create(cl.UserId, &req)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "signal created")
}

// GetById 获取信号详情
func (sa *SignalApi) GetById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid signal id")
		return
	}

	viewerId := GetUserId(c)

	result, err := signalService.GetById(uint(id), viewerId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "success")
}

// List 获取信号列表
func (sa *SignalApi) List(c *gin.Context) {
	var req request.SignalListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	viewerId := GetUserId(c)

	result, err := signalService.List(&req, viewerId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "success")
}

// ListByUserId 获取用户的信号列表
func (sa *SignalApi) ListByUserId(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}

	page, pageSize := parsePagination(c)
	viewerId := GetUserId(c)

	result, err := signalService.ListByUserId(uint(userId), page, pageSize, viewerId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "success")
}

// GetFeed 获取关注的人的信号流
func (sa *SignalApi) GetFeed(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	page, pageSize := parsePagination(c)

	result, err := signalService.GetFeed(cl.UserId, page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "success")
}

// CloseSignal 平仓信号
func (sa *SignalApi) CloseSignal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid signal id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	var req struct {
		ClosedPrice float64 `json:"closed_price" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	result, err := signalService.CloseSignal(uint(id), cl.UserId, req.ClosedPrice)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "signal closed")
}

// Follow 跟单/取消跟单
func (sa *SignalApi) Follow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid signal id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := signalService.Follow(uint(id), cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "success")
}

// Delete 删除信号
func (sa *SignalApi) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid signal id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := signalService.Delete(uint(id), cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "signal deleted")
}
