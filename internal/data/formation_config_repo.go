package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	redisClient "github.com/b7777777v/fish_server/internal/data/redis"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

const (
	// Redis key for formation config
	redisKeyFormationConfig = "game:formation:config"
	// Redis expiration time for formation config (24 hours)
	formationConfigExpiration = 24 * time.Hour
)

// FormationConfigRepo 陣型配置倉儲接口
type FormationConfigRepo interface {
	// GetConfig 獲取陣型配置
	GetConfig(ctx context.Context) (*game.FormationSpawnConfig, error)

	// SaveConfig 保存陣型配置
	SaveConfig(ctx context.Context, config *game.FormationSpawnConfig) error

	// GetPresetConfig 獲取預設配置
	GetPresetConfig(difficulty string) (*game.FormationSpawnConfig, error)
}

// formationConfigRepo 陣型配置倉儲實現
type formationConfigRepo struct {
	redis  *redisClient.Client
	logger logger.Logger
}

// NewFormationConfigRepo 創建陣型配置倉儲
func NewFormationConfigRepo(
	redis *redisClient.Client,
	logger logger.Logger,
) FormationConfigRepo {
	return &formationConfigRepo{
		redis:  redis,
		logger: logger.With("component", "formation_config_repo"),
	}
}

// GetConfig 獲取陣型配置
func (r *formationConfigRepo) GetConfig(ctx context.Context) (*game.FormationSpawnConfig, error) {
	// 從 Redis 獲取配置
	data, err := r.redis.Get(ctx, redisKeyFormationConfig)
	if err != nil {
		// 如果 Redis 中沒有配置，返回默認配置
		r.logger.Warnf("Failed to get config from Redis, using default: %v", err)
		defaultConfig := game.GetDefaultFormationSpawnConfig()

		// 嘗試保存默認配置到 Redis
		if saveErr := r.SaveConfig(ctx, &defaultConfig); saveErr != nil {
			r.logger.Errorf("Failed to save default config to Redis: %v", saveErr)
		}

		return &defaultConfig, nil
	}

	// 解析配置
	var config game.FormationSpawnConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		r.logger.Errorf("Failed to unmarshal config: %v", err)
		defaultConfig := game.GetDefaultFormationSpawnConfig()
		return &defaultConfig, nil
	}

	r.logger.Debugf("Loaded formation config from Redis")
	return &config, nil
}

// SaveConfig 保存陣型配置
func (r *formationConfigRepo) SaveConfig(ctx context.Context, config *game.FormationSpawnConfig) error {
	// 序列化配置
	data, err := json.Marshal(config)
	if err != nil {
		r.logger.Errorf("Failed to marshal config: %v", err)
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 保存到 Redis
	if err := r.redis.Set(ctx, redisKeyFormationConfig, data, formationConfigExpiration); err != nil {
		r.logger.Errorf("Failed to save config to Redis: %v", err)
		return fmt.Errorf("failed to save config to Redis: %w", err)
	}

	r.logger.Infof("Saved formation config to Redis")
	return nil
}

// GetPresetConfig 獲取預設配置
func (r *formationConfigRepo) GetPresetConfig(difficulty string) (*game.FormationSpawnConfig, error) {
	var config game.FormationSpawnConfig

	switch difficulty {
	case "easy":
		config = game.GetEasyFormationConfig()
	case "normal":
		config = game.GetNormalFormationConfig()
	case "hard":
		config = game.GetHardFormationConfig()
	case "boss_rush":
		config = game.GetBossRushConfig()
	default:
		return nil, fmt.Errorf("unknown difficulty: %s", difficulty)
	}

	return &config, nil
}
