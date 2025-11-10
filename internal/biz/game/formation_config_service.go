package game

import (
	"context"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// FormationConfigRepo 陣型配置倉儲接口（在 biz 層定義）
type FormationConfigRepo interface {
	GetConfig(ctx context.Context) (*FormationSpawnConfig, error)
	SaveConfig(ctx context.Context, config *FormationSpawnConfig) error
	LoadConfigFromDB(ctx context.Context) error
	GetPresetConfig(difficulty string) (*FormationSpawnConfig, error)
}

// FormationConfigService 陣型配置服務
type FormationConfigService struct {
	repo   FormationConfigRepo
	logger logger.Logger
}

// NewFormationConfigService 創建陣型配置服務
func NewFormationConfigService(
	repo FormationConfigRepo,
	logger logger.Logger,
) *FormationConfigService {
	return &FormationConfigService{
		repo:   repo,
		logger: logger.With("component", "formation_config_service"),
	}
}

// LoadConfig 載入配置（優先從 Redis，未命中則從 DB）
func (s *FormationConfigService) LoadConfig(ctx context.Context) (*FormationSpawnConfig, error) {
	config, err := s.repo.GetConfig(ctx)
	if err != nil {
		s.logger.Errorf("Failed to load config: %v", err)
		return nil, err
	}

	s.logger.Infof("Loaded formation config successfully")
	return config, nil
}

// SaveConfig 保存配置（同時寫入 DB 和 Redis）
func (s *FormationConfigService) SaveConfig(ctx context.Context, config *FormationSpawnConfig) error {
	if err := s.repo.SaveConfig(ctx, config); err != nil {
		s.logger.Errorf("Failed to save config: %v", err)
		return err
	}

	s.logger.Infof("Saved formation config to DB and Redis successfully")
	return nil
}

// LoadConfigFromDB 從資料庫載入配置到 Redis（服務啟動時調用）
func (s *FormationConfigService) LoadConfigFromDB(ctx context.Context) error {
	if err := s.repo.LoadConfigFromDB(ctx); err != nil {
		s.logger.Errorf("Failed to load config from DB to Redis: %v", err)
		return err
	}

	s.logger.Infof("Loaded formation config from DB to Redis successfully")
	return nil
}

// LoadPresetConfig 載入預設配置
func (s *FormationConfigService) LoadPresetConfig(ctx context.Context, difficulty string) (*FormationSpawnConfig, error) {
	config, err := s.repo.GetPresetConfig(difficulty)
	if err != nil {
		s.logger.Errorf("Failed to load preset config: %v", err)
		return nil, err
	}

	s.logger.Infof("Loaded preset config for difficulty: %s", difficulty)
	return config, nil
}

// ApplyConfigToSpawner 將配置應用到 Spawner
func (s *FormationConfigService) ApplyConfigToSpawner(spawner *FishSpawner, config *FormationSpawnConfig) {
	if config == nil {
		s.logger.Warn("Config is nil, using default")
		defaultConfig := GetDefaultFormationSpawnConfig()
		config = &defaultConfig
	}

	spawner.UpdateFormationConfig(*config)
	s.logger.Infof("Applied config to spawner")
}

// AutoSaveConfig 自動保存配置（後台任務）
func (s *FormationConfigService) AutoSaveConfig(ctx context.Context, spawner *FishSpawner, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Auto-save config stopped")
			return
		case <-ticker.C:
			config := spawner.GetFormationConfig()
			if err := s.SaveConfig(ctx, &config); err != nil {
				s.logger.Errorf("Auto-save config failed: %v", err)
			} else {
				s.logger.Debug("Auto-saved formation config")
			}
		}
	}
}
