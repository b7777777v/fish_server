// internal/data/redis/redis.go
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// Config 是 Redis 的配置
type Config struct {
	Redis *conf.Redis
}

// Client 是 Redis 客戶端
type Client struct {
	Redis  *redis.Client
	Logger logger.Logger
}

// NewClient 創建一個新的 Redis 客戶端
func NewClient(cfg *Config, logger logger.Logger) (*Client, error) {
	// 驗證配置
	if cfg.Redis == nil {
		logger.Error("redis config is nil")
		return nil, fmt.Errorf("redis config is nil")
	}
	
	return NewClientFromRedis(cfg.Redis, logger)
}

// NewClientFromRedis 直接從 Redis 配置創建客戶端
func NewClientFromRedis(redisConfig *conf.Redis, logger logger.Logger) (*Client, error) {
	// 驗證配置
	if redisConfig == nil {
		logger.Error("redis config is nil")
		return nil, fmt.Errorf("redis config is nil")
	}
	
	if redisConfig.Addr == "" {
		logger.Error("redis address is empty")
		return nil, fmt.Errorf("redis address is empty")
	}

	// 創建 Redis 客戶端
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
		// 連接池配置
		PoolSize:     10,
		MinIdleConns: 5,
		IdleTimeout:  time.Minute * 5,
	})

	// 測試連接
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Errorf("failed to connect to redis: %v", err)
		return nil, err
	}

	return &Client{
		Redis:  client,
		Logger: logger.With("module", "data/redis"),
	}, nil
}

// Close 關閉 Redis 連接
func (c *Client) Close() error {
	return c.Redis.Close()
}

// Set 設置 key-value 對
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Redis.Set(ctx, key, value, expiration).Err()
}

// Get 獲取 key 的值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Redis.Get(ctx, key).Result()
}

// Del 刪除 key
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.Redis.Del(ctx, keys...).Err()
}

// Exists 檢查 key 是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (bool, error) {
	result, err := c.Redis.Exists(ctx, keys...).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Expire 設置 key 的過期時間
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.Redis.Expire(ctx, key, expiration).Err()
}

// Incr 增加 key 的數值
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.Redis.Incr(ctx, key).Result()
}

// Decr 減少 key 的數值
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.Redis.Decr(ctx, key).Result()
}

// IncrBy 增加 key 的數值（指定增量）
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.Redis.IncrBy(ctx, key, value).Result()
}

// DecrBy 減少 key 的數值（指定減量）
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.Redis.DecrBy(ctx, key, value).Result()
}

// GetInt64 獲取 key 的整數值
func (c *Client) GetInt64(ctx context.Context, key string) (int64, error) {
	val, err := c.Redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil // key 不存在時返回 0
	}
	return val, err
}