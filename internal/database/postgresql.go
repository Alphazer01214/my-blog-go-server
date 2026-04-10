package database

import (
	"blog.alphazer01214.top/internal/config"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/logs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var log = logs.NewLogman("db", "dev", 0)

func ConnectPostgres(dbCfg *config.Postgres) *gorm.DB {
	if dbCfg == nil {
		panic("postgres config is nil")
	}

	db, err := gorm.Open(postgres.Open(dbCfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Error(err)
		panic(err)
	}
	return db
}

func ClearTokenBlacklist(db *gorm.DB) error {
	return db.Where("1=1").Delete(&entity.TokenBlacklist{}).Error
}
