package game

import (
	"time"
)

// ========================================
// 魚群陣型生成配置系統
// ========================================

// FormationSpawnConfig 陣型生成配置
type FormationSpawnConfig struct {
	// 基礎配置
	Enabled                bool          `json:"enabled"`                  // 是否啟用陣型生成
	MinInterval            time.Duration `json:"min_interval"`             // 最小生成間隔
	MaxInterval            time.Duration `json:"max_interval"`             // 最大生成間隔
	BaseSpawnChance        float64       `json:"base_spawn_chance"`        // 基礎生成概率 (0.0-1.0)

	// 陣型類型概率配置
	FormationWeights       map[FishFormationType]float64 `json:"formation_weights"` // 各陣型權重

	// 規模配置
	MinFishCount           int           `json:"min_fish_count"`           // 最少魚數量
	MaxFishCount           int           `json:"max_fish_count"`           // 最多魚數量
	FishCountByFormation   map[FishFormationType]FishCountRange `json:"fish_count_by_formation"` // 各陣型的魚數量範圍

	// 路線配置
	RoutePreferences       map[FishRouteType]float64 `json:"route_preferences"` // 路線類型偏好
	AllowRandomRoute       bool          `json:"allow_random_route"`       // 是否允許隨機路線

	// 魚類型配置
	FishSizePreferences    map[string]float64 `json:"fish_size_preferences"` // 魚尺寸偏好 (small/medium/large/boss)
	UniformTypeChance      float64       `json:"uniform_type_chance"`      // 統一魚類型的概率

	// 高級配置
	MaxConcurrentFormations int          `json:"max_concurrent_formations"` // 最大並發陣型數
	DynamicDifficulty       bool         `json:"dynamic_difficulty"`       // 是否根據玩家數量動態調整難度
	SpecialEventMultiplier  float64      `json:"special_event_multiplier"` // 特殊事件生成倍率
}

// FishCountRange 魚數量範圍
type FishCountRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// FormationSpawnController 陣型生成控制器
type FormationSpawnController struct {
	config              FormationSpawnConfig
	lastSpawnTime       time.Time
	currentFormations   int
	totalSpawned        int
	successfulSpawns    int
	failedSpawns        int
}

// NewFormationSpawnController 創建陣型生成控制器
func NewFormationSpawnController(config FormationSpawnConfig) *FormationSpawnController {
	// 設置默認值
	if config.MinInterval == 0 {
		config.MinInterval = 20 * time.Second
	}
	if config.MaxInterval == 0 {
		config.MaxInterval = 60 * time.Second
	}
	if config.BaseSpawnChance == 0 {
		config.BaseSpawnChance = 0.3
	}

	// 設置默認陣型權重
	if len(config.FormationWeights) == 0 {
		config.FormationWeights = GetDefaultFormationWeights()
	}

	// 設置默認魚數量範圍
	if len(config.FishCountByFormation) == 0 {
		config.FishCountByFormation = GetDefaultFishCountRanges()
	}

	// 設置默認路線偏好
	if len(config.RoutePreferences) == 0 {
		config.RoutePreferences = GetDefaultRoutePreferences()
	}

	// 設置默認魚尺寸偏好
	if len(config.FishSizePreferences) == 0 {
		config.FishSizePreferences = GetDefaultFishSizePreferences()
	}

	return &FormationSpawnController{
		config:            config,
		lastSpawnTime:     time.Now(),
		currentFormations: 0,
		totalSpawned:      0,
		successfulSpawns:  0,
		failedSpawns:      0,
	}
}

// ShouldSpawnFormation 判斷是否應該生成陣型
func (fsc *FormationSpawnController) ShouldSpawnFormation(currentPlayerCount int) bool {
	// 檢查是否啟用
	if !fsc.config.Enabled {
		return false
	}

	// 檢查並發數量限制
	if fsc.currentFormations >= fsc.config.MaxConcurrentFormations {
		return false
	}

	// 檢查時間間隔
	now := time.Now()
	timeSinceLastSpawn := now.Sub(fsc.lastSpawnTime)

	if timeSinceLastSpawn < fsc.config.MinInterval {
		return false
	}

	// 計算生成概率
	spawnChance := fsc.calculateSpawnChance(timeSinceLastSpawn, currentPlayerCount)

	// 隨機判斷
	return randomFloat() < spawnChance
}

