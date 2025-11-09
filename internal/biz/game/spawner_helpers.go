package game

import (
	"math/rand"
	"time"
)

// ========================================
// FishSpawner 輔助函數
// ========================================

// generateFormationFishesWithSize 生成指定尺寸的陣型魚群
func (fs *FishSpawner) generateFormationFishesWithSize(count int, preferredSize string, config RoomConfig) []*Fish {
	fishes := make([]*Fish, 0, count)

	// 獲取指定尺寸的魚類型
	fishTypesOfSize := fs.GetFishTypesBySize(preferredSize)
	if len(fishTypesOfSize) == 0 {
		// 如果沒有該尺寸的魚，使用默認方法
		return fs.generateFormationFishes(count, config)
	}

	// 隨機選擇一個該尺寸的魚類型作為主要類型
	primaryType := fishTypesOfSize[fs.rng.Intn(len(fishTypesOfSize))]

	for i := 0; i < count; i++ {
		fish := fs.createFormationFish(&primaryType, config)
		fishes = append(fishes, fish)

		// 添加小延遲避免ID衝突
		// time.Sleep(1 * time.Millisecond) // 移除延遲，改用更好的ID生成
	}

	return fishes
}

// selectRouteByType 根據路線類型選擇路線
func (fs *FishSpawner) selectRouteByType(routeType FishRouteType) *FishRoute {
	routes := fs.formationManager.GetRoutesByType(routeType)
	if len(routes) == 0 {
		// 如果沒有該類型的路線，使用隨機路線
		return fs.formationManager.GetRandomRoute()
	}

	return routes[fs.rng.Intn(len(routes))]
}

// UpdateFormationConfig 更新陣型生成配置
func (fs *FishSpawner) UpdateFormationConfig(config FormationSpawnConfig) {
	fs.formationSpawnController.UpdateConfig(config)
	fs.logger.Infof("Updated formation spawn config")
}

// GetFormationConfig 獲取陣型生成配置
func (fs *FishSpawner) GetFormationConfig() FormationSpawnConfig {
	return fs.formationSpawnController.GetConfig()
}

// GetFormationSpawnStats 獲取陣型生成統計
func (fs *FishSpawner) GetFormationSpawnStats() map[string]interface{} {
	return fs.formationSpawnController.GetStats()
}

// SetFormationDifficulty 設置陣型難度（快捷方法）
func (fs *FishSpawner) SetFormationDifficulty(difficulty string) {
	var config FormationSpawnConfig

	switch difficulty {
	case "easy":
		config = GetEasyFormationConfig()
	case "normal":
		config = GetNormalFormationConfig()
	case "hard":
		config = GetHardFormationConfig()
	case "boss_rush":
		config = GetBossRushConfig()
	default:
		fs.logger.Warnf("Unknown difficulty: %s, using normal", difficulty)
		config = GetNormalFormationConfig()
	}

	fs.UpdateFormationConfig(config)
	fs.logger.Infof("Set formation difficulty to: %s", difficulty)
}

// EnableFormationSpawn 啟用/禁用陣型生成
func (fs *FishSpawner) EnableFormationSpawn(enabled bool) {
	config := fs.formationSpawnController.GetConfig()
	config.Enabled = enabled
	fs.formationSpawnController.UpdateConfig(config)
	fs.logger.Infof("Formation spawn enabled: %v", enabled)
}

// SetFormationSpawnRate 設置陣型生成率（快捷方法）
func (fs *FishSpawner) SetFormationSpawnRate(minInterval, maxInterval int, baseChance float64) {
	config := fs.formationSpawnController.GetConfig()
	config.MinInterval = time.Duration(minInterval) * time.Second
	config.MaxInterval = time.Duration(maxInterval) * time.Second
	config.BaseSpawnChance = baseChance
	fs.formationSpawnController.UpdateConfig(config)
	fs.logger.Infof("Set formation spawn rate: min=%ds, max=%ds, chance=%.2f",
		minInterval, maxInterval, baseChance)
}

// SetMaxConcurrentFormations 設置最大並發陣型數
func (fs *FishSpawner) SetMaxConcurrentFormations(max int) {
	config := fs.formationSpawnController.GetConfig()
	config.MaxConcurrentFormations = max
	fs.formationSpawnController.UpdateConfig(config)
	fs.logger.Infof("Set max concurrent formations: %d", max)
}

// SetFormationTypeWeights 設置陣型類型權重
func (fs *FishSpawner) SetFormationTypeWeights(weights map[FishFormationType]float64) {
	config := fs.formationSpawnController.GetConfig()
	config.FormationWeights = weights
	fs.formationSpawnController.UpdateConfig(config)
	fs.logger.Infof("Updated formation type weights")
}

// SetFishSizePreferences 設置魚尺寸偏好
func (fs *FishSpawner) SetFishSizePreferences(preferences map[string]float64) {
	config := fs.formationSpawnController.GetConfig()
	config.FishSizePreferences = preferences
	fs.formationSpawnController.UpdateConfig(config)
	fs.logger.Infof("Updated fish size preferences")
}

// TriggerSpecialEvent 觸發特殊事件（臨時提高生成率）
func (fs *FishSpawner) TriggerSpecialEvent(multiplier float64, duration time.Duration) {
	config := fs.formationSpawnController.GetConfig()
	oldMultiplier := config.SpecialEventMultiplier
	config.SpecialEventMultiplier = multiplier
	fs.formationSpawnController.UpdateConfig(config)

	fs.logger.Infof("Triggered special event: multiplier=%.2f, duration=%v", multiplier, duration)

	// 在持續時間後恢復原來的倍率
	go func() {
		time.Sleep(duration)
		config := fs.formationSpawnController.GetConfig()
		config.SpecialEventMultiplier = oldMultiplier
		fs.formationSpawnController.UpdateConfig(config)
		fs.logger.Infof("Special event ended, restored multiplier to %.2f", oldMultiplier)
	}()
}

// NotifyFormationComplete 通知陣型完成（用於更新統計）
func (fs *FishSpawner) NotifyFormationComplete(formationID string) {
	fs.formationSpawnController.RecordFormationComplete()
	fs.logger.Debugf("Formation completed: %s", formationID)
}

// ========================================
// 隨機數輔助函數
// ========================================

func randomFloat() float64 {
	return rand.Float64()
}

func randomInt(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}
