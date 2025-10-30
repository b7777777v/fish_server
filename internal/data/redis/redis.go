// internal/data/redis/redis.go
package redis

import (
	"context"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// Config 是 Redis 的配置
type Config struct {
	Addr     string
	Password string
	DB       int
}

// Client 是 Redis 客戶端
type Client struct {
	Redis  *redis.Client
	Logger logger.Logger
}

// NewClient 創建一個新的 Redis 客戶端
func NewClient(cfg *Config, logger logger.Logger) (*Client, error) {
	// 創建 Redis 客戶端
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
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