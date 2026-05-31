package database

import (
	"fmt"
	"time"

	"blog.alphazer01214.top/internal/config"
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

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Error(err)
		panic(err)
	}

	// 设置最大打开连接数
	if dbCfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100) // 默认值
	}

	// 设置最大空闲连接数
	if dbCfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10) // 默认值
	}

	// 设置连接最大生命周期
	if dbCfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(dbCfg.ConnMaxLifetime) * time.Second)
	} else {
		sqlDB.SetConnMaxLifetime(3600 * time.Second) // 默认1小时
	}

	// 设置空闲连接最大生命周期
	if dbCfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(dbCfg.ConnMaxIdleTime) * time.Second)
	} else {
		sqlDB.SetConnMaxIdleTime(300 * time.Second) // 默认5分钟
	}

	log.Info(fmt.Sprintf("PostgreSQL connection pool configured, maxOpenConns=%d", sqlDB.Stats().MaxOpenConnections))

	return db
}
