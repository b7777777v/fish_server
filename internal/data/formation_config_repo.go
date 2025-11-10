package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	pgClient "github.com/b7777777v/fish_server/internal/data/postgres"
	redisClient "github.com/b7777777v/fish_server/internal/data/redis"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

const (
	// Redis key for formation config (cache)
	redisKeyFormationConfig = "game:formation:config:default"
	// Redis expiration time for formation config (1 hour)
	formationConfigExpiration = 1 * time.Hour
	// Database config key
	dbConfigKeyDefault = "default"
)

// FormationConfigRepo 陣型配置倉儲接口
type FormationConfigRepo interface {
	// GetConfig 獲取陣型配置（優先從 Redis 讀取，未命中則從 DB 載入）
	GetConfig(ctx context.Context) (*game.FormationSpawnConfig, error)

	// SaveConfig 保存陣型配置（同時寫入 DB 和 Redis）
	SaveConfig(ctx context.Context, config *game.FormationSpawnConfig) error

	// LoadConfigFromDB 從資料庫載入配置到 Redis（啟動時調用）
	LoadConfigFromDB(ctx context.Context) error

	// GetPresetConfig 獲取預設配置（不從 DB 讀取，直接返回預定義配置）
	GetPresetConfig(difficulty string) (*game.FormationSpawnConfig, error)
}

// formationConfigRepo 陣型配置倉儲實現
type formationConfigRepo struct {
	pg     *pgClient.Client
	redis  *redisClient.Client
	logger logger.Logger
}

// NewFormationConfigRepo 創建陣型配置倉儲
func NewFormationConfigRepo(
	pg *pgClient.Client,
	redis *redisClient.Client,
	logger logger.Logger,
) FormationConfigRepo {
	return &formationConfigRepo{
		pg:     pg,
		redis:  redis,
		logger: logger.With("component", "formation_config_repo"),
	}
}

// GetConfig 獲取陣型配置（優先從 Redis 讀取）
func (r *formationConfigRepo) GetConfig(ctx context.Context) (*game.FormationSpawnConfig, error) {
	// 1. 嘗試從 Redis 讀取（快取）
	data, err := r.redis.Get(ctx, redisKeyFormationConfig)
	if err == nil && data != "" {
		var config game.FormationSpawnConfig
		if err := json.Unmarshal([]byte(data), &config); err == nil {
			r.logger.Debugf("Loaded formation config from Redis cache")
			return &config, nil
		}
		r.logger.Warnf("Failed to unmarshal config from Redis: %v", err)
	}

	// 2. Redis 未命中，從資料庫讀取
	config, err := r.getConfigFromDB(ctx, dbConfigKeyDefault)
	if err != nil {
		r.logger.Errorf("Failed to load config from database: %v", err)
		// 返回默認配置
		defaultConfig := game.GetDefaultFormationSpawnConfig()
		return &defaultConfig, nil
	}

	// 3. 寫入 Redis 快取
	if err := r.saveConfigToRedis(ctx, config); err != nil {
		r.logger.Warnf("Failed to cache config to Redis: %v", err)
	}

	return config, nil
}

// SaveConfig 保存陣型配置（同時寫入 DB 和 Redis）
func (r *formationConfigRepo) SaveConfig(ctx context.Context, config *game.FormationSpawnConfig) error {
	// 1. 保存到資料庫（主存儲）
	if err := r.saveConfigToDB(ctx, dbConfigKeyDefault, config, "當前使用的陣型配置"); err != nil {
		r.logger.Errorf("Failed to save config to database: %v", err)
		return fmt.Errorf("failed to save config to database: %w", err)
	}

	// 2. 更新 Redis 快取
	if err := r.saveConfigToRedis(ctx, config); err != nil {
		r.logger.Warnf("Failed to update config in Redis: %v", err)
		// Redis 失敗不影響整體操作，僅記錄日誌
	}

	r.logger.Infof("Saved formation config to DB and Redis")
	return nil
}

// LoadConfigFromDB 從資料庫載入配置到 Redis（服務啟動時調用）
func (r *formationConfigRepo) LoadConfigFromDB(ctx context.Context) error {
	r.logger.Infof("Loading formation config from database to Redis...")

	// 從資料庫讀取
	config, err := r.getConfigFromDB(ctx, dbConfigKeyDefault)
	if err != nil {
		r.logger.Errorf("Failed to load config from database: %v", err)
		// 如果資料庫沒有配置，使用默認配置並保存
		defaultConfig := game.GetDefaultFormationSpawnConfig()
		if saveErr := r.saveConfigToDB(ctx, dbConfigKeyDefault, &defaultConfig, "默認陣型配置"); saveErr != nil {
			r.logger.Errorf("Failed to save default config to database: %v", saveErr)
			return saveErr
		}
		config = &defaultConfig
	}

	// 寫入 Redis
	if err := r.saveConfigToRedis(ctx, config); err != nil {
		r.logger.Errorf("Failed to cache config to Redis: %v", err)
		return err
	}

	r.logger.Infof("Successfully loaded formation config from DB to Redis")
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

// ========================================
// 私有輔助方法
// ========================================

// getConfigFromDB 從資料庫讀取配置
func (r *formationConfigRepo) getConfigFromDB(ctx context.Context, configKey string) (*game.FormationSpawnConfig, error) {
	query := `
		SELECT config_data
		FROM formation_configs
		WHERE config_key = $1 AND is_active = true
		LIMIT 1
	`

	var configJSON []byte
	err := r.pg.Pool.QueryRow(ctx, query, configKey).Scan(&configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to query config from database: %w", err)
	}

	var config game.FormationSpawnConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from database: %w", err)
	}

	return &config, nil
}

// saveConfigToDB 保存配置到資料庫
func (r *formationConfigRepo) saveConfigToDB(ctx context.Context, configKey string, config *game.FormationSpawnConfig, description string) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO formation_configs (config_key, config_data, description, is_active)
		VALUES ($1, $2, $3, true)
		ON CONFLICT (config_key)
		DO UPDATE SET
			config_data = EXCLUDED.config_data,
			description = EXCLUDED.description,
			updated_at = NOW()
	`

	_, err = r.pg.Pool.Exec(ctx, query, configKey, configJSON, description)
	if err != nil {
		return fmt.Errorf("failed to save config to database: %w", err)
	}

	return nil
}

// saveConfigToRedis 保存配置到 Redis（快取）
func (r *formationConfigRepo) saveConfigToRedis(ctx context.Context, config *game.FormationSpawnConfig) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := r.redis.Set(ctx, redisKeyFormationConfig, configJSON, formationConfigExpiration); err != nil {
		return fmt.Errorf("failed to save config to Redis: %w", err)
	}

	return nil
}
