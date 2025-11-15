package data

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/data/postgres"
	"github.com/b7777777v/fish_server/internal/data/redis"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// roomConfigRepo 房间配置仓库（组合 DB 和 Redis）
type roomConfigRepo struct {
	pgRepo  *postgres.RoomConfigRepo
	cache   *redis.RoomConfigCache
	logger  logger.Logger
}

// NewRoomConfigRepo 创建新的 RoomConfigRepo
func NewRoomConfigRepo(
	dbManager *postgres.DBManager,
	redisClient *redis.Client,
	logger logger.Logger,
) game.RoomConfigRepo {
	return &roomConfigRepo{
		pgRepo: postgres.NewRoomConfigRepo(dbManager),
		cache:  redis.NewRoomConfigCache(redisClient.Redis),
		logger: logger.With("module", "data/room_config_repo"),
	}
}

// GetRoomConfig 获取房间配置（优先从缓存，缓存未命中则从DB加载并缓存）
func (r *roomConfigRepo) GetRoomConfig(ctx context.Context, roomType string) (*game.RoomConfig, error) {
	// 1. 尝试从缓存获取
	config, err := r.cache.GetRoomConfig(ctx, roomType)
	if err != nil {
		r.logger.Warnf("Failed to get room config from cache: %v", err)
	} else if config != nil {
		r.logger.Debugf("Cache hit for room config: %s", roomType)
		return config, nil
	}

	// 2. 缓存未命中，从DB加载
	r.logger.Debugf("Cache miss for room config: %s, fetching from DB", roomType)
	config, err = r.pgRepo.GetRoomConfig(ctx, roomType)
	if err != nil {
		r.logger.Errorf("Failed to get room config from DB: %v", err)
		return nil, err
	}

	// 3. 写入缓存
	if err := r.cache.SetRoomConfig(ctx, roomType, config); err != nil {
		r.logger.Warnf("Failed to cache room config: %v", err)
	}

	return config, nil
}

// GetAllRoomConfigs 获取所有房间配置
func (r *roomConfigRepo) GetAllRoomConfigs(ctx context.Context) (map[string]*game.RoomConfig, error) {
	// 直接从DB加载所有配置
	configs, err := r.pgRepo.GetAllRoomConfigs(ctx)
	if err != nil {
		r.logger.Errorf("Failed to get all room configs from DB: %v", err)
		return nil, err
	}

	// 批量写入缓存
	for roomType, config := range configs {
		if err := r.cache.SetRoomConfig(ctx, roomType, config); err != nil {
			r.logger.Warnf("Failed to cache room config for %s: %v", roomType, err)
		}
	}

	return configs, nil
}