// calculateSpawnChance 計算生成概率
func (fsc *FormationSpawnController) calculateSpawnChance(timeSinceLastSpawn time.Duration, playerCount int) float64 {
	baseChance := fsc.config.BaseSpawnChance

	// 時間因素：時間越長，概率越高
	timeMultiplier := float64(timeSinceLastSpawn) / float64(fsc.config.MaxInterval)
	if timeMultiplier > 2.0 {
		timeMultiplier = 2.0
	}

	// 動態難度：玩家越多，概率越高
	playerMultiplier := 1.0
	if fsc.config.DynamicDifficulty && playerCount > 0 {
		playerMultiplier = 1.0 + float64(playerCount-1)*0.15 // 每多一個玩家增加15%
		if playerMultiplier > 2.0 {
			playerMultiplier = 2.0
		}
	}

	// 特殊事件倍率
	eventMultiplier := fsc.config.SpecialEventMultiplier
	if eventMultiplier == 0 {
		eventMultiplier = 1.0
	}

	finalChance := baseChance * timeMultiplier * playerMultiplier * eventMultiplier

	// 限制在0-1之間
	if finalChance > 1.0 {
		finalChance = 1.0
	}
	if finalChance < 0 {
		finalChance = 0
	}

	return finalChance
}

// SelectFormationType 選擇陣型類型（基於權重）
func (fsc *FormationSpawnController) SelectFormationType() FishFormationType {
	return selectWeightedFormationType(fsc.config.FormationWeights)
}

// SelectFishCount 選擇魚數量
func (fsc *FormationSpawnController) SelectFishCount(formationType FishFormationType) int {
	countRange, exists := fsc.config.FishCountByFormation[formationType]
	if !exists {
		// 使用默認範圍
		return fsc.config.MinFishCount + randomInt(fsc.config.MaxFishCount-fsc.config.MinFishCount+1)
	}

	if countRange.Min >= countRange.Max {
		return countRange.Min
	}

	return countRange.Min + randomInt(countRange.Max-countRange.Min+1)
}

// SelectRouteType 選擇路線類型
func (fsc *FormationSpawnController) SelectRouteType() FishRouteType {
	if fsc.config.AllowRandomRoute && randomFloat() < 0.1 {
		return RouteTypeRandom
	}

	return selectWeightedRouteType(fsc.config.RoutePreferences)
}

// SelectFishSize 選擇魚尺寸
func (fsc *FormationSpawnController) SelectFishSize() string {
	return selectWeightedFishSize(fsc.config.FishSizePreferences)
}

// ShouldUseUniformType 是否使用統一魚類型
func (fsc *FormationSpawnController) ShouldUseUniformType() bool {
	return randomFloat() < fsc.config.UniformTypeChance
}

// RecordSpawn 記錄生成
func (fsc *FormationSpawnController) RecordSpawn(success bool) {
	fsc.totalSpawned++
	if success {
		fsc.successfulSpawns++
		fsc.currentFormations++
		fsc.lastSpawnTime = time.Now()
	} else {
		fsc.failedSpawns++
	}
}

// RecordFormationComplete 記錄陣型完成
func (fsc *FormationSpawnController) RecordFormationComplete() {
	if fsc.currentFormations > 0 {
		fsc.currentFormations--
	}
}

// GetStats 獲取統計信息
func (fsc *FormationSpawnController) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_spawned":       fsc.totalSpawned,
		"successful_spawns":   fsc.successfulSpawns,
		"failed_spawns":       fsc.failedSpawns,
		"current_formations":  fsc.currentFormations,
		"last_spawn_time":     fsc.lastSpawnTime,
		"success_rate":        float64(fsc.successfulSpawns) / float64(max(1, fsc.totalSpawned)),
	}
}

// UpdateConfig 更新配置
func (fsc *FormationSpawnController) UpdateConfig(newConfig FormationSpawnConfig) {
	fsc.config = newConfig
}

// GetConfig 獲取配置
func (fsc *FormationSpawnController) GetConfig() FormationSpawnConfig {
	return fsc.config
}

// ========================================
// 默認配置和輔助函數
// ========================================

// GetDefaultFormationSpawnConfig 獲取默認陣型生成配置
func GetDefaultFormationSpawnConfig() FormationSpawnConfig {
	return FormationSpawnConfig{
		Enabled:                  true,
		MinInterval:              20 * time.Second,
		MaxInterval:              60 * time.Second,
		BaseSpawnChance:          0.3,
		FormationWeights:         GetDefaultFormationWeights(),
		MinFishCount:             5,
		MaxFishCount:             20,
		FishCountByFormation:     GetDefaultFishCountRanges(),
		RoutePreferences:         GetDefaultRoutePreferences(),
		AllowRandomRoute:         true,
		FishSizePreferences:      GetDefaultFishSizePreferences(),
		UniformTypeChance:        0.7,
		MaxConcurrentFormations:  3,
		DynamicDifficulty:        true,
		SpecialEventMultiplier:   1.0,
	}
}

// GetDefaultFormationWeights 獲取默認陣型權重
func GetDefaultFormationWeights() map[FishFormationType]float64 {
	return map[FishFormationType]float64{
		FormationTypeV:        0.25,  // 25% V字型
		FormationTypeLine:     0.20,  // 20% 直線型
		FormationTypeCircle:   0.15,  // 15% 圓形
		FormationTypeTriangle: 0.15,  // 15% 三角形
		FormationTypeDiamond:  0.10,  // 10% 菱形
		FormationTypeWave:     0.10,  // 10% 波浪型
		FormationTypeSpiral:   0.05,  // 5% 螺旋型（稀有）
	}
}

