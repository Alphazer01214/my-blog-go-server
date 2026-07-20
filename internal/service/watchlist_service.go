package service

import (
	"errors"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
)

type WatchlistService struct{}

// Create 创建关注列表
func (ws *WatchlistService) Create(userId uint, req *request.WatchlistCreateRequest) (*response.WatchlistDetail, error) {
	watchlist := &entity.Watchlist{
		UserId:      userId,
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}

	if err := global.GetDB().Create(watchlist).Error; err != nil {
		return nil, err
	}

	return ws.toWatchlistDetail(watchlist, nil), nil
}

// GetById 获取关注列表详情
func (ws *WatchlistService) GetById(watchlistId, viewerId uint) (*response.WatchlistDetail, error) {
	watchlist, err := ws.getEntityById(watchlistId)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if !watchlist.IsPublic && watchlist.UserId != viewerId {
		return nil, errors.New("this watchlist is private")
	}

	// 获取列表项
	var items []entity.WatchlistItem
	global.GetDB().Where("watchlist_id = ?", watchlistId).Order("sort_order asc").Find(&items)

	return ws.toWatchlistDetail(watchlist, items), nil
}

// List 获取用户的所有关注列表
func (ws *WatchlistService) List(userId, viewerId uint, page, pageSize int) (response.WatchlistList, error) {
	var watchlists []entity.Watchlist
	var total int64

	db := global.GetDB().Model(&entity.Watchlist{}).Where("user_id = ?", userId)

	// 非本人只能看公开列表
	if userId != viewerId {
		db = db.Where("is_public = ?", true)
	}

	if err := db.Count(&total).Error; err != nil {
		return response.WatchlistList{}, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&watchlists).Error; err != nil {
		return response.WatchlistList{}, err
	}

	items := make([]response.WatchlistDetail, 0, len(watchlists))
	for _, wl := range watchlists {
		items = append(items, *ws.toWatchlistDetail(&wl, nil))
	}

	return response.WatchlistList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// ListPublic 获取公开的关注列表
func (ws *WatchlistService) ListPublic(page, pageSize int) (response.WatchlistList, error) {
	var watchlists []entity.Watchlist
	var total int64

	db := global.GetDB().Model(&entity.Watchlist{}).Where("is_public = ?", true)
	if err := db.Count(&total).Error; err != nil {
		return response.WatchlistList{}, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&watchlists).Error; err != nil {
		return response.WatchlistList{}, err
	}

	items := make([]response.WatchlistDetail, 0, len(watchlists))
	for _, wl := range watchlists {
		items = append(items, *ws.toWatchlistDetail(&wl, nil))
	}

	return response.WatchlistList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// Update 更新关注列表
func (ws *WatchlistService) Update(watchlistId, userId uint, req *request.WatchlistUpdateRequest) error {
	watchlist, err := ws.getEntityById(watchlistId)
	if err != nil {
		return err
	}
	if watchlist.UserId != userId {
		return errors.New("not the watchlist owner")
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	return global.GetDB().Model(watchlist).Updates(updates).Error
}

// Delete 删除关注列表
func (ws *WatchlistService) Delete(watchlistId, userId uint) error {
	watchlist, err := ws.getEntityById(watchlistId)
	if err != nil {
		return err
	}
	if watchlist.UserId != userId {
		return errors.New("not the watchlist owner")
	}

	// 删除列表项
	global.GetDB().Where("watchlist_id = ?", watchlistId).Delete(&entity.WatchlistItem{})
	// 删除列表
	return global.GetDB().Delete(&entity.Watchlist{}, watchlistId).Error
}

// AddStock 添加股票到关注列表
func (ws *WatchlistService) AddStock(watchlistId, userId uint, req *request.WatchlistAddStockRequest) error {
	watchlist, err := ws.getEntityById(watchlistId)
	if err != nil {
		return err
	}
	if watchlist.UserId != userId {
		return errors.New("not the watchlist owner")
	}

	// 检查是否已存在
	var existing entity.WatchlistItem
	if err := global.GetDB().Where("watchlist_id = ? AND symbol = ?", watchlistId, req.Symbol).First(&existing).Error; err == nil {
		return errors.New("stock already in watchlist")
	}

	// 获取当前最大排序号
	var maxOrder int
	global.GetDB().Model(&entity.WatchlistItem{}).Where("watchlist_id = ?", watchlistId).
		Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	item := &entity.WatchlistItem{
		WatchlistId: watchlistId,
		Symbol:      req.Symbol,
		StockName:   req.StockName,
		AddedPrice:  req.AddedPrice,
		Note:        req.Note,
		SortOrder:   maxOrder + 1,
	}

	return global.GetDB().Create(item).Error
}

// RemoveStock 从关注列表移除股票
func (ws *WatchlistService) RemoveStock(watchlistId, userId uint, symbol string) error {
	watchlist, err := ws.getEntityById(watchlistId)
	if err != nil {
		return err
	}
	if watchlist.UserId != userId {
		return errors.New("not the watchlist owner")
	}

	return global.GetDB().Where("watchlist_id = ? AND symbol = ?", watchlistId, symbol).
		Delete(&entity.WatchlistItem{}).Error
}

// getEntityById 获取关注列表实体
func (ws *WatchlistService) getEntityById(id uint) (*entity.Watchlist, error) {
	var watchlist entity.Watchlist
	err := global.GetDB().Where("id = ?", id).First(&watchlist).Error
	return &watchlist, err
}

// toWatchlistDetail 转换为响应结构
func (ws *WatchlistService) toWatchlistDetail(watchlist *entity.Watchlist, items []entity.WatchlistItem) *response.WatchlistDetail {
	detail := &response.WatchlistDetail{
		ID:          watchlist.ID,
		CreatedAt:   watchlist.CreatedAt,
		UpdatedAt:   watchlist.UpdatedAt,
		UserId:      watchlist.UserId,
		Name:        watchlist.Name,
		Description: watchlist.Description,
		IsPublic:    watchlist.IsPublic,
	}

	// 获取列表项
	if items == nil {
		global.GetDB().Where("watchlist_id = ?", watchlist.ID).Order("sort_order asc").Find(&items)
	}

	detail.ItemCount = len(items)
	detail.Items = make([]response.WatchlistItemDetail, 0, len(items))
	for _, item := range items {
		detail.Items = append(detail.Items, response.WatchlistItemDetail{
			ID:         item.ID,
			Symbol:     item.Symbol,
			StockName:  item.StockName,
			AddedPrice: item.AddedPrice,
			Note:       item.Note,
			SortOrder:  item.SortOrder,
			CreatedAt:  item.CreatedAt,
		})
	}

	return detail
}
