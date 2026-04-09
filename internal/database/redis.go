package database

import (
	"net"
	"time"

	"blog.alphazer01214.top/internal/config"
	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *config.Redis) *redis.Client {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		WriteTimeout: time.Second * time.Duration(cfg.Timeout),
		ReadTimeout:  time.Second * time.Duration(cfg.Timeout),
		PoolSize:     cfg.PoolSize,
	})

	return client
}
