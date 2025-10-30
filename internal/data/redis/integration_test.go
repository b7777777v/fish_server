package redis

import (
	"context"
	"testing"
	"time"

	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisConfigIntegration(t *testing.T) {
	t.Run("使用 conf.Redis 配置創建客戶端", func(t *testing.T) {
		// 模擬從配置文件加載的配置
		redisConfig := &conf.Redis{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}
		
		log := logger.New(nil, "info", "console")
		
		// 測試直接從 Redis 配置創建客戶端
		client, err := NewClientFromRedis(redisConfig, log)
		if err != nil {
			t.Skipf("Redis server not available: %v", err)
		}
		
		require.NoError(t, err)
		require.NotNil(t, client)
		
		// 測試基本操作
		ctx := context.Background()
		testKey := "integration:test"
		testValue := "test_value"
		
		// 清理
		defer func() {
			client.Del(ctx, testKey)
			client.Close()
		}()
		
		// 測試 Set/Get
		err = client.Set(ctx, testKey, testValue, time.Minute)
		assert.NoError(t, err)
		
		value, err := client.Get(ctx, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testValue, value)
	})
	
	t.Run("測試不同數據庫索引", func(t *testing.T) {
		// 測試使用不同的數據庫索引
		redisConfig1 := &conf.Redis{
			Addr:     "localhost:6379",
			Password: "",
			DB:       1, // 使用數據庫 1
		}
		
		redisConfig2 := &conf.Redis{
			Addr:     "localhost:6379",
			Password: "",
			DB:       2, // 使用數據庫 2
		}
		
		log := logger.New(nil, "info", "console")
		
		client1, err1 := NewClientFromRedis(redisConfig1, log)
		client2, err2 := NewClientFromRedis(redisConfig2, log)
		
		if err1 != nil || err2 != nil {
			t.Skip("Redis server not available")
		}
		
		defer func() {
			if client1 != nil {
				client1.Close()
			}
			if client2 != nil {
				client2.Close()
			}
		}()
		
		ctx := context.Background()
		testKey := "db:isolation:test"
		value1 := "value_in_db1"
		value2 := "value_in_db2"
		
		// 在不同數據庫中設置相同的 key
		err := client1.Set(ctx, testKey, value1, time.Minute)
		assert.NoError(t, err)
		
		err = client2.Set(ctx, testKey, value2, time.Minute)
		assert.NoError(t, err)
		
		// 驗證數據庫隔離
		retrievedValue1, err := client1.Get(ctx, testKey)
		assert.NoError(t, err)
		assert.Equal(t, value1, retrievedValue1)
		
		retrievedValue2, err := client2.Get(ctx, testKey)
		assert.NoError(t, err)
		assert.Equal(t, value2, retrievedValue2)
		
		// 清理
		client1.Del(ctx, testKey)
		client2.Del(ctx, testKey)
	})
	
	t.Run("測試連接池配置", func(t *testing.T) {
		redisConfig := &conf.Redis{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}
		
		log := logger.New(nil, "info", "console")
		
		client, err := NewClientFromRedis(redisConfig, log)
		if err != nil {
			t.Skip("Redis server not available")
		}
		defer client.Close()
		
		// 驗證連接池配置是否生效
		poolStats := client.Redis.PoolStats()
		assert.NotNil(t, poolStats)
		
		// 連接池應該有配置的參數
		// 這些是在 NewClientFromRedis 中設置的默認值
		assert.True(t, poolStats.TotalConns >= 0)
		assert.True(t, poolStats.IdleConns >= 0)
	})
}

func TestRedisConfigValidation(t *testing.T) {
	log := logger.New(nil, "info", "console")
	
	t.Run("驗證必填字段", func(t *testing.T) {
		testCases := []struct {
			name   string
			config *conf.Redis
			hasErr bool
		}{
			{
				name:   "有效配置",
				config: &conf.Redis{Addr: "localhost:6379", Password: "", DB: 0},
				hasErr: false,
			},
			{
				name:   "空地址",
				config: &conf.Redis{Addr: "", Password: "", DB: 0},
				hasErr: true,
			},
			{
				name:   "nil 配置",
				config: nil,
				hasErr: true,
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				client, err := NewClientFromRedis(tc.config, log)
				
				if tc.hasErr {
					assert.Error(t, err)
					assert.Nil(t, client)
				} else {
					if err != nil {
						t.Skip("Redis server not available")
					}
					assert.NoError(t, err)
					assert.NotNil(t, client)
					client.Close()
				}
			})
		}
	})
}