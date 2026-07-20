package cmd

import (
	"errors"
	"log"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
)

func MigrateDB() error {
	db := global.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	// 处理 notifications 表结构变更（旧表 user_id 是主键，新表用 id 做主键）
	// 先删除旧表，由 AutoMigrate 重建
	if db.Migrator().HasTable(&entity.Notification{}) {
		log.Println("dropping old notifications table for schema migration")
		if err := db.Migrator().DropTable(&entity.Notification{}); err != nil {
			log.Printf("drop notifications table failed: %v (will try AutoMigrate anyway)", err)
		}
	}

	// 处理 trade_signals 和 signal_follows 表（新增）
	// 处理 watchlists 和 watchlist_items 表（新增）

	// 迁移数据库结构
	log.Println("migrate db")
	return db.AutoMigrate(
		&entity.User{}, &entity.UserProfile{}, &entity.UserSetting{}, &entity.UserFollow{},
		&entity.Post{}, &entity.Comment{}, &entity.Tag{},
		&entity.Action{},
		&entity.Agent{}, &entity.ChatSession{},
		&entity.Video{},
		&entity.UploadSession{}, &entity.UploadedChunk{},
		&entity.Notification{},
		&entity.TradeSignal{}, &entity.SignalFollow{},
		&entity.Watchlist{}, &entity.WatchlistItem{})
}
