package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TODO: 實現魚潮系統的 HTTP API handlers
// 此檔案提供魚潮配置和管理的 RESTful API 端點

// RegisterFishTideRoutes 註冊魚潮相關的路由
func RegisterFishTideRoutes(r *gin.Engine /* TODO: 添加 FishTideManager 參數 */) {
	// TODO: 實現路由註冊
	// 建議路由結構（管理員功能）：
	// GET    /api/v1/admin/fish-tides          - 獲取所有魚潮配置
	// POST   /api/v1/admin/fish-tides          - 建立新的魚潮配置
	// PUT    /api/v1/admin/fish-tides/:id      - 更新魚潮配置
	// DELETE /api/v1/admin/fish-tides/:id      - 刪除魚潮配置
	// POST   /api/v1/admin/fish-tides/:id/start - 手動觸發魚潮（針對指定房間）
	// POST   /api/v1/admin/fish-tides/:id/stop  - 手動停止魚潮（針對指定房間）
}

// CreateFishTideRequest 建立魚潮請求
type CreateFishTideRequest struct {
	Name            string  `json:"name" binding:"required"`
	FishTypeID      int32   `json:"fish_type_id" binding:"required"`
	FishCount       int     `json:"fish_count" binding:"required,min=1"`
	DurationSeconds int     `json:"duration_seconds" binding:"required,min=1"`
	IntervalMs      int     `json:"interval_ms" binding:"required,min=1"`
	SpeedMultiplier float64 `json:"speed_multiplier" binding:"required,min=0.1"`
	TriggerRule     string  `json:"trigger_rule" binding:"required"`
	IsActive        bool    `json:"is_active"`
}

// UpdateFishTideRequest 更新魚潮請求
type UpdateFishTideRequest struct {
	Name            string  `json:"name"`
	FishTypeID      int32   `json:"fish_type_id"`
	FishCount       int     `json:"fish_count" binding:"min=1"`
	DurationSeconds int     `json:"duration_seconds" binding:"min=1"`
	IntervalMs      int     `json:"interval_ms" binding:"min=1"`
	SpeedMultiplier float64 `json:"speed_multiplier" binding:"min=0.1"`
	TriggerRule     string  `json:"trigger_rule"`
	IsActive        bool    `json:"is_active"`
}

// TriggerFishTideRequest 觸發魚潮請求
type TriggerFishTideRequest struct {
	RoomID string `json:"room_id" binding:"required"`
}

// handleGetFishTides 獲取所有魚潮配置
func handleGetFishTides(c *gin.Context) {
	// TODO: 實現獲取所有魚潮配置
	// 1. 呼叫 FishTideRepo.GetActiveTides
	// 2. 返回魚潮列表
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleCreateFishTide 建立新的魚潮配置
func handleCreateFishTide(c *gin.Context) {
	// TODO: 實現建立魚潮配置
	// 1. 驗證管理員權限
	// 2. 綁定並驗證請求
	// 3. 呼叫 FishTideRepo.CreateTide
	// 4. 返回成功訊息
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleUpdateFishTide 更新魚潮配置
func handleUpdateFishTide(c *gin.Context) {
	// TODO: 實現更新魚潮配置
	// 1. 驗證管理員權限
	// 2. 解析 tide_id
	// 3. 綁定並驗證請求
	// 4. 呼叫 FishTideRepo.UpdateTide
	// 5. 觸發配置熱更新（通知所有 Game Server）
	// 6. 返回成功訊息
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleDeleteFishTide 刪除魚潮配置
func handleDeleteFishTide(c *gin.Context) {
	// TODO: 實現刪除魚潮配置
	// 1. 驗證管理員權限
	// 2. 解析 tide_id
	// 3. 呼叫 FishTideRepo.DeleteTide
	// 4. 返回成功訊息
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleStartFishTide 手動觸發魚潮
func handleStartFishTide(c *gin.Context) {
	// TODO: 實現手動觸發魚潮
	// 1. 驗證管理員權限
	// 2. 解析 tide_id
	// 3. 綁定並驗證請求（獲取 room_id）
	// 4. 呼叫 FishTideManager.StartTide
	// 5. 返回成功訊息
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleStopFishTide 手動停止魚潮
func handleStopFishTide(c *gin.Context) {
	// TODO: 實現手動停止魚潮
	// 1. 驗證管理員權限
	// 2. 綁定並驗證請求（獲取 room_id）
	// 3. 呼叫 FishTideManager.StopTide
	// 4. 返回成功訊息
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
