package api

import (
	"fmt"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StockApi struct{}

// GetBoard 获取股票讨论帖列表
func (sa *StockApi) GetBoard(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		response.ErrorWithMsg(c, "symbol is required")
		return
	}

	page, pageSize := parsePagination(c)
	viewerId := GetUserId(c)

	// 使用 category 过滤
	category := fmt.Sprintf("stock:%s", symbol)

	var posts []entity.Post
	var total int64

	db := globalGetDB().Model(&entity.Post{}).Where("category = ?", category)
	if err := db.Count(&total).Error; err != nil {
		response.ErrorWithMsg(c, "query failed")
		return
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		response.ErrorWithMsg(c, "query failed")
		return
	}

	postService := &service.Service.PostService
	items := make([]response.PostDetail, 0, len(posts))
	for _, post := range posts {
		author, _ := service.Service.UserService.GetUserInfoById(post.UserId, 0)
		pd := postService.ConvertToDetail(&post, &author, viewerId)
		items = append(items, *pd)
	}

	response.SuccessWithDetail(c, response.PostList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, "success")
}

// CreatePost 在股票讨论区发帖
func (sa *StockApi) CreatePost(c *gin.Context) {
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

	var req request.PostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	// 强制设置分类
	req.Category = fmt.Sprintf("stock:%s", symbol)

	post := &entity.Post{
		EnvInfo:       req.Env,
		UserId:        cl.UserId,
		Title:         req.Title,
		Cover:         req.Cover,
		Tags:          req.Tags,
		Category:      req.Category,
		Keywords:      req.Keywords,
		Content:       req.Content,
		Public:        req.Public,
		ForbidComment: req.ForbidComment,
		ForbidShare:   req.ForbidShare,
	}

	result, err := service.Service.PostService.Create(post)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, result, "post created")
}

// GetInfo 获取股票基本信息
func (sa *StockApi) GetInfo(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		response.ErrorWithMsg(c, "symbol is required")
		return
	}

	// 返回股票代码信息，前端可结合行情接口获取实时数据
	response.SuccessWithDetail(c, gin.H{
		"symbol":    symbol,
		"board_url": fmt.Sprintf("/api/stock/%s/board", symbol),
	}, "success")
}

// 辅助函数
func globalGetDB() *gorm.DB {
	return global.GetDB()
}
