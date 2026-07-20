package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// RedisCache 泛型缓存工具，内置三重防护
type RedisCache[T any] struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
	sf         singleflight.Group
}

// NewRedisCache 创建一个新的泛型缓存实例
func NewRedisCache[T any](client *redis.Client, prefix string, defaultTTL time.Duration) *RedisCache[T] {
	return &RedisCache[T]{
		client:     client,
		prefix:     prefix,
		defaultTTL: defaultTTL,
	}
}

// Get 从缓存获取数据
func (c *RedisCache[T]) Get(ctx context.Context, key string) (*T, error) {
	val, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err != nil {
		return nil, err
	}
	// 检查空值缓存
	if string(val) == "NULL" {
		return nil, nil
	}
	var result T
	if err := json.Unmarshal(val, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Set 写入缓存
func (c *RedisCache[T]) Set(ctx context.Context, key string, value *T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	ttl := c.ttlWithJitter()
	return c.client.Set(ctx, c.prefix+key, data, ttl).Err()
}

// SetWithTTL 写入缓存（指定 TTL）
func (c *RedisCache[T]) SetWithTTL(ctx context.Context, key string, value *T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.prefix+key, data, ttl).Err()
}

// SetNull 缓存空值（防穿透）
func (c *RedisCache[T]) SetNull(ctx context.Context, key string) error {
	return c.client.Set(ctx, c.prefix+key, []byte("NULL"), 60*time.Second).Err()
}

// Delete 删除缓存
func (c *RedisCache[T]) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.prefix+key).Err()
}

// DeletePattern 批量删除（Lua 脚本）
func (c *RedisCache[T]) DeletePattern(ctx context.Context, pattern string) error {
	script := redis.NewScript(`
		local keys = redis.call('KEYS', ARGV[1])
		if #keys > 0 then
			for i=1,#keys,5000 do
				redis.call('DEL', unpack(keys, i, math.min(i+4999, #keys)))
			end
		end
		return #keys
	`)
	return script.Run(ctx, c.client, []string{}, c.prefix+pattern).Err()
}

// GetOrLoad 缓存未命中时自动加载（singleflight 防击穿）
func (c *RedisCache[T]) GetOrLoad(ctx context.Context, key string, loader func() (*T, error)) (*T, error) {
	// 1. 先查缓存
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil // 命中（包括空值）
	}
	if err != redis.Nil {
		// Redis 异常，降级到直接加载
		return loader()
	}

	// 2. 缓存未命中，使用 singleflight 防击穿
	v, err, _ := c.sf.Do(c.prefix+key, func() (interface{}, error) {
		// double-check：防止并发期间其他 goroutine 已写入缓存
		val, err := c.Get(ctx, key)
		if err == nil {
			return val, nil
		}

		// 3. 从 DB 加载
		data, err := loader()
		if err != nil {
			return nil, err
		}

		// 4. 防穿透：空值也缓存
		if data == nil {
			_ = c.SetNull(ctx, key)
			return nil, nil
		}

		// 5. 写入缓存（TTL 加随机抖动防雪崩）
		_ = c.Set(ctx, key, data)
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	return v.(*T), nil
}

// ttlWithJitter 计算带随机抖动的 TTL
func (c *RedisCache[T]) ttlWithJitter() time.Duration {
	jitter := time.Duration(rand.Intn(120)) * time.Second
	return c.defaultTTL + jitter
}

// CacheKey 生成缓存 key
func CacheKey(parts ...interface{}) string {
	return fmt.Sprintf("%v", parts)
}
