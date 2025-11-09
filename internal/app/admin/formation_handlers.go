package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/gin-gonic/gin"
)

// FormationConfigResponse 陣型配置響應
type FormationConfigResponse struct {
	Enabled                  bool                               `json:"enabled"`
	MinInterval              string                             `json:"min_interval"`
	MaxInterval              string                             `json:"max_interval"`
	BaseSpawnChance          float64                            `json:"base_spawn_chance"`
	FormationWeights         map[string]float64                 `json:"formation_weights"`
	MinFishCount             int                                `json:"min_fish_count"`
	MaxFishCount             int                                `json:"max_fish_count"`
	FishCountByFormation     map[string]game.FishCountRange     `json:"fish_count_by_formation"`
	RoutePreferences         map[string]float64                 `json:"route_preferences"`
	AllowRandomRoute         bool                               `json:"allow_random_route"`
	FishSizePreferences      map[string]float64                 `json:"fish_size_preferences"`
	UniformTypeChance        float64                            `json:"uniform_type_chance"`
	MaxConcurrentFormations  int                                `json:"max_concurrent_formations"`
	DynamicDifficulty        bool                               `json:"dynamic_difficulty"`
	SpecialEventMultiplier   float64                            `json:"special_event_multiplier"`
}

// UpdateFormationConfigRequest 更新陣型配置請求
type UpdateFormationConfigRequest struct {
	Enabled                  *bool                              `json:"enabled,omitempty"`
	MinInterval              *int                               `json:"min_interval,omitempty"`              // seconds
	MaxInterval              *int                               `json:"max_interval,omitempty"`              // seconds
	BaseSpawnChance          *float64                           `json:"base_spawn_chance,omitempty"`
	FormationWeights         map[string]float64                 `json:"formation_weights,omitempty"`
	MinFishCount             *int                               `json:"min_fish_count,omitempty"`
	MaxFishCount             *int                               `json:"max_fish_count,omitempty"`
	FishCountByFormation     map[string]game.FishCountRange     `json:"fish_count_by_formation,omitempty"`
	RoutePreferences         map[string]float64                 `json:"route_preferences,omitempty"`
	AllowRandomRoute         *bool                              `json:"allow_random_route,omitempty"`
	FishSizePreferences      map[string]float64                 `json:"fish_size_preferences,omitempty"`
	UniformTypeChance        *float64                           `json:"uniform_type_chance,omitempty"`
	MaxConcurrentFormations  *int                               `json:"max_concurrent_formations,omitempty"`
	DynamicDifficulty        *bool                              `json:"dynamic_difficulty,omitempty"`
	SpecialEventMultiplier   *float64                           `json:"special_event_multiplier,omitempty"`
}

// SetDifficultyRequest 設置難度請求
type SetDifficultyRequest struct {
	Difficulty string `json:"difficulty" binding:"required"` // easy, normal, hard, boss_rush
}

// SetSpawnRateRequest 設置生成率請求
type SetSpawnRateRequest struct {
	MinInterval int     `json:"min_interval" binding:"required"` // seconds
	MaxInterval int     `json:"max_interval" binding:"required"` // seconds
	BaseChance  float64 `json:"base_chance" binding:"required"`
}

// TriggerEventRequest 觸發特殊事件請求
type TriggerEventRequest struct {
	Multiplier float64 `json:"multiplier" binding:"required"`
	Duration   int     `json:"duration" binding:"required"` // seconds
}

// GetFormationConfig 獲取當前陣型配置
func (s *AdminService) GetFormationConfig(c *gin.Context) {
	config := s.gameApp.GetGameUsecase().GetFormationConfig()

	// Convert to response format
	response := formatFormationConfigResponse(config)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateFormationConfig 更新陣型配置
func (s *AdminService) UpdateFormationConfig(c *gin.Context) {
	var req UpdateFormationConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Get current config
	currentConfig := s.gameApp.GetGameUsecase().GetFormationConfig()

	// Apply updates (only update non-nil fields)
	if req.Enabled != nil {
		currentConfig.Enabled = *req.Enabled
	}
	if req.MinInterval != nil {
		currentConfig.MinInterval = time.Duration(*req.MinInterval) * time.Second
	}
	if req.MaxInterval != nil {
		currentConfig.MaxInterval = time.Duration(*req.MaxInterval) * time.Second
	}
	if req.BaseSpawnChance != nil {
		currentConfig.BaseSpawnChance = *req.BaseSpawnChance
	}
	if req.FormationWeights != nil {
		// Convert string keys to FishFormationType
		weights := make(map[game.FishFormationType]float64)
		for k, v := range req.FormationWeights {
			weights[game.FishFormationType(k)] = v
		}
		currentConfig.FormationWeights = weights
	}
	if req.MinFishCount != nil {
		currentConfig.MinFishCount = *req.MinFishCount
	}
	if req.MaxFishCount != nil {
		currentConfig.MaxFishCount = *req.MaxFishCount
	}
	if req.FishCountByFormation != nil {
		ranges := make(map[game.FishFormationType]game.FishCountRange)
		for k, v := range req.FishCountByFormation {
			ranges[game.FishFormationType(k)] = v
		}
		currentConfig.FishCountByFormation = ranges
	}
	if req.RoutePreferences != nil {
		prefs := make(map[game.FishRouteType]float64)
		for k, v := range req.RoutePreferences {
			prefs[game.FishRouteType(k)] = v
		}
		currentConfig.RoutePreferences = prefs
	}
	if req.AllowRandomRoute != nil {
		currentConfig.AllowRandomRoute = *req.AllowRandomRoute
	}
	if req.FishSizePreferences != nil {
		currentConfig.FishSizePreferences = req.FishSizePreferences
	}
	if req.UniformTypeChance != nil {
		currentConfig.UniformTypeChance = *req.UniformTypeChance
	}
	if req.MaxConcurrentFormations != nil {
		currentConfig.MaxConcurrentFormations = *req.MaxConcurrentFormations
	}
	if req.DynamicDifficulty != nil {
		currentConfig.DynamicDifficulty = *req.DynamicDifficulty
	}
	if req.SpecialEventMultiplier != nil {
		currentConfig.SpecialEventMultiplier = *req.SpecialEventMultiplier
	}

	// Update config
	s.gameApp.GetGameUsecase().UpdateFormationConfig(currentConfig)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Formation config updated successfully",
		"data":    formatFormationConfigResponse(currentConfig),
	})
}

