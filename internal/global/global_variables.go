package global

import (
	"log"
	"sync"

	"blog.alphazer01214.top/internal/config"
	"blog.alphazer01214.top/internal/database"
	"blog.alphazer01214.top/internal/logs"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

// 这是临时的解决方案

var (
	Config       *config.Config
	DB           *gorm.DB
	Log          *logs.Logman
	Redis        *redis.Client
	JWTBlacklist map[string]bool
)

// User key: id value: token
var User = sync.Map{}

func Init() {
	Config = config.LoadConfig()
	DB = database.ConnectPostgres(Config.Postgres)
	Log = logs.NewLogman("server", "debug", 0)
	Redis = database.ConnectRedis(Config.Redis)
	JWTBlacklist = make(map[string]bool)
}

func GetConfig() *config.Config {
	if Config == nil {
		log.Print("nil config")
		panic("nil config")
		return nil
	}
	return Config
}

func GetDB() *gorm.DB {
	if DB == nil {
		log.Print("nil db")
		panic("nil db")
		return nil
	}
	return DB
}

func GetRedis() *redis.Client {
	if Redis == nil {
		log.Print("nil redis")
		panic("nil redis")
		return nil
	}

	return Redis
}
