package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testClient *Client
	testLogger logger.Logger
)

// 設置測試環境
func setupTestRedis(t *testing.T) {
	// 使用環境變量或默認值設置測試 Redis 連接
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		// 嘗試多種常見的 Redis 配置
		testAddrs := []string{
			"localhost:6379", // 默認配置
			"127.0.0.1:6379", // 本地配置
			"redis:6379",     // Docker 配置
		}
		
		// 嘗試連接到可用的 Redis 實例
		testLogger = logger.New(os.Stdout, "info", "console")
		
		for _, testAddr := range testAddrs {
			redisConfig := &conf.Redis{
				Addr:     testAddr,
				Password: "",
				DB:       0,
			}
			client, err := NewClientFromRedis(redisConfig, testLogger)
			if err == nil {
				client.Close()
				addr = testAddr
				break
			}
		}
		
		if addr == "" {
			t.Skip("Skipping test: no accessible Redis server found. Please start Redis or set TEST_REDIS_ADDR environment variable.")
		}
	}

	// 創建測試客戶端
	redisConfig := &conf.Redis{
		Addr:     addr,
		Password: os.Getenv("TEST_REDIS_PASSWORD"),
		DB:       0,
	}
	
	var err error
	testClient, err = NewClientFromRedis(redisConfig, testLogger)
	require.NoError(t, err)
	require.NotNil(t, testClient)
}

// 清理測試環境
func teardownTestRedis() {
	if testClient != nil {
		testClient.Close()
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	teardownTestRedis()
	os.Exit(code)
}

func TestNewClientFromRedis(t *testing.T) {
	setupTestRedis(t)
	
	t.Run("成功創建客戶端", func(t *testing.T) {
		redisConfig := &conf.Redis{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}
		
		client, err := NewClientFromRedis(redisConfig, testLogger)
		if err != nil {
			t.Skip("Redis server not available")
		}
		
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Redis)
		assert.NotNil(t, client.Logger)
		
		client.Close()
	})
	
	t.Run("配置為 nil", func(t *testing.T) {
		client, err := NewClientFromRedis(nil, testLogger)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "redis config is nil")
	})
	
	t.Run("地址為空", func(t *testing.T) {
		redisConfig := &conf.Redis{
			Addr:     "",
			Password: "",
			DB:       0,
		}
		
		client, err := NewClientFromRedis(redisConfig, testLogger)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "redis address is empty")
	})
	
	t.Run("無效地址", func(t *testing.T) {
		redisConfig := &conf.Redis{
			Addr:     "invalid:99999",
			Password: "",
			DB:       0,
		}
		
		client, err := NewClientFromRedis(redisConfig, testLogger)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestNewClient(t *testing.T) {
	setupTestRedis(t)
	
	t.Run("使用 Config 結構成功創建", func(t *testing.T) {
		cfg := &Config{
			Redis: &conf.Redis{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			},
		}
		
		client, err := NewClient(cfg, testLogger)
		if err != nil {
			t.Skip("Redis server not available")
		}
		
		assert.NoError(t, err)
		assert.NotNil(t, client)
		
		client.Close()
	})
	
	t.Run("Config.Redis 為 nil", func(t *testing.T) {
		cfg := &Config{Redis: nil}
		
		client, err := NewClient(cfg, testLogger)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "redis config is nil")
	})
}

func TestRedisOperations(t *testing.T) {
	setupTestRedis(t)
	
	if testClient == nil {
		t.Skip("Test client not available")
	}
	
	ctx := context.Background()
	testKey := "test:key"
	testValue := "test_value"
	
	// 清理測試數據
	defer func() {
		testClient.Del(ctx, testKey)
	}()
	
	t.Run("Set 和 Get 操作", func(t *testing.T) {
		// 設置值
		err := testClient.Set(ctx, testKey, testValue, time.Minute)
		assert.NoError(t, err)
		
		// 獲取值
		value, err := testClient.Get(ctx, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testValue, value)
	})
	
	t.Run("Exists 操作", func(t *testing.T) {
		// 設置值
		err := testClient.Set(ctx, testKey, testValue, time.Minute)
		assert.NoError(t, err)
		
		// 檢查存在
		exists, err := testClient.Exists(ctx, testKey)
		assert.NoError(t, err)
		assert.True(t, exists)
		
		// 檢查不存在的 key
		exists, err = testClient.Exists(ctx, "nonexistent:key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	
	t.Run("Expire 操作", func(t *testing.T) {
		// 設置值
		err := testClient.Set(ctx, testKey, testValue, time.Minute)
		assert.NoError(t, err)
		
		// 設置過期時間
		err = testClient.Expire(ctx, testKey, time.Second)
		assert.NoError(t, err)
		
		// 等待過期
		time.Sleep(time.Second * 2)
		
		// 檢查是否已過期
		exists, err := testClient.Exists(ctx, testKey)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	
	t.Run("Del 操作", func(t *testing.T) {
		// 設置值
		err := testClient.Set(ctx, testKey, testValue, time.Minute)
		assert.NoError(t, err)
		
		// 刪除值
		err = testClient.Del(ctx, testKey)
		assert.NoError(t, err)
		
		// 檢查是否已刪除
		exists, err := testClient.Exists(ctx, testKey)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	
	t.Run("Get 不存在的 key", func(t *testing.T) {
		_, err := testClient.Get(ctx, "nonexistent:key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis: nil")
	})
}

func TestRedisConnectionPool(t *testing.T) {
	setupTestRedis(t)
	
	if testClient == nil {
		t.Skip("Test client not available")
	}
	
	t.Run("並發操作測試", func(t *testing.T) {
		ctx := context.Background()
		numGoroutines := 10
		numOperations := 100
		
		// 使用 channel 來等待所有 goroutine 完成
		done := make(chan bool, numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()
				
				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("test:concurrent:%d:%d", id, j)
					value := fmt.Sprintf("value_%d_%d", id, j)
					
					// Set
					err := testClient.Set(ctx, key, value, time.Minute)
					assert.NoError(t, err)
					
					// Get
					retrievedValue, err := testClient.Get(ctx, key)
					assert.NoError(t, err)
					assert.Equal(t, value, retrievedValue)
					
					// Del
					err = testClient.Del(ctx, key)
					assert.NoError(t, err)
				}
			}(i)
		}
		
		// 等待所有 goroutine 完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}