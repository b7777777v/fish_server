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
	dbManager *postgres.DBManager
	redis     *redis.Client
}

// DBManager 返回資料庫管理器，供 Repository 層使用
func (d *Data) DBManager() *postgres.DBManager {
	return d.dbManager
}

// NewData .創建一個新的 Data 結構
func NewData(c *conf.Data, logger logger.Logger) (*Data, func(), error) {
	// 初始化 PostgreSQL 資料庫管理器（支持讀寫分離）
	// 使用配置中的寫庫和讀庫配置
	writeDB := c.GetWriteDatabase()
	readDB := c.GetReadDatabase()

	dbManager, err := postgres.NewDBManagerWithConfig(writeDB, readDB, logger)
	if err != nil {
		logger.Errorf("failed to create postgres db manager: %v", err)
		return nil, nil, err
	}

	// 初始化 Redis 客戶端
	redisClient, err := redis.NewClientFromRedis(c.Redis, logger)
	if err != nil {
		logger.Errorf("failed to create redis client: %v", err)
		// 關閉已創建的資源
		dbManager.Close()
		return nil, nil, err
	}

	cleanup := func() {
		logger.Info("closing the data resources")
		dbManager.Close()
		redisClient.Close()
	}

	return &Data{dbManager: dbManager, redis: redisClient}, cleanup, nil
}

// NewAccountRepo creates a new AccountRepo
func NewAccountRepo(dbManager *postgres.DBManager) account.AccountRepo {
	return postgres.NewAccountRepo(dbManager)
}

// NewLobbyRepo creates a new LobbyRepo
func NewLobbyRepo(dbManager *postgres.DBManager) lobby.LobbyRepo {
	return postgres.NewLobbyRepo(dbManager)
}

// NewRoomCache creates a new RoomCache
func NewRoomCache(redisClient *redis.Client) lobby.RoomCache {
	return redis.NewRoomCache(redisClient.Redis)
}
