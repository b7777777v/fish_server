// internal/data/data.go
package data

import (
	"github.com/b7777777v/fish_server/internal/biz/account"
	"github.com/b7777777v/fish_server/internal/biz/lobby"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/data/postgres"
	"github.com/b7777777v/fish_server/internal/data/redis"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// Data .data 包含了所有數據源的客戶端，例如 db 和 redis
type Data struct {
	db    *postgres.Client
	redis *redis.Client
}

// NewData .創建一個新的 Data 結構
func NewData(c *conf.Data, logger logger.Logger) (*Data, func(), error) {
	// 初始化 PostgreSQL 客戶端
	pgClient, err := postgres.NewClientFromDatabase(c.Database, logger)
	if err != nil {
		logger.Errorf("failed to create postgres client: %v", err)
		return nil, nil, err
	}

	// 初始化 Redis 客戶端
	redisClient, err := redis.NewClientFromRedis(c.Redis, logger)
	if err != nil {
		logger.Errorf("failed to create redis client: %v", err)
		// 關閉已創建的資源
		pgClient.Close()
		return nil, nil, err
	}

	cleanup := func() {
		logger.Info("closing the data resources")
		pgClient.Close()
		redisClient.Close()
	}

	return &Data{db: pgClient, redis: redisClient}, cleanup, nil
}

// NewAccountRepo creates a new AccountRepo
func NewAccountRepo(pgClient *postgres.Client) account.AccountRepo {
	return postgres.NewAccountRepo(pgClient)
}

// NewLobbyRepo creates a new LobbyRepo
func NewLobbyRepo(pgClient *postgres.Client) lobby.LobbyRepo {
	return postgres.NewLobbyRepo(pgClient)
}

// NewRoomCache creates a new RoomCache
func NewRoomCache(redisClient *redis.Client) lobby.RoomCache {
	return redis.NewRoomCache(redisClient.Redis)
}
