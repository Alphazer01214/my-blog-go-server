package api

import (
	"strconv"

	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/service"
	"github.com/gin-gonic/gin"
)

type WatchlistApi struct{}

var watchlistService = &service.Service.WatchlistService

// Create 创建关注列表
func (wa *WatchlistApi) Create(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	var req request.WatchlistCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	result, err := watchlistService.Create(cl.UserId, &req)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "watchlist created")
}

// GetById 获取关注列表详情
func (wa *WatchlistApi) GetById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid watchlist id")
		return
	}

	viewerId := GetUserId(c)

	result, err := watchlistService.GetById(uint(id), viewerId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "success")
}

// List 获取用户的所有关注列表
func (wa *WatchlistApi) List(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}

	page, pageSize := parsePagination(c)
	viewerId := GetUserId(c)

	result, err := watchlistService.List(uint(userId), viewerId, page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "success")
}

// ListPublic 获取公开的关注列表
func (wa *WatchlistApi) ListPublic(c *gin.Context) {
	page, pageSize := parsePagination(c)

	result, err := watchlistService.ListPublic(page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "success")
}

// Update 更新关注列表
func (wa *WatchlistApi) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid watchlist id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	var req request.WatchlistUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	if err := watchlistService.Update(uint(id), cl.UserId, &req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "watchlist updated")
}

// Delete 删除关注列表
func (wa *WatchlistApi) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid watchlist id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := watchlistService.Delete(uint(id), cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "watchlist deleted")
}

// AddStock 添加股票到关注列表
func (wa *WatchlistApi) AddStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid watchlist id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	var req request.WatchlistAddStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	if err := watchlistService.AddStock(uint(id), cl.UserId, &req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "stock added")
}

// RemoveStock 从关注列表移除股票
func (wa *WatchlistApi) RemoveStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid watchlist id")
		return
	}

	symbol := c.Param("symbol")
	if symbol == "" {
		response.ErrorWithMsg(c, "symbol is required")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := watchlistService.RemoveStock(uint(id), cl.UserId, symbol); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "stock removed")
}
