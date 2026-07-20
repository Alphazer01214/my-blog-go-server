package global

import (
	"log"
	"sync"

	"blog.alphazer01214.top/internal/config"
	"blog.alphazer01214.top/internal/database"
	"blog.alphazer01214.top/internal/logs"
	"blog.alphazer01214.top/pkg/kafka"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 这是临时的解决方案

var (
	Config       *config.Config
	DB           *gorm.DB
	Log          *logs.Logman
	Redis        *redis.Client
	JWTBlacklist map[string]bool
	KafkaWriter  *kafka.Producer // Kafka 生产者
)

// User key: id value: token
var User = sync.Map{}

func Init() {
	Config = config.LoadConfig()
	DB = database.ConnectPostgres(Config.Postgres)
	Log = logs.NewLogman("server", "debug", 0)
	Redis = database.ConnectRedis(Config.Redis)
}

func GetConfig() *config.Config {
	if Config == nil {
		panic("nil config")
	}
	return Config
}

func GetDB() *gorm.DB {
	if DB == nil {
		panic("nil db")
	}
	return DB
}

func GetRedis() *redis.Client {
	if Redis == nil {
		log.Print("nil redis")
		panic("nil redis")
	}

	return Redis
}

func GetKafka() *kafka.Producer {
	if KafkaWriter == nil {
		log.Print("nil kafka producer")
		return nil
	}
	return KafkaWriter
}