// GetDefaultFishCountRanges 獲取默認魚數量範圍
func GetDefaultFishCountRanges() map[FishFormationType]FishCountRange {
	return map[FishFormationType]FishCountRange{
		FormationTypeV:        {Min: 5, Max: 12},
		FormationTypeLine:     {Min: 4, Max: 10},
		FormationTypeCircle:   {Min: 6, Max: 15},
		FormationTypeTriangle: {Min: 6, Max: 14},
		FormationTypeDiamond:  {Min: 5, Max: 11},
		FormationTypeWave:     {Min: 8, Max: 19},
		FormationTypeSpiral:   {Min: 10, Max: 20},
	}
}

// GetDefaultRoutePreferences 獲取默認路線偏好
func GetDefaultRoutePreferences() map[FishRouteType]float64 {
	return map[FishRouteType]float64{
		RouteTypeStraight:  0.30,  // 30% 直線
		RouteTypeCurved:    0.35,  // 35% 曲線
		RouteTypeZigzag:    0.15,  // 15% Z字型
		RouteTypeCircular:  0.15,  // 15% 圓形
		RouteTypeRandom:    0.05,  // 5% 隨機
	}
}

// GetDefaultFishSizePreferences 獲取默認魚尺寸偏好
func GetDefaultFishSizePreferences() map[string]float64 {
	return map[string]float64{
		"small":  0.50,  // 50% 小型魚
		"medium": 0.35,  // 35% 中型魚
		"large":  0.12,  // 12% 大型魚
		"boss":   0.03,  // 3% Boss魚
	}
}

// ========================================
// 預設配置方案
// ========================================

// GetEasyFormationConfig 簡單難度配置
func GetEasyFormationConfig() FormationSpawnConfig {
	config := GetDefaultFormationSpawnConfig()
	config.MinInterval = 30 * time.Second
	config.MaxInterval = 90 * time.Second
	config.BaseSpawnChance = 0.2
	config.MaxConcurrentFormations = 2
	config.FishSizePreferences = map[string]float64{
		"small":  0.70,
		"medium": 0.25,
		"large":  0.04,
		"boss":   0.01,
	}
	return config
}

// GetNormalFormationConfig 普通難度配置
func GetNormalFormationConfig() FormationSpawnConfig {
	return GetDefaultFormationSpawnConfig()
}

// GetHardFormationConfig 困難難度配置
func GetHardFormationConfig() FormationSpawnConfig {
	config := GetDefaultFormationSpawnConfig()
	config.MinInterval = 15 * time.Second
	config.MaxInterval = 45 * time.Second
	config.BaseSpawnChance = 0.4
	config.MaxConcurrentFormations = 4
	config.FishSizePreferences = map[string]float64{
		"small":  0.30,
		"medium": 0.40,
		"large":  0.25,
		"boss":   0.05,
	}
	return config
}

// GetBossRushConfig Boss模式配置
func GetBossRushConfig() FormationSpawnConfig {
	config := GetDefaultFormationSpawnConfig()
	config.MinInterval = 10 * time.Second
	config.MaxInterval = 30 * time.Second
	config.BaseSpawnChance = 0.6
	config.MaxConcurrentFormations = 5
	config.FishSizePreferences = map[string]float64{
		"small":  0.10,
		"medium": 0.20,
		"large":  0.40,
		"boss":   0.30,
	}
	config.FormationWeights = map[FishFormationType]float64{
		FormationTypeV:        0.30,
		FormationTypeCircle:   0.25,
		FormationTypeTriangle: 0.20,
		FormationTypeDiamond:  0.15,
		FormationTypeSpiral:   0.10,
	}
	return config
}

// ========================================
// 輔助函數
// ========================================

func selectWeightedFormationType(weights map[FishFormationType]float64) FishFormationType {
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	randomValue := randomFloat() * totalWeight
	currentWeight := 0.0

	for formationType, weight := range weights {
		currentWeight += weight
		if randomValue <= currentWeight {
			return formationType
		}
	}

	// 默認返回V字型
	return FormationTypeV
}

func selectWeightedRouteType(weights map[FishRouteType]float64) FishRouteType {
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	randomValue := randomFloat() * totalWeight
	currentWeight := 0.0

	for routeType, weight := range weights {
		currentWeight += weight
		if randomValue <= currentWeight {
			return routeType
		}
	}

	// 默認返回直線
	return RouteTypeStraight
}

func selectWeightedFishSize(weights map[string]float64) string {
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	randomValue := randomFloat() * totalWeight
	currentWeight := 0.0

	for size, weight := range weights {
		currentWeight += weight
		if randomValue <= currentWeight {
			return size
		}
	}

	// 默認返回小型
	return "small"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
