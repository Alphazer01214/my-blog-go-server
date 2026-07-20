package service

import (
	"errors"
	"time"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	pkgKafka "blog.alphazer01214.top/pkg/kafka"

	"gorm.io/gorm"
)

type SignalService struct{}

// Create 创建交易信号
func (ss *SignalService) Create(userId uint, req *request.SignalCreateRequest) (*response.SignalDetail, error) {
	signal := &entity.TradeSignal{
		UserId:      userId,
		Symbol:      req.Symbol,
		StockName:   req.StockName,
		Direction:   req.Direction,
		EntryPrice:  req.EntryPrice,
		TargetPrice: req.TargetPrice,
		StopLoss:    req.StopLoss,
		Title:       req.Title,
		Analysis:    req.Analysis,
		Tags:        req.Tags,
		Status:      "active",
	}

	if err := global.GetDB().Create(signal).Error; err != nil {
		return nil, err
	}

	// 发布 Kafka 事件
	go publishPostEvent(pkgKafka.ActionCreate, userId, signal.ID)

	return ss.toSignalDetail(signal, userId), nil
}

// GetById 获取信号详情
func (ss *SignalService) GetById(signalId, viewerId uint) (*response.SignalDetail, error) {
	signal, err := ss.getEntityById(signalId)
	if err != nil {
		return nil, err
	}

	// 增加浏览量
	global.GetDB().Model(signal).Update("view_count", gorm.Expr("view_count + 1"))

	return ss.toSignalDetail(signal, viewerId), nil
}

// List 获取信号列表
func (ss *SignalService) List(req *request.SignalListRequest, viewerId uint) (response.SignalList, error) {
	var signals []entity.TradeSignal
	var total int64

	db := global.GetDB().Model(&entity.TradeSignal{})

	// 筛选条件
	if req.Symbol != "" {
		db = db.Where("symbol = ?", req.Symbol)
	}
	if req.Direction != "" {
		db = db.Where("direction = ?", req.Direction)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return response.SignalList{}, err
	}

	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&signals).Error; err != nil {
		return response.SignalList{}, err
	}

	items := make([]response.SignalDetail, 0, len(signals))
	for _, signal := range signals {
		items = append(items, *ss.toSignalDetail(&signal, viewerId))
	}

	return response.SignalList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// ListByUserId 获取用户的信号列表