// SetFormationDifficulty 設置陣型難度（快捷方式）
func (s *AdminService) SetFormationDifficulty(c *gin.Context) {
	var req SetDifficultyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate difficulty
	validDifficulties := map[string]bool{
		"easy":      true,
		"normal":    true,
		"hard":      true,
		"boss_rush": true,
	}
	if !validDifficulties[req.Difficulty] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid difficulty level",
			"details": "Must be one of: easy, normal, hard, boss_rush",
		})
		return
	}

	if err := s.gameApp.GetGameUsecase().SetFormationDifficulty(req.Difficulty); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to set difficulty",
			"details": err.Error(),
		})
		return
	}

	newConfig := s.gameApp.GetGameUsecase().GetFormationConfig()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Formation difficulty set to " + req.Difficulty,
		"data":    formatFormationConfigResponse(newConfig),
	})
}

// SetFormationSpawnRate 設置陣型生成率
func (s *AdminService) SetFormationSpawnRate(c *gin.Context) {
	var req SetSpawnRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate
	if req.MinInterval <= 0 || req.MaxInterval <= 0 || req.MinInterval > req.MaxInterval {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid interval values",
			"details": "MinInterval and MaxInterval must be positive, and MinInterval <= MaxInterval",
		})
		return
	}
	if req.BaseChance < 0 || req.BaseChance > 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid base chance",
			"details": "BaseChance must be between 0 and 1",
		})
		return
	}

	s.gameApp.GetGameUsecase().SetFormationSpawnRate(req.MinInterval, req.MaxInterval, req.BaseChance)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Formation spawn rate updated successfully",
	})
}

// EnableFormationSpawn 啟用/禁用陣型生成
func (s *AdminService) EnableFormationSpawn(c *gin.Context) {
	enabledStr := c.Query("enabled")
	if enabledStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing 'enabled' query parameter",
		})
		return
	}

	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid 'enabled' value",
			"details": "Must be true or false",
		})
		return
	}

	s.gameApp.GetGameUsecase().EnableFormationSpawn(enabled)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Formation spawn enabled status updated",
		"enabled": enabled,
	})
}

// TriggerSpecialFormationEvent 觸發特殊陣型事件
func (s *AdminService) TriggerSpecialFormationEvent(c *gin.Context) {
	var req TriggerEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if req.Multiplier <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid multiplier",
			"details": "Multiplier must be positive",
		})
		return
	}
	if req.Duration <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid duration",
			"details": "Duration must be positive",
		})
		return
	}

	duration := time.Duration(req.Duration) * time.Second
	s.gameApp.GetGameUsecase().TriggerSpecialFormationEvent(req.Multiplier, duration)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Special formation event triggered",
		"multiplier": req.Multiplier,
		"duration": req.Duration,
	})
}

// GetFormationStats 獲取陣型生成統計
func (s *AdminService) GetFormationStats(c *gin.Context) {
	stats := s.gameApp.GetGameUsecase().GetFormationSpawnStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// formatFormationConfigResponse 格式化配置響應
func formatFormationConfigResponse(config game.FormationSpawnConfig) FormationConfigResponse {
	// Convert formation weights
	formationWeights := make(map[string]float64)
	for k, v := range config.FormationWeights {
		formationWeights[string(k)] = v
	}

	// Convert fish count by formation
	fishCountByFormation := make(map[string]game.FishCountRange)
	for k, v := range config.FishCountByFormation {
		fishCountByFormation[string(k)] = v
	}

	// Convert route preferences
	routePreferences := make(map[string]float64)
	for k, v := range config.RoutePreferences {
		routePreferences[string(k)] = v
	}

	return FormationConfigResponse{
		Enabled:                  config.Enabled,
		MinInterval:              config.MinInterval.String(),
		MaxInterval:              config.MaxInterval.String(),
		BaseSpawnChance:          config.BaseSpawnChance,
		FormationWeights:         formationWeights,
		MinFishCount:             config.MinFishCount,
		MaxFishCount:             config.MaxFishCount,
		FishCountByFormation:     fishCountByFormation,
		RoutePreferences:         routePreferences,
		AllowRandomRoute:         config.AllowRandomRoute,
		FishSizePreferences:      config.FishSizePreferences,
		UniformTypeChance:        config.UniformTypeChance,
		MaxConcurrentFormations:  config.MaxConcurrentFormations,
		DynamicDifficulty:        config.DynamicDifficulty,
		SpecialEventMultiplier:   config.SpecialEventMultiplier,
	}
}
