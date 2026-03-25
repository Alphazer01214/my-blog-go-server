package database

import (
	"fmt"
	"time"

	"blog.alphazer01214.top/internal/config"
	"github.com/go-redis/redis"
)

func ConnectRedis(cfg *config.Redis) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		WriteTimeout: time.Second * time.Duration(cfg.Timeout),
		ReadTimeout:  time.Second * time.Duration(cfg.Timeout),
		PoolSize:     cfg.PoolSize,
	})

	return client
}