func (ss *SignalService) ListByUserId(userId uint, page, pageSize int, viewerId uint) (response.SignalList, error) {
	var signals []entity.TradeSignal
	var total int64

	db := global.GetDB().Model(&entity.TradeSignal{}).Where("user_id = ?", userId)
	if err := db.Count(&total).Error; err != nil {
		return response.SignalList{}, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&signals).Error; err != nil {
		return response.SignalList{}, err
	}

	items := make([]response.SignalDetail, 0, len(signals))
	for _, signal := range signals {
		items = append(items, *ss.toSignalDetail(&signal, viewerId))
	}

	return response.SignalList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// GetFeed 获取关注的人的信号流
func (ss *SignalService) GetFeed(userId uint, page, pageSize int) (response.SignalList, error) {
	// 获取关注的用户 ID
	var followingIds []uint
	if err := global.GetDB().Model(&entity.UserFollow{}).Where("follower_id = ?", userId).Pluck("following_id", &followingIds).Error; err != nil {
		return response.SignalList{}, err
	}

	if len(followingIds) == 0 {
		return response.SignalList{Items: []response.SignalDetail{}, Page: page, PageSize: pageSize}, nil
	}

	var signals []entity.TradeSignal
	var total int64

	db := global.GetDB().Model(&entity.TradeSignal{}).Where("user_id IN ?", followingIds)
	if err := db.Count(&total).Error; err != nil {
		return response.SignalList{}, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&signals).Error; err != nil {
		return response.SignalList{}, err
	}

	items := make([]response.SignalDetail, 0, len(signals))
	for _, signal := range signals {
		items = append(items, *ss.toSignalDetail(&signal, userId))
	}

	return response.SignalList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// CloseSignal 平仓信号
func (ss *SignalService) CloseSignal(signalId, userId uint, closedPrice float64) (*response.SignalDetail, error) {
	signal, err := ss.getEntityById(signalId)
	if err != nil {
		return nil, err
	}

	if signal.UserId != userId {
		return nil, errors.New("not the signal owner")
	}

	if signal.Status != "active" {
		return nil, errors.New("signal is not active")
	}

	now := time.Now()
	var pnl float64
	if signal.EntryPrice > 0 {
		if signal.Direction == "long" {
			pnl = (closedPrice - signal.EntryPrice) / signal.EntryPrice * 100
		} else if signal.Direction == "short" {
			pnl = (signal.EntryPrice - closedPrice) / signal.EntryPrice * 100
		}
	}

	updates := map[string]interface{}{
		"status":       "closed",
		"closed_price": closedPrice,
		"closed_at":    now,
		"pnl":          pnl,
	}

	if err := global.GetDB().Model(signal).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 通知跟单用户
	go func() {
		var follows []entity.SignalFollow
		global.GetDB().Where("signal_id = ?", signalId).Find(&follows)
		for _, follow := range follows {
			publishNotification("signal_close", userId, signalId, "signal", follow.UserId,
				"你跟单的信号已平仓")
		}
	}()

	// 重新查询
	signal, _ = ss.getEntityById(signalId)
	return ss.toSignalDetail(signal, userId), nil
}

// Follow 跟单
func (ss *SignalService) Follow(signalId, userId uint) error {
	signal, err := ss.getEntityById(signalId)
	if err != nil {
		return err
	}

	var existing entity.SignalFollow
	result := global.GetDB().Where("user_id = ? AND signal_id = ?", userId, signalId).First(&existing)
	if result.RowsAffected > 0 {
		// 已跟单，取消
		if err := global.GetDB().Delete(&existing).Error; err != nil {
			return err
		}
		global.GetDB().Model(signal).Update("follow_count", gorm.Expr("GREATEST(follow_count - 1, 0)"))
		return nil
	}

	// 新增跟单
	follow := entity.SignalFollow{
		UserId:   userId,
		SignalId: signalId,
	}
	if err := global.GetDB().Create(&follow).Error; err != nil {
		return err
	}
	global.GetDB().Model(signal).Update("follow_count", gorm.Expr("follow_count + 1"))
	return nil
}

// Delete 删除信号
func (ss *SignalService) Delete(signalId, userId uint) error {
	signal, err := ss.getEntityById(signalId)
	if err != nil {
		return err
	}
	if signal.UserId != userId {
		return errors.New("not the signal owner")
	}
	return global.GetDB().Delete(&entity.TradeSignal{}, signalId).Error
}

// getEntityById 获取信号实体
func (ss *SignalService) getEntityById(id uint) (*entity.TradeSignal, error) {
	var signal entity.TradeSignal
	err := global.GetDB().Where("id = ?", id).First(&signal).Error
	return &signal, err
}

// toSignalDetail 转换为响应结构
func (ss *SignalService) toSignalDetail(signal *entity.TradeSignal, viewerId uint) *response.SignalDetail {
	detail := &response.SignalDetail{
		ID:          signal.ID,
		CreatedAt:   signal.CreatedAt,
		UpdatedAt:   signal.UpdatedAt,
		UserId:      signal.UserId,
		Symbol:      signal.Symbol,
		StockName:   signal.StockName,
		Direction:   signal.Direction,
		EntryPrice:  signal.EntryPrice,
		TargetPrice: signal.TargetPrice,
		StopLoss:    signal.StopLoss,
		Status:      signal.Status,
		ClosedPrice: signal.ClosedPrice,
		ClosedAt:    signal.ClosedAt,
		PnL:         signal.PnL,
		Title:       signal.Title,
		Analysis:    signal.Analysis,
		Tags:        signal.Tags,
		ViewCount:   signal.ViewCount,
		LikeCount:   signal.LikeCount,
		FollowCount: signal.FollowCount,
	}

	// 获取作者信息
	author, err := Service.UserService.GetUserInfoById(signal.UserId, 0)
	if err == nil {
		detail.Author = author
	}

	// 检查是否在跟单
	if viewerId > 0 {
		var follow entity.SignalFollow
		if err := global.GetDB().Where("user_id = ? AND signal_id = ?", viewerId, signal.ID).First(&follow).Error; err == nil {
			detail.IsFollowing = true
		}
	}

	return detail
}
