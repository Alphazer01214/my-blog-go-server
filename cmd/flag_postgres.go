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
	// 迁移数据库结构
	log.Println("migrate db")
	return db.AutoMigrate(&entity.User{}, &entity.Post{}, &entity.JWTBlacklist{})
}
