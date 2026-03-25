package database

import (
	"blog.alphazer01214.top/internal/config"
	"blog.alphazer01214.top/internal/logs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var log = logs.NewLogman("db", "dev", 0)

func ConnectPostgres(dbCfg *config.Postgres) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dbCfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Error(err)
		panic(err)
	}
	return db
}
